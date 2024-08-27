package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type Credentials struct {
	APIToken string `json:"api_token"`
	ZoneID   string `json:"zone_id"`
}

func loadCredentials(filename string) (Credentials, error) {
	var creds Credentials
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return creds, err
	}
	err = json.Unmarshal(file, &creds)
	return creds, err
}

func addDNSRecord(api *cloudflare.API, zoneID, recordType, name, content string, ttl int) error {
	rc := cloudflare.ZoneIdentifier(zoneID)
	_, err := api.CreateDNSRecord(context.Background(), rc, cloudflare.CreateDNSRecordParams{
		Type:    recordType,
		Name:    name,
		Content: content,
		TTL:     ttl,
	})
	return err
}

func addSegmentedTXTRecord(api *cloudflare.API, zoneID, name string, contents []string, ttl int) error {
	for _, content := range contents {
		parts := strings.SplitN(content, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid content format: %s", content)
		}
		segmentNum, segmentContent := parts[0], parts[1]
		err := addDNSRecord(api, zoneID, "TXT", name, fmt.Sprintf("%s:%s", segmentNum, segmentContent), ttl)
		if err != nil {
			return err
		}
		time.Sleep(time.Second) // Avoid rate limiting
	}
	return nil
}

func getDNSRecords(api *cloudflare.API, zoneID string) ([]cloudflare.DNSRecord, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := api.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{})
	return records, err
}

func getDNSRecordsByType(api *cloudflare.API, zoneID, recordType string) ([]cloudflare.DNSRecord, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := api.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{
		Type: recordType,
	})
	return records, err
}

func processDNSRecords(records []cloudflare.DNSRecord) {
	fmt.Println("DNS Records:")
	fmt.Println("------------")
	for _, record := range records {
		proxiedStatus := "false"
		if record.Proxied != nil && *record.Proxied {
			proxiedStatus = "true"
		}
		fmt.Printf("%s | %s | %s | TTL: %d | Proxied: %s\n", record.Type, record.Name, record.Content, record.TTL, proxiedStatus)
	}

	fmt.Println("\nSummary:")
	fmt.Println("--------")
	typeCounts := make(map[string]int)
	for _, record := range records {
		typeCounts[record.Type]++
	}
	for recordType, count := range typeCounts {
		fmt.Printf("%s: %d\n", recordType, count)
	}

	fmt.Printf("Total records: %d\n", len(records))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go {add|get_all|get <type>}")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go add A example.com 192.0.2.1 3600")
		fmt.Println("  go run main.go add TXT example.com '1:content part 1' '2:content part 2' '3:content part 3' 3600")
		fmt.Println("  go run main.go get_all")
		fmt.Println("  go run main.go get A")
		os.Exit(1)
	}

	creds, err := loadCredentials("cloudflare_credentials.json")
	if err != nil {
		log.Fatalf("Error loading credentials: %v", err)
	}

	api, err := cloudflare.NewWithAPIToken(creds.APIToken)
	if err != nil {
		log.Fatalf("Error creating Cloudflare API client: %v", err)
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 6 {
			log.Fatal("Not enough arguments for add command")
		}
		recordType := os.Args[2]
		name := os.Args[3]
		ttl := 0
		fmt.Sscan(os.Args[len(os.Args)-1], &ttl)

		if recordType == "TXT" && strings.Contains(os.Args[4], ":") {
			contents := os.Args[4 : len(os.Args)-1]
			err = addSegmentedTXTRecord(api, creds.ZoneID, name, contents, ttl)
		} else {
			content := os.Args[4]
			err = addDNSRecord(api, creds.ZoneID, recordType, name, content, ttl)
		}
		if err != nil {
			log.Fatalf("Error adding DNS record: %v", err)
		}
		fmt.Println("DNS record added successfully")

	case "get_all":
		records, err := getDNSRecords(api, creds.ZoneID)
		if err != nil {
			log.Fatalf("Error getting DNS records: %v", err)
		}
		processDNSRecords(records)

	case "get":
		if len(os.Args) < 3 {
			log.Fatal("Record type is required for get command")
		}
		recordType := os.Args[2]
		records, err := getDNSRecordsByType(api, creds.ZoneID, recordType)
		if err != nil {
			log.Fatalf("Error getting DNS records: %v", err)
		}
		processDNSRecords(records)

	default:
		log.Fatalf("Unknown command: %s", os.Args[1])
	}
}