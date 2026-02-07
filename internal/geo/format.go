// Package geo implements parsers for Mihomo/Clash.Meta geodata format.
// Format spec: https://github.com/MetaCubeX/mihomo/tree/Meta/component/geodata
//
// GeoIP.dat and GeoSite.dat are based on V2Ray's protobuf format with extensions.
// Download sources:
// - GeoSite: https://cdn.jsdelivr.net/gh/DustinWin/ruleset_geodata@mihomo-geodata/geosite-lite.dat
// - GeoIP: https://cdn.jsdelivr.net/gh/DustinWin/ruleset_geodata@mihomo-geodata/geoip.dat
package geo

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// Format constants
const (
	// GeoIP.dat format
	geoIPTagCN     = "CN" // China
	geoIPTagPrivate = "PRIVATE"
	
	// GeoSite.dat format  
	geoSiteTagCN       = "cn"
	geoSiteTagGoogle   = "google"
	geoSiteTagYoutube  = "youtube"
	geoSiteTagNetflix  = "netflix"
	geoSiteTagTelegram = "telegram"
)

// Entry represents a single geo entry (IP range or domain rule)
type Entry struct {
	Tag   string // Country code or site category
	Type  string // "ip", "domain", "domain_suffix", "domain_keyword", "regexp"
	Value string // IP CIDR or domain pattern
}

// GeoIPDatabase represents parsed GeoIP.dat
type GeoIPDatabase struct {
	version   uint32
	countries map[string]*ipTrie // country code -> IP trie
	mu        sync.RWMutex
}

// GeoSiteDatabase represents parsed GeoSite.dat
type GeoSiteDatabase struct {
	version    uint32
	categories map[string]*domainMatcher // category -> domain matcher
	mu         sync.RWMutex
}

// ipTrie is a compressed prefix tree for IP lookups
type ipTrie struct {
	root *ipNode
}

type ipNode struct {
	children [2]*ipNode // 0, 1 bits
	isEnd    bool
	tag      string
}

// domainMatcher handles multiple domain matching strategies
type domainMatcher struct {
	fullDomains    map[string]struct{} // exact match
	suffixDomains  []string            // suffix match (ends with)
	keywordDomains []string            // substring match
	regexpPatterns []string            // regex patterns (rarely used)
}

// LoadGeoIP loads and parses GeoIP.dat from reader
func LoadGeoIP(r io.Reader) (*GeoIPDatabase, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read geip data: %w", err)
	}

	// Try gzip decompression first
	var rawData []byte
	if bytes.HasPrefix(data, []byte{0x1f, 0x8b}) {
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("gzip decompress: %w", err)
		}
		rawData, err = io.ReadAll(gr)
		if err != nil {
			return nil, fmt.Errorf("read gzip: %w", err)
		}
		gr.Close()
	} else {
	rawData = data
	}

	db := &GeoIPDatabase{
		countries: make(map[string]*ipTrie),
	}

	if err := db.parse(rawData); err != nil {
		return nil, err
	}

	return db, nil
}

// LoadGeoSite loads and parses GeoSite.dat from reader
func LoadGeoSite(r io.Reader) (*GeoSiteDatabase, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read geosite data: %w", err)
	}

	// Try gzip decompression
	var rawData []byte
	if bytes.HasPrefix(data, []byte{0x1f, 0x8b}) {
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("gzip decompress: %w", err)
		}
		rawData, err = io.ReadAll(gr)
		if err != nil {
			return nil, fmt.Errorf("read gzip: %w", err)
		}
		gr.Close()
	} else {
		rawData = data
	}

	db := &GeoSiteDatabase{
		categories: make(map[string]*domainMatcher),
	}

	if err := db.parse(rawData); err != nil {
		return nil, err
	}

	return db, nil
}

// Country lookups IP's country code (e.g., "CN", "US")
func (db *GeoIPDatabase) Country(ip net.IP) (string, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Convert to 4-byte representation for IPv4-mapped IPv6
	ip = ip.To4()
	if ip == nil {
		ip = ip.To16()
		if ip == nil {
			return "", false
		}
	}

	// Check each country's trie
	for code, trie := range db.countries {
		if trie.contains(ip) {
			return code, true
		}
	}

	return "", false
}

// IsCN checks if IP is China
func (db *GeoIPDatabase) IsCN(ip net.IP) bool {
	code, ok := db.Country(ip)
	return ok && code == geoIPTagCN
}

// IsPrivate checks if IP is private/LAN
func (db *GeoIPDatabase) IsPrivate(ip net.IP) bool {
	db.mu.RLock()
	trie, ok := db.countries[geoIPTagPrivate]
	db.mu.RUnlock()
	
	if !ok {
		// Fallback to standard private ranges
		return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()
	}
	
	return trie.contains(ip)
}

// Match checks if domain matches category (e.g., "google", "cn")
func (db *GeoSiteDatabase) Match(domain, category string) bool {
	db.mu.RLock()
	matcher, ok := db.categories[category]
	db.mu.RUnlock()
	
	if !ok {
		return false
	}

	return matcher.match(domain)
}

// Categories returns all available categories
func (db *GeoSiteDatabase) Categories() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	cats := make([]string, 0, len(db.categories))
	for k := range db.categories {
		cats = append(cats, k)
	}
	return cats
}

// ipTrie methods

func newIPTrie() *ipTrie {
	return &ipTrie{root: &ipNode{}}
}

func (t *ipTrie) insert(cidr string, tag string) error {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	ip := ipnet.IP.To4()
	if ip == nil {
		ip = ipnet.IP.To16()
	}
	
	mask, _ := ipnet.Mask.Size()
	
	node := t.root
	for i := 0; i < mask; i++ {
		bit := (ip[i/8] >> (7 - i%8)) & 1
		if node.children[bit] == nil {
			node.children[bit] = &ipNode{}
		}
		node = node.children[bit]
	}
	node.isEnd = true
	node.tag = tag
	return nil
}

func (t *ipTrie) contains(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		ip = ip.To16()
	}

	node := t.root
	for i := 0; i < len(ip)*8; i++ {
		if node.isEnd {
			return true
		}
		bit := (ip[i/8] >> (7 - i%8)) & 1
		if node.children[bit] == nil {
			return false
		}
		node = node.children[bit]
	}
	return node.isEnd
}

// domainMatcher methods

func (m *domainMatcher) match(domain string) bool {
	domain = strings.ToLower(domain)
	
	// 1. Full match
	if _, ok := m.fullDomains[domain]; ok {
		return true
	}
	
	// 2. Suffix match (longest suffix wins)
	for _, suffix := range m.suffixDomains {
		if strings.HasSuffix(domain, suffix) {
			return true
		}
	}
	
	// 3. Keyword match
	for _, kw := range m.keywordDomains {
		if strings.Contains(domain, kw) {
			return true
		}
	}
	
	return false
}

// parse methods (simplified - actual implementation needs protobuf parsing)
// These are stubs - the actual format uses V2Ray's proto format

func (db *GeoIPDatabase) parse(data []byte) error {
	// TODO: Implement actual proto parsing
	// Format: https://github.com/v2fly/geoip/blob/main/proto/geoip.proto
	return fmt.Errorf("proto parsing not yet implemented")
}

func (db *GeoSiteDatabase) parse(data []byte) error {
	// TODO: Implement actual proto parsing
	// Format: https://github.com/v2fly/domain-list-community/blob/master/proto/geosite.proto
	return fmt.Errorf("proto parsing not yet implemented")
}
