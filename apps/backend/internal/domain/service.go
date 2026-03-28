package domain

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/podland/backend/internal/cloudflare"
	"github.com/podland/backend/internal/repository"
)

// DomainService handles domain operations
type DomainService struct {
	dnsManager *cloudflare.DNSManager
	db         *sql.DB
	vmRepo     repository.VMRepository
}

// Domain represents a domain mapping
type Domain struct {
	ID        string `json:"id"`
	VMID      string `json:"vm_id"`
	UserID    string `json:"user_id"`
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
	Status    string `json:"status"` // pending, active, error
	CreatedAt string `json:"created_at"`
}

// NewDomainService creates a new domain service
func NewDomainService(dnsManager *cloudflare.DNSManager, db *sql.DB, vmRepo repository.VMRepository) *DomainService {
	return &DomainService{
		dnsManager: dnsManager,
		db:         db,
		vmRepo:     vmRepo,
	}
}

// GetDomainsByUserID returns all domains for a user
func (s *DomainService) GetDomainsByUserID(ctx context.Context, userID string) ([]*Domain, error) {
	query := `
		SELECT id, vm_id, domain, domain_status, created_at
		FROM vms
		WHERE user_id = $1 AND domain IS NOT NULL
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*Domain
	for rows.Next() {
		var d Domain
		var status sql.NullString
		if err := rows.Scan(&d.ID, &d.VMID, &d.Domain, &status, &d.CreatedAt); err != nil {
			return nil, err
		}
		// Parse subdomain from full domain
		d.Subdomain = parseSubdomain(d.Domain)
		d.UserID = userID
		if status.Valid {
			d.Status = status.String
		} else {
			d.Status = "pending"
		}
		domains = append(domains, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

// DeleteDomain deletes a domain and its DNS record
func (s *DomainService) DeleteDomain(ctx context.Context, domainID, userID string) error {
	// Get VM to check ownership and get domain
	vm, err := s.vmRepo.GetVMByIDAndUser(ctx, domainID, userID)
	if err != nil {
		return fmt.Errorf("domain not found")
	}

	if vm.Domain == nil || *vm.Domain == "" {
		return fmt.Errorf("domain not assigned")
	}

	// Delete DNS record
	dnsRecord, err := s.dnsManager.GetRecordByName(ctx, *vm.Domain)
	if err == nil && dnsRecord != nil {
		if err := s.dnsManager.DeleteRecord(ctx, dnsRecord.ID); err != nil {
			return fmt.Errorf("failed to delete DNS record: %w", err)
		}
	}

	// Update database - remove domain assignment
	if err := s.vmRepo.UpdateVM(ctx, domainID, repository.VMUpdateInput{
		Domain:       nil,
		DomainStatus: nil,
	}); err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	return nil
}

// Helper: parse subdomain from full domain
func parseSubdomain(domain string) string {
	// vm-name.podland.app → vm-name
	const suffix = ".podland.app"
	if len(domain) > len(suffix) && strings.HasSuffix(domain, suffix) {
		return domain[:len(domain)-len(suffix)]
	}
	return domain
}

// FormatTime formats a time.Time to RFC3339 string
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
