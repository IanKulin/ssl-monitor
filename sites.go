package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type Site struct {
	Name    string    `json:"name"`
	URL     string    `json:"url"`
	Enabled bool      `json:"enabled"`
	Added   time.Time `json:"added"`
}

type SitesList struct {
	Sites []Site `json:"sites"`
}

type CertResult struct {
	URL        string    `json:"url"`
	Name       string    `json:"name"`
	ExpiryDate time.Time `json:"expiry_date"`
	DaysLeft   int       `json:"days_left"`
	LastCheck  time.Time `json:"last_check"`
	Error      string    `json:"error,omitempty"`
}

type ScanResults struct {
	LastScan time.Time    `json:"last_scan"`
	Results  []CertResult `json:"results"`
}

func loadSites() ([]Site, error) {
	data, err := os.ReadFile("data/sites.json")
	if err != nil {
		return nil, err
	}

	var sitesList SitesList
	err = json.Unmarshal(data, &sitesList)
	if err != nil {
		return nil, err
	}

	return sitesList.Sites, nil
}

func checkCertificate(site Site) CertResult {
	result := CertResult{
		URL:       site.URL,
		Name:      site.Name,
		LastCheck: time.Now(),
	}

	// Set up connection with timeout
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", site.URL+":443", &tls.Config{
		ServerName: site.URL,
	})

	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer conn.Close()

	// Get the certificate
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		result.Error = "No certificates found"
		return result
	}

	// Use the first certificate (leaf certificate)
	cert := certs[0]
	result.ExpiryDate = cert.NotAfter
	result.DaysLeft = int(time.Until(cert.NotAfter).Hours() / 24)

	return result
}

func scanAllSites(sites []Site) ScanResults {
	results := ScanResults{
		LastScan: time.Now(),
		Results:  make([]CertResult, 0),
	}

	for _, site := range sites {
		if !site.Enabled {
			continue
		}

		fmt.Printf("Checking %s (%s)...\n", site.Name, site.URL)
		result := checkCertificate(site)
		results.Results = append(results.Results, result)

		if result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		} else {
			fmt.Printf("  Expires: %s (%d days)\n",
				result.ExpiryDate.Format("2006-01-02"), result.DaysLeft)
		}
	}

	return results
}

func saveResults(results ScanResults) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("data/results.json", data, 0644)
}
