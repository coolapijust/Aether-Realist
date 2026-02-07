// geo-test is a CLI tool for testing Geo data parsing and lookup
//
// Usage:
//   go run cmd/geo-test/main.go -download
//   go run cmd/geo-test/main.go -ip 8.8.8.8
//   go run cmd/geo-test/main.go -domain www.google.com
//
// Data sources:
//   GeoIP: https://cdn.jsdelivr.net/gh/DustinWin/ruleset_geodata@mihomo-geodata/geoip.dat
//   GeoSite: https://cdn.jsdelivr.net/gh/DustinWin/ruleset_geodata@mihomo-geodata/geosite-lite.dat
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"aether-rea/internal/geo"
)

const (
	geoIPURL   = "https://cdn.jsdelivr.net/gh/DustinWin/ruleset_geodata@mihomo-geodata/geoip.dat"
	geoSiteURL = "https://cdn.jsdelivr.net/gh/DustinWin/ruleset_geodata@mihomo-geodata/geosite-lite.dat"
)

func main() {
	var (
		download = flag.Bool("download", false, "Download geo data files")
		dataDir  = flag.String("dir", "./geo-data", "Directory for geo data files")
		
		testIP     = flag.String("ip", "", "Test IP lookup (e.g., 8.8.8.8)")
		testDomain = flag.String("domain", "", "Test domain lookup (e.g., www.google.com)")
		category   = flag.String("cat", "", "Test specific category (e.g., cn, google)")
		
		listCategories = flag.Bool("list", false, "List all available categories")
		benchmark      = flag.Bool("bench", false, "Run benchmark")
	)
	flag.Parse()

	geoIPPath := filepath.Join(*dataDir, "geoip.dat")
	geoSitePath := filepath.Join(*dataDir, "geosite-lite.dat")

	if *download {
		if err := downloadFile(geoIPPath, geoIPURL); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to download GeoIP: %v\n", err)
			os.Exit(1)
		}
		if err := downloadFile(geoSitePath, geoSiteURL); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to download GeoSite: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Download complete")
		return
	}

	// Load databases
	var geoIPDB *geo.GeoIPDatabase
	var geoSiteDB *geo.GeoSiteDatabase

	if _, err := os.Stat(geoIPPath); err == nil {
		fmt.Println("Loading GeoIP database...")
		start := time.Now()
		
		f, err := os.Open(geoIPPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open GeoIP: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		
		geoIPDB, err = geo.LoadGeoIP(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse GeoIP: %v\n", err)
			// Continue anyway
		} else {
			fmt.Printf("Loaded GeoIP in %v\n", time.Since(start))
		}
	}

	if _, err := os.Stat(geoSitePath); err == nil {
		fmt.Println("Loading GeoSite database...")
		start := time.Now()
		
		f, err := os.Open(geoSitePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open GeoSite: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		
		geoSiteDB, err = geo.LoadGeoSite(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse GeoSite: %v\n", err)
		} else {
			fmt.Printf("Loaded GeoSite in %v\n", time.Since(start))
		}
	}

	// List categories
	if *listCategories && geoSiteDB != nil {
		fmt.Println("\nAvailable categories:")
		for _, cat := range geoSiteDB.Categories() {
			fmt.Printf("  - %s\n", cat)
		}
	}

	// Test IP lookup
	if *testIP != "" && geoIPDB != nil {
		ip := net.ParseIP(*testIP)
		if ip == nil {
			fmt.Fprintf(os.Stderr, "Invalid IP: %s\n", *testIP)
			os.Exit(1)
		}
		
		fmt.Printf("\nIP Lookup: %s\n", *testIP)
		
		start := time.Now()
		country, ok := geoIPDB.Country(ip)
		elapsed := time.Since(start)
		
		if ok {
			fmt.Printf("  Country: %s (lookup took %v)\n", country, elapsed)
		} else {
			fmt.Printf("  Country: not found (lookup took %v)\n", elapsed)
		}
		
		isPrivate := geoIPDB.IsPrivate(ip)
		fmt.Printf("  IsPrivate: %v\n", isPrivate)
		
		isCN := geoIPDB.IsCN(ip)
		fmt.Printf("  IsCN: %v\n", isCN)
	}

	// Test domain lookup
	if *testDomain != "" && geoSiteDB != nil {
		fmt.Printf("\nDomain Lookup: %s\n", *testDomain)
		
		categories := geoSiteDB.Categories()
		
		if *category != "" {
			// Test specific category
			start := time.Now()
			matched := geoSiteDB.Match(*testDomain, *category)
			elapsed := time.Since(start)
			
			fmt.Printf("  Category '%s': %v (lookup took %v)\n", *category, matched, elapsed)
		} else {
			// Test all categories
			matched := []string{}
			start := time.Now()
			
			for _, cat := range categories {
				if geoSiteDB.Match(*testDomain, cat) {
					matched = append(matched, cat)
				}
			}
			
			elapsed := time.Since(start)
			
			if len(matched) > 0 {
				fmt.Printf("  Matched categories: %v (lookup took %v)\n", matched, elapsed)
			} else {
				fmt.Printf("  No match found (lookup took %v)\n", elapsed)
			}
		}
	}

	// Benchmark
	if *benchmark && geoIPDB != nil {
		runBenchmark(geoIPDB)
	}
}

func downloadFile(path, url string) error {
	fmt.Printf("Downloading %s...\n", url)
	
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	// Create directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Write file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	
	fmt.Printf("  Saved to %s (%d bytes)\n", path, written)
	return nil
}

func runBenchmark(db *geo.GeoIPDatabase) {
	fmt.Println("\nBenchmarking IP lookups...")
	
	// Test IPs
	testIPs := []string{
		"8.8.8.8",        // Google
		"1.1.1.1",        // Cloudflare
		"114.114.114.114", // China
		"223.5.5.5",      // Alibaba
		"192.168.1.1",    // Private
		"10.0.0.1",       // Private
		"172.16.0.1",     // Private
	}
	
	iterations := 10000
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		ipStr := testIPs[i%len(testIPs)]
		ip := net.ParseIP(ipStr)
		db.Country(ip)
	}
	
	elapsed := time.Since(start)
	qps := float64(iterations) / elapsed.Seconds()
	
	fmt.Printf("  %d lookups in %v (%.0f lookups/sec)\n", iterations, elapsed, qps)
	fmt.Printf("  Average: %v per lookup\n", elapsed/time.Duration(iterations))
}
