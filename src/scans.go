package main

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Site struct {
	Name    string    `json:"name"`
	URL     string    `json:"url"`
	Enabled bool      `json:"enabled"`
	Added   time.Time `json:"added"`
}

type SitesList struct {
	Sites        []Site    `json:"sites"`
	LastModified time.Time `json:"last_modified"`
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

func checkCertificate(site Site) CertResult {
	result := CertResult{
		URL:       site.URL,
		Name:      site.Name,
		LastCheck: time.Now(),
	}

	LogDebug("Connecting to %s:443", site.URL)

	// Set up connection with timeout
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", site.URL+":443", &tls.Config{
		ServerName: site.URL,
	})

	if err != nil {
		result.Error = err.Error()
		LogWarning("Certificate check failed for %s: %s", site.URL, err.Error())
		return result
	}
	defer conn.Close()

	// Get the certificate
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		result.Error = "No certificates found"
		LogWarning("No certificates found for %s", site.URL)
		return result
	}

	// Use the first certificate (leaf certificate)
	cert := certs[0]
	result.ExpiryDate = cert.NotAfter
	result.DaysLeft = int(time.Until(cert.NotAfter).Hours() / 24)

	LogDebug("Certificate for %s expires %s (%d days)", site.URL, result.ExpiryDate.Format("2006-01-02"), result.DaysLeft)

	return result
}

func scanAllSites(sites []Site) ScanResults {
	results := ScanResults{
		LastScan: time.Now(),
		Results:  make([]CertResult, 0),
	}

	enabledCount := 0
	for _, site := range sites {
		if site.Enabled {
			enabledCount++
		}
	}

	LogInfo("Scanning %d enabled sites", enabledCount)

	for _, site := range sites {
		if !site.Enabled {
			LogDebug("Skipping disabled site %s", site.Name)
			continue
		}

		LogDebug("Checking %s (%s)", site.Name, site.URL)
		result := checkCertificate(site)
		results.Results = append(results.Results, result)

		if result.Error != "" {
			LogWarning("Error checking %s: %s", site.Name, result.Error)
		} else {
			LogInfo("Certificate for %s expires %s (%d days)", site.Name, result.ExpiryDate.Format("2006-01-02"), result.DaysLeft)
		}
	}

	LogInfo("Scan completed for %d sites", len(results.Results))
	return results
}

func saveResults(results ScanResults) error {
	resultsFilePath := filepath.Join(dataDirPath, "results.json")
	LogDebug("Saving scan results to %s", resultsFilePath)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		LogError("Error marshaling scan results: %v", err)
		return err
	}

	err = os.WriteFile(resultsFilePath, data, 0644)
	if err != nil {
		LogError("Error writing scan results file %s: %v", resultsFilePath, err)
		return err
	}

	LogDebug("Scan results saved successfully")
	return nil
}
