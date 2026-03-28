package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

// DNSManager handles Cloudflare DNS operations
type DNSManager struct {
	client *cloudflare.Client
	zoneID string
}

// DNSRecord represents a DNS record with simplified fields
type DNSRecord struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Proxied    bool   `json:"proxied"`
	TTL        int    `json:"ttl"`
	CreatedAt  string `json:"created_at,omitempty"`
	ModifiedAt string `json:"modified_at,omitempty"`
}

// NewDNSManager creates a new DNS manager with the given API token and zone ID
func NewDNSManager(apiToken, zoneID string) *DNSManager {
	return &DNSManager{
		client: cloudflare.NewClient(option.WithAPIToken(apiToken)),
		zoneID: zoneID,
	}
}

// CreateCNAME creates a CNAME record for the given subdomain
// subdomain should be the full FQDN (e.g., "vm-name.podland.app")
// target is the CNAME target (e.g., "tunnel.podland.app")
func (m *DNSManager) CreateCNAME(ctx context.Context, subdomain, target string) (*DNSRecord, error) {
	proxied := true
	record, err := m.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.F(m.zoneID),
		Body: dns.RecordNewParamsBody{
			Name:    cloudflare.F(subdomain),
			Type:    cloudflare.F[dns.RecordNewParamsBodyType]("cname"),
			Content: cloudflare.F(target),
			Proxied: cloudflare.F(proxied),
			TTL:     cloudflare.F[dns.TTL](1), // Auto TTL
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create CNAME: %w", err)
	}

	return m.toDNSRecord(record), nil
}

// DeleteRecord deletes a DNS record by ID
func (m *DNSManager) DeleteRecord(ctx context.Context, recordID string) error {
	_, err := m.client.DNS.Records.Delete(ctx, recordID, dns.RecordDeleteParams{
		ZoneID: cloudflare.F(m.zoneID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}
	return nil
}

// GetRecordByID gets a DNS record by ID
func (m *DNSManager) GetRecordByID(ctx context.Context, recordID string) (*DNSRecord, error) {
	record, err := m.client.DNS.Records.Get(ctx, recordID, dns.RecordGetParams{
		ZoneID: cloudflare.F(m.zoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}
	return m.toDNSRecord(record), nil
}

// GetRecordByName gets a DNS record by name (FQDN)
func (m *DNSManager) GetRecordByName(ctx context.Context, name string) (*DNSRecord, error) {
	iter := m.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(m.zoneID),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(name),
		}),
		Type: cloudflare.F[dns.RecordListParamsType]("CNAME"),
	})

	for iter.Next() {
		record := iter.Current()
		return m.toDNSRecord(&record), nil
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("DNS record not found: %s", name)
}

// ListRecords lists all DNS records for the zone with optional filtering
func (m *DNSManager) ListRecords(ctx context.Context, name, recordType string) ([]*DNSRecord, error) {
	params := dns.RecordListParams{
		ZoneID: cloudflare.F(m.zoneID),
	}

	if name != "" {
		params.Name = cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(name),
		})
	}
	if recordType != "" {
		params.Type = cloudflare.F[dns.RecordListParamsType](dns.RecordListParamsType(recordType))
	}

	iter := m.client.DNS.Records.ListAutoPaging(ctx, params)

	var records []*DNSRecord
	for iter.Next() {
		record := iter.Current()
		records = append(records, m.toDNSRecord(&record))
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// UpdateRecord updates a DNS record's content
func (m *DNSManager) UpdateRecord(ctx context.Context, recordID, content string) (*DNSRecord, error) {
	record, err := m.client.DNS.Records.Edit(ctx, recordID, dns.RecordEditParams{
		ZoneID: cloudflare.F(m.zoneID),
		Body: dns.RecordEditParamsBody{
			Content: cloudflare.F(content),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update DNS record: %w", err)
	}
	return m.toDNSRecord(record), nil
}

// toDNSRecord converts a Cloudflare API record to our internal DNSRecord type
func (m *DNSManager) toDNSRecord(record *dns.RecordResponse) *DNSRecord {
	if record == nil {
		return nil
	}

	return &DNSRecord{
		ID:         record.ID,
		Name:       record.Name,
		Type:       string(record.Type),
		Content:    record.Content,
		Proxied:    record.Proxied,
		TTL:        int(record.TTL),
		CreatedAt:  record.CreatedOn.Format(time.RFC3339),
		ModifiedAt: record.ModifiedOn.Format(time.RFC3339),
	}
}

// WaitForDNSActive polls until a DNS record is active (proxied and propagating)
// Uses 10s polling interval, max 5 minutes
func (m *DNSManager) WaitForDNSActive(ctx context.Context, name string) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("DNS propagation timeout after 5 minutes")
		case <-ticker.C:
			record, err := m.GetRecordByName(ctx, name)
			if err != nil {
				// Record not found yet, keep polling
				continue
			}

			// Record exists and is proxied - consider it active
			// Cloudflare propagates proxied records within their network immediately
			if record.Proxied {
				return nil
			}
		}
	}
}
