package domain

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/podland/backend/internal/cloudflare"
	"github.com/podland/backend/internal/repository"
)

// DNSPoller polls DNS propagation status and updates VM domain status
type DNSPoller struct {
	dnsManager *cloudflare.DNSManager
	db         repository.VMRepository
}

// NewDNSPoller creates a new DNS poller
func NewDNSPoller(dnsManager *cloudflare.DNSManager, db repository.VMRepository) *DNSPoller {
	return &DNSPoller{
		dnsManager: dnsManager,
		db:         db,
	}
}

// WaitForDNS polls DNS status and updates VM domain status
// Runs as a background goroutine with 10s poll interval and 5min timeout
func (p *DNSPoller) WaitForDNS(ctx context.Context, vmID, subdomain string) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			// Update VM status to error on timeout
			if err := p.db.UpdateVM(ctx, vmID, repository.VMUpdateInput{
				DomainStatus: stringPtr("error"),
			}); err != nil {
				log.Printf("Failed to update VM domain status to error: %v", err)
			}
			return fmt.Errorf("DNS propagation timeout after 5 minutes")
		case <-ticker.C:
			record, err := p.dnsManager.GetRecordByName(ctx, subdomain)
			if err != nil {
				// Record not found yet, keep polling
				continue
			}

			// Check if DNS is active (Cloudflare proxy status)
			if record.Proxied {
				// Update VM domain status to active
				if err := p.db.UpdateVM(ctx, vmID, repository.VMUpdateInput{
					DomainStatus: stringPtr("active"),
				}); err != nil {
					log.Printf("Failed to update VM domain status to active: %v", err)
				}
				return nil
			}
		}
	}
}

// StartDNSPoller starts a background goroutine to wait for DNS propagation
func (p *DNSPoller) StartDNSPoller(vmID, subdomain string) {
	go func() {
		ctx := context.Background()
		err := p.WaitForDNS(ctx, vmID, subdomain)
		if err != nil {
			log.Printf("DNS propagation failed for VM %s (%s): %v", vmID, subdomain, err)
		} else {
			log.Printf("DNS propagation successful for VM %s (%s)", vmID, subdomain)
		}
	}()
}

func stringPtr(s string) *string {
	return &s
}
