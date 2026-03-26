# NetSuite Mock API

## Executive Summary

A NetSuite REST API mock server which includes:

- ✅ **10 API endpoints** covering customers, invoices, payments, and journal entries
- ✅ **Unit and integration tests**
- ✅ **Zero external dependencies** (Go standard library only)
- ✅ **Pluggable storage architecture** ready for SQLite/PostgreSQL

### What This Does

This mock server mimics NetSuite's REST API behavior, allowing you to:
- Develop and test integration code locally
- Validate business logic with proper error handling

## 📊 Implementation Details

### API Endpoints (10 total)

| Method | Endpoint                                        | Purpose                 |
|--------|-------------------------------------------------|-------------------------|
| POST   | `/services/rest/query/v1/suiteql`               | Execute SuiteQL queries |
| POST   | `/services/rest/record/v1/customer`             | Create customer         |
| GET    | `/services/rest/record/v1/customer/{id}`        | Get customer            |
| PATCH  | `/services/rest/record/v1/customer/{id}`        | Update customer         |
| POST   | `/services/rest/record/v1/invoice`              | Create invoice          |
| GET    | `/services/rest/record/v1/invoice/{id}`         | Get invoice             |
| POST   | `/services/rest/record/v1/customerPayment`      | Create payment          |
| GET    | `/services/rest/record/v1/customerPayment/{id}` | Get payment             |
| POST   | `/services/rest/record/v1/journalEntry`         | Create journal entry    |
| GET    | `/services/rest/record/v1/journalEntry/{id}`    | Get journal entry       |

### Data Models

**Specific models:**
- **Customer** - With address book, subsidiary, inactive status
- **Invoice** - With line items and transaction ID generation
- **Journal Entry** - With debit/credit validation
- **Customer Payment** - With flexible payment amounts
- **Employee** - With household ministry ID
- **Account** - With account number indexing

### Validation Features

**Field Validation**
- Required fields enforced (firstName, lastName, memo)
- Empty externalId allows multiple records
- Non-empty externalId enforces uniqueness (409 Conflict)

**Business Logic**
- Referential integrity (customer/account existence)
- Journal entry debits must equal credits
- Minimum 2 line items per journal entry

**ID Handling**
- Accepts: numbers (100), strings ("100"), objects ({"id": "100"})
- All normalized to int64 internally
- Auto-incrementing per record type
- Transaction IDs auto-generated

## 🧪 Testing

### Test Coverage

- **Unit tests** for storage and utilities
- **Integration tests** for API endpoints
- **Error scenario tests** for all validation rules

### Run Unit Tests

```bash
go test ./...
```

### Key Tests
- Duplicate externalId handling (409 Conflict)
- Empty externalId uniqueness (allowed)
- Customer CRUD operations
- Invoice and payment creation
- Journal entry debit/credit validation
- Flexible ID input parsing
- Query execution

### Run Integration Tests

1. Build

```bash

go build -o netsuite-mock .
```

2. Run

```bash
./netsuite-mock -port 8080
```

3. Test

```bash
# In another terminal
./test-api.sh
```

Or make a single request:
```bash
curl -X POST http://localhost:8080/services/rest/record/v1/customer \
  -H "Content-Type: application/json" \
  -d '{
    "firstName": "John",
    "lastName": "Doe",
    "externalId": "JOHN_001",
    "subsidiary": 1
  }'
```

## Architecture

### Storage Layer (Pluggable)

```go
// Current implementation: In-memory
store := storage.NewInMemoryStore()

// Future: Just implement the Store interface
store := storage.NewSQLiteStore("db.sqlite3")
store := storage.NewPostgresStore(connString)
```

### Handler Layer (Clean)
```
HTTP Handlers
    ↓
Business Logic (Storage)
    ↓
Data Models
    ↓
Utilities (ID Parsing)
```

### Error Handling
All errors return consistent JSON format:
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "type": "CLIENT_ERROR"
  }
}
```

## 🔮 Future Enhancements

The architecture supports easy addition of:

1. **SQL Backend**
   - Implement `Store` interface once
   - Switch with a flag
   - No handler changes needed

2. **Advanced Queries**
   - IN operator support
   - ORDER BY/LIMIT
   - Complex WHERE conditions

3. **Performance**
   - Query caching
   - Bulk operations
   - Pagination support

4. **Security**
   - OAuth token validation
   - Rate limiting
   - Audit logging
