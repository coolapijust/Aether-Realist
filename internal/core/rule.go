// Package core provides rule-based traffic routing
// Rules are executed by Core, not GUI
package core

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// ActionType defines what to do with matched traffic
type ActionType string

const (
	ActionDirect ActionType = "direct" // Direct connection (bypass proxy)
	ActionProxy  ActionType = "proxy"  // Route through Aether
	ActionBlock  ActionType = "block"  // Drop connection silently
	ActionReject ActionType = "reject" // Reject with error
)

// MatchType defines how to match traffic
type MatchType string

const (
	MatchDomain       MatchType = "domain"        // Exact domain match
	MatchDomainSuffix MatchType = "domain_suffix" // Suffix match (*.example.com)
	MatchDomainKeyword MatchType = "domain_keyword" // Substring match
	MatchGeoSite      MatchType = "geosite"       // Category from GeoSite (e.g., "google")
	MatchIP           MatchType = "ip"            // Exact IP match
	MatchIPCIDR       MatchType = "ip_cidr"       // CIDR range
	MatchGeoIP        MatchType = "geoip"         // Country code (e.g., "CN")
	MatchPort         MatchType = "port"          // Port number or range (80,443 or 1000-2000)
	MatchProcess      MatchType = "process"       // Process name (platform-specific)
)

// Rule defines a single routing rule
type Rule struct {
	ID       string     `json:"id"`       // Unique identifier
	Name     string     `json:"name"`     // Human-readable name
	Priority int        `json:"priority"` // Higher = evaluated first
	Enabled  bool       `json:"enabled"`
	
	// Match conditions (AND logic)
	Matches []MatchCondition `json:"matches"`
	
	// Action to take when matched
	Action   ActionType `json:"action"`
	Target   string     `json:"target,omitempty"` // For future: specific outbound tag
}

// MatchCondition defines a single match criterion
type MatchCondition struct {
	Type  MatchType `json:"type"`
	Value string    `json:"value"`
	Not   bool      `json:"not,omitempty"` // Negate match
}

// RuleEngine executes rules against connection requests
type RuleEngine struct {
	rules      []*Rule
	geoIP      GeoIPMatcher
	geoSite    GeoSiteMatcher
	mu         sync.RWMutex
	
	// Metrics
	matchCount   map[string]int64 // rule ID -> count
	defaultAction ActionType
}

// GeoIPMatcher is the interface for GeoIP lookups
type GeoIPMatcher interface {
	Country(ip net.IP) (string, bool)
	IsCN(ip net.IP) bool
	IsPrivate(ip net.IP) bool
}

// GeoSiteMatcher is the interface for GeoSite lookups
type GeoSiteMatcher interface {
	Match(domain, category string) bool
	Categories() []string
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine(defaultAction ActionType) *RuleEngine {
	return &RuleEngine{
		rules:         make([]*Rule, 0),
		matchCount:    make(map[string]int64),
		defaultAction: defaultAction,
	}
}

// SetGeoDatabases sets the geo databases for lookups
func (re *RuleEngine) SetGeoDatabases(geoIP GeoIPMatcher, geoSite GeoSiteMatcher) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.geoIP = geoIP
	re.geoSite = geoSite
}

// UpdateRules replaces all rules (atomic)
func (re *RuleEngine) UpdateRules(rules []*Rule) error {
	// Validate rules
	for _, r := range rules {
		if err := re.validateRule(r); err != nil {
			return fmt.Errorf("invalid rule %s: %w", r.ID, err)
		}
	}
	
	re.mu.Lock()
	re.rules = rules
	re.mu.Unlock()
	
	return nil
}

// AddRule adds a single rule
func (re *RuleEngine) AddRule(rule *Rule) error {
	if err := re.validateRule(rule); err != nil {
		return err
	}
	
	re.mu.Lock()
	re.rules = append(re.rules, rule)
	re.mu.Unlock()
	
	return nil
}

// RemoveRule removes a rule by ID
func (re *RuleEngine) RemoveRule(ruleID string) bool {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	for i, r := range re.rules {
		if r.ID == ruleID {
			re.rules = append(re.rules[:i], re.rules[i+1:]...)
			return true
		}
	}
	return false
}

// GetRules returns current rules (copy)
func (re *RuleEngine) GetRules() []*Rule {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	result := make([]*Rule, len(re.rules))
	copy(result, re.rules)
	return result
}

// Match evaluates rules against a connection request
// Returns the matching action and the rule ID (if matched)
func (re *RuleEngine) Match(req *MatchRequest) (*MatchResult, error) {
	re.mu.RLock()
	rules := make([]*Rule, len(re.rules))
	copy(rules, re.rules)
	geoIP := re.geoIP
	geoSite := re.geoSite
	re.mu.RUnlock()
	
	// Sort by priority (descending)
	sortRulesByPriority(rules)
	
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		
		matched, err := re.evaluateRule(rule, req, geoIP, geoSite)
		if err != nil {
			return nil, err
		}
		
		if matched {
			re.recordMatch(rule.ID)
			return &MatchResult{
				Action: rule.Action,
				RuleID: rule.ID,
				RuleName: rule.Name,
			}, nil
		}
	}
	
	// No match, return default
	return &MatchResult{
		Action: re.defaultAction,
		RuleID: "",
		RuleName: "default",
	}, nil
}

// MatchRequest contains connection info for rule matching
type MatchRequest struct {
	Domain  string   // Target domain (if available)
	IP      net.IP   // Target IP (resolved or overridden)
	Port    int      // Target port
	Process string   // Process name (platform-specific, may be empty)
	UID     int      // User ID (platform-specific, may be 0)
}

// MatchResult is the outcome of rule matching
type MatchResult struct {
	Action   ActionType
	RuleID   string
	RuleName string
}

// evaluateRule checks if a rule matches the request
func (re *RuleEngine) evaluateRule(rule *Rule, req *MatchRequest, geoIP GeoIPMatcher, geoSite GeoSiteMatcher) (bool, error) {
	// All conditions must match (AND logic)
	for _, cond := range rule.Matches {
		matched, err := re.evaluateCondition(cond, req, geoIP, geoSite)
		if err != nil {
			return false, err
		}
		
		if cond.Not {
			matched = !matched
		}
		
		if !matched {
			return false, nil
		}
	}
	
	return true, nil
}

// evaluateCondition checks a single match condition
func (re *RuleEngine) evaluateCondition(cond MatchCondition, req *MatchRequest, geoIP GeoIPMatcher, geoSite GeoSiteMatcher) (bool, error) {
	switch cond.Type {
	case MatchDomain:
		return strings.EqualFold(req.Domain, cond.Value), nil
		
	case MatchDomainSuffix:
		domain := strings.ToLower(req.Domain)
		suffix := strings.ToLower(cond.Value)
		if !strings.HasPrefix(suffix, ".") {
			suffix = "." + suffix
		}
		return strings.HasSuffix(domain, suffix), nil
		
	case MatchDomainKeyword:
		return strings.Contains(strings.ToLower(req.Domain), strings.ToLower(cond.Value)), nil
		
	case MatchGeoSite:
		if geoSite == nil {
			return false, nil
		}
		return geoSite.Match(req.Domain, cond.Value), nil
		
	case MatchIP:
		ip := net.ParseIP(cond.Value)
		if ip == nil {
			return false, fmt.Errorf("invalid IP: %s", cond.Value)
		}
		return req.IP.Equal(ip), nil
		
	case MatchIPCIDR:
		_, ipnet, err := net.ParseCIDR(cond.Value)
		if err != nil {
			return false, fmt.Errorf("invalid CIDR: %s", cond.Value)
		}
		return ipnet.Contains(req.IP), nil
		
	case MatchGeoIP:
		if geoIP == nil {
			return false, nil
		}
		country, ok := geoIP.Country(req.IP)
		if !ok {
			return false, nil
		}
		return strings.EqualFold(country, cond.Value), nil
		
	case MatchPort:
		// Parse port or range (e.g., "80", "1000-2000")
		return matchPort(req.Port, cond.Value)
		
	case MatchProcess:
		return strings.EqualFold(req.Process, cond.Value), nil
		
	default:
		return false, fmt.Errorf("unknown match type: %s", cond.Type)
	}
}

// validateRule validates a rule
func (re *RuleEngine) validateRule(rule *Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if len(rule.Matches) == 0 {
		return fmt.Errorf("rule must have at least one match condition")
	}
	
	validActions := []ActionType{ActionDirect, ActionProxy, ActionBlock, ActionReject}
	found := false
	for _, a := range validActions {
		if rule.Action == a {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid action: %s", rule.Action)
	}
	
	return nil
}

// recordMatch records a rule match for metrics
func (re *RuleEngine) recordMatch(ruleID string) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.matchCount[ruleID]++
}

// GetMatchStats returns match statistics
func (re *RuleEngine) GetMatchStats() map[string]int64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	result := make(map[string]int64, len(re.matchCount))
	for k, v := range re.matchCount {
		result[k] = v
	}
	return result
}

// Helper functions

func sortRulesByPriority(rules []*Rule) {
	// Simple bubble sort (small rule sets)
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority < rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}

func matchPort(port int, spec string) (bool, error) {
	// Single port: "80"
	// Range: "1000-2000"
	// List: "80,443,8080"
	
	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// Range
			var start, end int
			_, err := fmt.Sscanf(part, "%d-%d", &start, &end)
			if err != nil {
				return false, fmt.Errorf("invalid port range: %s", part)
			}
			if port >= start && port <= end {
				return true, nil
			}
		} else {
			// Single
			var p int
			_, err := fmt.Sscanf(part, "%d", &p)
			if err != nil {
				return false, fmt.Errorf("invalid port: %s", part)
			}
			if port == p {
				return true, nil
			}
		}
	}
	return false, nil
}
