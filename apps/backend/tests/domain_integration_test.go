package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/podland/backend/internal/cloudflare"
	"github.com/podland/backend/internal/domain"
	"github.com/podland/backend/internal/repository"
	"github.com/podland/backend/internal/database"
)

// TestDomainCreationAndDeletion tests the full domain lifecycle
// Requires CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID environment variables
func TestDomainCreationAndDeletion(t *testing.T) {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		t.Skip("CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID must be set for integration tests")
	}

	// Setup
	dnsManager := cloudflare.NewDNSManager(apiToken, zoneID)
	ctx := context.Background()

	// Test domain assignment
	testSubdomain := "test-vm-" + time.Now().Format("20060102150405") + ".podland.app"
	testTarget := "tunnel.podland.app"

	t.Run("CreateDNSRecord", func(t *testing.T) {
		record, err := dnsManager.CreateCNAME(ctx, testSubdomain, testTarget)
		if err != nil {
			t.Fatalf("Failed to create DNS record: %v", err)
		}

		if record.Name != testSubdomain {
			t.Errorf("Expected name %s, got %s", testSubdomain, record.Name)
		}

		if !record.Proxied {
			t.Error("Expected record to be proxied")
		}
	})

	t.Run("GetDNSRecord", func(t *testing.T) {
		record, err := dnsManager.GetRecordByName(ctx, testSubdomain)
		if err != nil {
			t.Fatalf("Failed to get DNS record: %v", err)
		}

		if record.Name != testSubdomain {
			t.Errorf("Expected name %s, got %s", testSubdomain, record.Name)
		}
	})

	t.Run("WaitForDNSActive", func(t *testing.T) {
		err := dnsManager.WaitForDNSActive(ctx, testSubdomain)
		if err != nil {
			t.Fatalf("DNS propagation check failed: %v", err)
		}
	})

	t.Run("DeleteDNSRecord", func(t *testing.T) {
		record, err := dnsManager.GetRecordByName(ctx, testSubdomain)
		if err != nil {
			t.Fatalf("Failed to get record before deletion: %v", err)
		}

		err = dnsManager.DeleteRecord(ctx, record.ID)
		if err != nil {
			t.Fatalf("Failed to delete DNS record: %v", err)
		}

		// Verify deletion
		_, err = dnsManager.GetRecordByName(ctx, testSubdomain)
		if err == nil {
			t.Error("Expected error when getting deleted record")
		}
	})
}

// TestDNSPoller tests the DNS poller functionality
func TestDNSPoller(t *testing.T) {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	dbURL := os.Getenv("DATABASE_URL")

	if apiToken == "" || zoneID == "" || dbURL == "" {
		t.Skip("CLOUDFLARE_API_TOKEN, CLOUDFLARE_ZONE_ID, and DATABASE_URL must be set")
	}

	// Setup
	db, err := database.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	dnsManager := cloudflare.NewDNSManager(apiToken, zoneID)
	vmRepo := repository.NewVMRepository(db)
	dnsPoller := domain.NewDNSPoller(dnsManager, vmRepo)

	ctx := context.Background()

	// Create test DNS record
	testSubdomain := "test-poll-" + time.Now().Format("20060102150405") + ".podland.app"
	_, err = dnsManager.CreateCNAME(ctx, testSubdomain, "tunnel.podland.app")
	if err != nil {
		t.Fatalf("Failed to create DNS record: %v", err)
	}
	defer dnsManager.DeleteRecord(ctx, testSubdomain)

	// Test polling (this should complete quickly since DNS is already active)
	err = dnsPoller.WaitForDNS(ctx, "test-vm-id", testSubdomain)
	if err != nil {
		t.Fatalf("DNS poller failed: %v", err)
	}
}

// TestDomainService tests the domain service layer
func TestDomainService(t *testing.T) {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	dbURL := os.Getenv("DATABASE_URL")

	if apiToken == "" || zoneID == "" || dbURL == "" {
		t.Skip("CLOUDFLARE_API_TOKEN, CLOUDFLARE_ZONE_ID, and DATABASE_URL must be set")
	}

	// Setup
	db, err := database.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	dnsManager := cloudflare.NewDNSManager(apiToken, zoneID)
	vmRepo := repository.NewVMRepository(db)
	_ = domain.NewDomainService(dnsManager, db, vmRepo)
	_ = context.Background()

	// Note: Full domain service testing requires a user and VM to exist
	// This is a placeholder for the test structure
	t.Run("GetDomainsByUserID", func(t *testing.T) {
		// This would require a test user with domains
		// This would require a test user with domains
		// domains, err := domainService.GetDomainsByUserID(ctx, "test-user-id")
		// if err != nil {
		// 	t.Fatalf("Failed to get domains: %v", err)
		// }
		// t.Logf("Found %d domains", len(domains))
		t.Skip("Requires test user with domains")
	})
}
