// Proto parsing for V2Ray/Mihomo geodata format
// Based on: https://github.com/v2fly/v2ray-core/blob/master/common/protoext/geodata.proto
package geo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
)

// V2Ray GeoIP format:
// message GeoIP {
//   string country_code = 1;
//   repeated CIDR cidr = 2;
// }
// message GeoIPList {
//   repeated GeoIP entry = 1;
// }

// V2Ray GeoSite format:
// message Domain {
//   enum Type {
//     Plain = 0;
//     Regex = 1;
//     RootDomain = 2;  // suffix match
//     Full = 3;
//   }
//   Type type = 1;
//   string value = 2;
// }
// message GeoSite {
//   string country_code = 1;
//   repeated Domain domain = 2;
// }
// message GeoSiteList {
//   repeated GeoSite entry = 1;
// }

// CIDR represents IP CIDR range
type CIDR struct {
	IP   net.IP
	Mask int
}

// Domain represents a domain rule
type Domain struct {
	Type  DomainType
	Value string
}

// DomainType enum
type DomainType int

const (
	DomainTypePlain      DomainType = 0 // keyword match
	DomainTypeRegex      DomainType = 1 // regex match (rare)
	DomainTypeRootDomain DomainType = 2 // suffix match (e.g., .google.com)
	DomainTypeFull       DomainType = 3 // exact match
)

// GeoIPList is the top-level structure
type GeoIPList struct {
	Entries []*GeoIPEntry
}

// GeoIPEntry represents one country's IP ranges
type GeoIPEntry struct {
	CountryCode string
	CIDRs       []*CIDR
}

// GeoSiteList is the top-level structure
type GeoSiteList struct {
	Entries []*GeoSiteEntry
}

// GeoSiteEntry represents one category's domains
type GeoSiteEntry struct {
	CountryCode string
	Domains     []*Domain
}

// ProtoReader reads V2Ray proto format (simplified wire format)
// Note: This is a simplified parser for the specific format used by mihomo
// Full proto parsing would require generated code from .proto files

type protoReader struct {
	r *bytes.Reader
}

func newProtoReader(data []byte) *protoReader {
	return &protoReader{r: bytes.NewReader(data)}
}

func (pr *protoReader) readGeoIPList() (*GeoIPList, error) {
	list := &GeoIPList{}
	
	// Skip header/version if present
	// Parse entries
	for pr.r.Len() > 0 {
		entry, err := pr.readGeoIPEntry()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if entry != nil {
			list.Entries = append(list.Entries, entry)
		}
	}
	
	return list, nil
}

func (pr *protoReader) readGeoIPEntry() (*GeoIPEntry, error) {
	entry := &GeoIPEntry{}
	
	// Read field tags and values
	// Wire format: (field_num << 3) | wire_type
	// String (wire_type=2): length (varint) + data
	// Embedded message (wire_type=2): length (varint) + data
	// Fixed32 (wire_type=5): 4 bytes
	
	for {
		tag, err := pr.readVarint()
		if err != nil {
			if err == io.EOF {
				return entry, nil
			}
			return nil, err
		}
		
		fieldNum := int(tag >> 3)
		wireType := int(tag & 0x7)
		
		switch wireType {
		case 2: // Length-delimited (string or embedded message)
			length, err := pr.readVarint()
			if err != nil {
				return nil, err
			}
			data := make([]byte, length)
			if _, err := io.ReadFull(pr.r, data); err != nil {
				return nil, err
			}
			
			switch fieldNum {
			case 1: // country_code
				entry.CountryCode = string(data)
			case 2: // CIDR (embedded message)
				cidr, err := parseCIDR(data)
				if err != nil {
					return nil, err
				}
				entry.CIDRs = append(entry.CIDRs, cidr)
			}
			
		case 0: // Varint
			_, err := pr.readVarint()
			if err != nil {
				return nil, err
			}
			
		case 5: // Fixed32
			data := make([]byte, 4)
			if _, err := io.ReadFull(pr.r, data); err != nil {
				return nil, err
			}
			
		case 1: // Fixed64
			data := make([]byte, 8)
			if _, err := io.ReadFull(pr.r, data); err != nil {
				return nil, err
			}
			
		default:
			return nil, fmt.Errorf("unknown wire type: %d", wireType)
		}
		
		// Check if we've reached end of entry (new entry starts or EOF)
		if pr.r.Len() == 0 {
			return entry, nil
		}
		
		// Peek next tag - if it's field 1, this is a new entry
		nextPos, _ := pr.r.Seek(0, io.SeekCurrent)
		nextTag, err := pr.readVarint()
		if err != nil {
			return entry, nil
		}
		pr.r.Seek(nextPos, io.SeekStart)
		
		if (nextTag >> 3) == 1 && entry.CountryCode != "" {
			// New entry starts
			return entry, nil
		}
	}
}

func (pr *protoReader) readVarint() (uint64, error) {
	var result uint64
	var shift uint
	
	for {
		b, err := pr.r.ReadByte()
		if err != nil {
			return 0, err
		}
		
		result |= uint64(b&0x7f) << shift
		if (b & 0x80) == 0 {
			return result, nil
		}
		
		shift += 7
		if shift >= 64 {
			return 0, fmt.Errorf("varint too long")
		}
	}
}

func parseCIDR(data []byte) (*CIDR, error) {
	// CIDR message format:
	// ip: bytes (field 1)
	// prefix: uint32 (field 2 or 3, depending on version)
	
	pr := &protoReader{r: bytes.NewReader(data)}
	
	var ip net.IP
	var prefix uint32
	
	for pr.r.Len() > 0 {
		tag, err := pr.readVarint()
		if err != nil {
			break
		}
		
		fieldNum := int(tag >> 3)
		wireType := int(tag & 0x7)
		
		switch wireType {
		case 2: // Length-delimited
			length, _ := pr.readVarint()
			ipData := make([]byte, length)
			io.ReadFull(pr.r, ipData)
			
			if fieldNum == 1 {
				ip = net.IP(ipData)
			}
			
		case 0: // Varint
			val, _ := pr.readVarint()
			if fieldNum == 2 || fieldNum == 3 {
				prefix = uint32(val)
			}
			
		case 5: // Fixed32
			data := make([]byte, 4)
			io.ReadFull(pr.r, data)
			if fieldNum == 2 || fieldNum == 3 {
				prefix = binary.LittleEndian.Uint32(data)
			}
		}
	}
	
	if ip == nil {
		return nil, fmt.Errorf("no IP in CIDR")
	}
	
	return &CIDR{
		IP:   ip,
		Mask: int(prefix),
	}, nil
}

// ParseGeoIPData parses raw GeoIP.dat content
func ParseGeoIPData(data []byte) (*GeoIPList, error) {
	pr := newProtoReader(data)
	return pr.readGeoIPList()
}

// ParseGeoSiteData parses raw GeoSite.dat content  
func ParseGeoSiteData(data []byte) (*GeoSiteList, error) {
	// Similar implementation for GeoSite
	// TODO: Implement full parser
	return nil, fmt.Errorf("geosite parsing not yet implemented")
}

// Helper: Convert GeoIPList to our internal format
func (list *GeoIPList) ToDatabase() *GeoIPDatabase {
	db := &GeoIPDatabase{
		countries: make(map[string]*ipTrie),
	}
	
	for _, entry := range list.Entries {
		if entry.CountryCode == "" {
			continue
		}
		
		trie := newIPTrie()
		for _, cidr := range entry.CIDRs {
			ipNet := &net.IPNet{
				IP:   cidr.IP,
				Mask: net.CIDRMask(cidr.Mask, len(cidr.IP)*8),
			}
			trie.insert(ipNet.String(), entry.CountryCode)
		}
		
		db.countries[entry.CountryCode] = trie
	}
	
	return db
}

// Helper: Convert GeoSiteList to our internal format
func (list *GeoSiteList) ToDatabase() *GeoSiteDatabase {
	db := &GeoSiteDatabase{
		categories: make(map[string]*domainMatcher),
	}
	
	for _, entry := range list.Entries {
		if entry.CountryCode == "" {
			continue
		}
		
		matcher := &domainMatcher{
			fullDomains: make(map[string]struct{}),
		}
		
		for _, domain := range entry.Domains {
			switch domain.Type {
			case DomainTypeFull:
				matcher.fullDomains[strings.ToLower(domain.Value)] = struct{}{}
			case DomainTypeRootDomain:
				matcher.suffixDomains = append(matcher.suffixDomains, "."+strings.ToLower(domain.Value))
			case DomainTypePlain:
				matcher.keywordDomains = append(matcher.keywordDomains, strings.ToLower(domain.Value))
			}
		}
		
		db.categories[entry.CountryCode] = matcher
	}
	
	return db
}
