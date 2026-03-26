#!/bin/bash
# Test script for NetSuite Mock API
# Run this script to verify the server is working correctly

API_URL="http://localhost:8080/services/rest"

echo "=========================================="
echo "NetSuite Mock API - Test Suite"
echo "=========================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

failures=0

# Helper function to print test results
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4

    echo -e "${YELLOW}Testing:${NC} $method $endpoint"

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint" \
            -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    exit_code=$?
    if [ $exit_code -ne 0 ]; then
      echo -e "${RED}x curl returned a non-zero exit code: $exit_code${NC}"
      echo ""
      failures=$((failures + 1))
      return
    fi

    status=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}✓ Status: $status${NC}"
        echo "Response: $body" | head -c 200
        echo ""
    else
        echo -e "${RED}✗ Expected: $expected_status, Got: $status${NC}"
        echo "Response: $body"
        echo ""
        failures=$((failures + 1))
    fi
}

curl --head --silent $API_URL > /dev/null
exit_code=$?
if [ $exit_code -ne 0 ]; then
  echo -e "${RED}x The mock server is not responding correctly. Is it running?${NC}"
  exit 1
fi

echo "1. Creating Customer..."
CUSTOMER_DATA='{
  "firstName": "John",
  "lastName": "Doe",
  "externalId": "JOHN_DOE_001",
  "subsidiary": 1
}'
test_endpoint "POST" "/record/v1/customer" "$CUSTOMER_DATA" "204"

echo ""
echo "2. Creating another customer (same externalId - should fail)..."
DUPLICATE_DATA='{
  "firstName": "Jane",
  "lastName": "Smith",
  "externalId": "JOHN_DOE_001",
  "subsidiary": 1
}'
test_endpoint "POST" "/record/v1/customer" "$DUPLICATE_DATA" "409"

echo ""
echo "3. Getting Customer (ID=1)..."
test_endpoint "GET" "/record/v1/customer/1" "" "200"

echo ""
echo "4. Updating Customer (activate)..."
UPDATE_DATA='{
  "isInactive": false
}'
test_endpoint "PATCH" "/record/v1/customer/1" "$UPDATE_DATA" "204"

echo ""
echo "5. Creating Invoice..."
INVOICE_DATA='{
  "entity": 1,
  "date": "2024-01-31T00:00:00Z",
  "subsidiary": 1,
  "currency": 1,
  "memo": "Test Invoice"
}'
test_endpoint "POST" "/record/v1/invoice" "$INVOICE_DATA" "204"

echo ""
echo "6. Creating Journal Entry (valid)..."
JOURNAL_DATA='{
  "memo": "Test Entry",
  "subsidiary": 1,
  "line": {
    "items": [
      {"debit": 100},
      {"credit": 100}
    ]
  }
}'
test_endpoint "POST" "/record/v1/journalEntry" "$JOURNAL_DATA" "204"

echo ""
echo "7. Creating Journal Entry (invalid - unbalanced)..."
INVALID_JOURNAL='{
  "memo": "Bad Entry",
  "subsidiary": 1,
  "line": {
    "items": [
      {"debit": 100},
      {"credit": 50}
    ]
  }
}'
test_endpoint "POST" "/record/v1/journalEntry" "$INVALID_JOURNAL" "400"

echo ""
echo "8. Running SuiteQL Query..."
QUERY_DATA='{
  "q": "SELECT id,firstName FROM customer WHERE externalId = '\''JOHN_DOE_001'\''"
}'
test_endpoint "POST" "/query/v1/suiteql" "$QUERY_DATA" "200"

echo ""
echo "=========================================="
echo "All tests completed!"
echo "=========================================="

if [ "$failures" -ne 0 ]; then
    echo -e "${RED}Tests failed:${NC} $failures failure(s)."
    exit 1
fi

echo -e "${GREEN}All tests passed!${NC}"
