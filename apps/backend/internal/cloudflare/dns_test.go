package cloudflare

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestDNSManager tests the DNS manager CRUD operations
// Note: These are integration tests that require a real Cloudflare API token
// For unit tests, use a mock Cloudflare client
func TestDNSManager(t *testing.T) {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		t.Skip("CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID must be set for integration tests")
	}

	manager := NewDNSManager(apiToken, zoneID)
	ctx := context.Background()

	// Test CNAME creation
	testSubdomain := "test-vm-" + time.Now().Format("20060102150405") + ".podland.app"
	testTarget := "tunnel.podland.app"

	t.Run("CreateCNAME", func(t *testing.T) {
		record, err := manager.CreateCNAME(ctx, testSubdomain, testTarget)
		if err != nil {
			t.Fatalf("Failed to create CNAME: %v", err)
		}

		if record.Name != testSubdomain {
			t.Errorf("Expected name %s, got %s", testSubdomain, record.Name)
		}

		if record.Content != testTarget {
			t.Errorf("Expected content %s, got %s", testTarget, record.Content)
		}

		if !record.Proxied {
			t.Error("Expected record to be proxied")
		}

		if record.TTL != 1 {
			t.Errorf("Expected TTL 1 (auto), got %d", record.TTL)
		}
	})

	t.Run("GetRecordByName", func(t *testing.T) {
		record, err := manager.GetRecordByName(ctx, testSubdomain)
		if err != nil {
			t.Fatalf("Failed to get record by name: %v", err)
		}

		if record.Name != testSubdomain {
			t.Errorf("Expected name %s, got %s", testSubdomain, record.Name)
		}
	})

	t.Run("ListRecords", func(t *testing.T) {
		records, err := manager.ListRecords(ctx, testSubdomain, "CNAME")
		if err != nil {
			t.Fatalf("Failed to list records: %v", err)
		}

		if len(records) != 1 {
			t.Fatalf("Expected 1 record, got %d", len(records))
		}

		if records[0].Name != testSubdomain {
			t.Errorf("Expected name %s, got %s", testSubdomain, records[0].Name)
		}
	})

	t.Run("UpdateRecord", func(t *testing.T) {
		// Get the record first
		record, err := manager.GetRecordByName(ctx, testSubdomain)
		if err != nil {
			t.Fatalf("Failed to get record: %v", err)
		}

		// Update content
		newTarget := "new-tunnel.podland.app"
		updated, err := manager.UpdateRecord(ctx, record.ID, newTarget)
		if err != nil {
			t.Fatalf("Failed to update record: %v", err)
		}

		if updated.Content != newTarget {
			t.Errorf("Expected content %s, got %s", newTarget, updated.Content)
		}
	})

	t.Run("DeleteRecord", func(t *testing.T) {
		// Get the record first
		record, err := manager.GetRecordByName(ctx, testSubdomain)
		if err != nil {
			t.Fatalf("Failed to get record: %v", err)
		}

		// Delete
		err = manager.DeleteRecord(ctx, record.ID)
		if err != nil {
			t.Fatalf("Failed to delete record: %v", err)
		}

		// Verify deletion
		_, err = manager.GetRecordByName(ctx, testSubdomain)
		if err == nil {
			t.Error("Expected error when getting deleted record")
		}
	})
}

// TestWaitForDNSActive tests the DNS propagation polling
func TestWaitForDNSActive(t *testing.T) {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		t.Skip("CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID must be set for integration tests")
	}

	manager := NewDNSManager(apiToken, zoneID)
	ctx := context.Background()

	// Create a test record
	testSubdomain := "test-poll-" + time.Now().Format("20060102150405") + ".podland.app"
	testTarget := "tunnel.podland.app"

	record, err := manager.CreateCNAME(ctx, testSubdomain, testTarget)
	if err != nil {
		t.Fatalf("Failed to create CNAME: %v", err)
	}
	defer manager.DeleteRecord(ctx, record.ID)

	// Test polling
	err = manager.WaitForDNSActive(ctx, testSubdomain)
	if err != nil {
		t.Fatalf("DNS propagation check failed: %v", err)
	}
}
