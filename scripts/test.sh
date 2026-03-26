#!/bin/bash

# Podland Phase 1 Test Script
# Run all tests for Phase 1

set -e

echo "=== Podland Phase 1 Tests ==="
echo ""

# Backend tests
echo "Running backend tests..."
cd apps/backend
go test ./... -v
cd ../..

echo ""
echo "=== All tests passed! ==="
