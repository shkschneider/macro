#!/bin/bash
# Integration test script for file type and read-only detection

set -e

MACRO="./macro"
cd "$(dirname "$0")"

echo "Building macro..."
go build -o macro

echo ""
echo "===== Integration Tests ====="
echo ""

# Test 1: Normal file
echo "Test 1: Normal writable text file"
echo "This is a normal text file for testing" > /tmp/integration_normal.txt
chmod 644 /tmp/integration_normal.txt
echo "✓ Created normal writable file"

# Test 2: Read-only file
echo ""
echo "Test 2: Read-only text file"
echo "This is a read-only text file" > /tmp/integration_readonly.txt
chmod 444 /tmp/integration_readonly.txt
echo "✓ Created read-only file (permissions: $(stat -c %a /tmp/integration_readonly.txt))"

# Test 3: Binary file
echo ""
echo "Test 3: Binary file"
dd if=/dev/urandom bs=100 count=1 of=/tmp/integration_binary.bin 2>/dev/null
chmod 644 /tmp/integration_binary.bin
echo "✓ Created binary file"

# Test 4: Empty file
echo ""
echo "Test 4: Empty text file"
touch /tmp/integration_empty.txt
chmod 644 /tmp/integration_empty.txt
echo "✓ Created empty file"

# Test 5: JSON file
echo ""
echo "Test 5: JSON file"
echo '{"name": "test", "value": 123}' > /tmp/integration_test.json
chmod 644 /tmp/integration_test.json
echo "✓ Created JSON file"

echo ""
echo "===== All integration test files created successfully! ====="
echo ""
echo "Files created:"
ls -lah /tmp/integration_*
echo ""
echo "To manually test:"
echo "  ./macro /tmp/integration_normal.txt     # Should open normally"
echo "  ./macro /tmp/integration_readonly.txt   # Should show [READ-ONLY] warning"
echo "  ./macro /tmp/integration_binary.bin     # Should show error about binary file"
echo "  ./macro /tmp/integration_empty.txt      # Should open with empty content"
echo "  ./macro /tmp/integration_test.json      # Should open normally"
