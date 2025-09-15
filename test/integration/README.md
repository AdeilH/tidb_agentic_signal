# Integration Tests

This directory contains comprehensive integration tests that validate the entire system with real TiDB, Kimi AI, and Binance APIs.

## Prerequisites

### 1. Environment Setup
Create a `.env` file in the project root with:
```bash
KIMI_API_KEY=your_kimi_api_key_here
BINANCE_TEST_KEY=your_binance_test_key_here
BINANCE_TEST_SECRET=your_binance_test_secret_here
TIDB_DSN=root:@tcp(localhost:4000)/sigforge?charset=utf8mb4&parseTime=True&loc=Local
SLACK_WEBHOOK_URL=  # Optional
```

### 2. TiDB Cluster
Start TiDB using docker compose:
```bash
cd /path/to/project
docker compose up -d
```

### 3. API Keys

#### Kimi AI
1. Visit [Kimi.ai](https://kimi.ai) and create an account
2. Get your API key from the dashboard
3. Add it to your `.env` file

#### Binance Testnet
1. Visit [Binance Testnet](https://testnet.binance.vision/)
2. Create an account and generate API keys
3. Add them to your `.env` file

## Running Tests

### Run All Integration Tests
```bash
go test ./test/integration -v
```

### Run Specific Test Suites
```bash
# Test TiDB features only
go test ./test/integration -v -run TestIntegrationFullSystem/TiDB

# Test Kimi AI integration only
go test ./test/integration -v -run TestIntegrationFullSystem/Kimi

# Test Binance integration only
go test ./test/integration -v -run TestIntegrationFullSystem/Binance

# Test end-to-end signal generation
go test ./test/integration -v -run TestIntegrationFullSystem/End_To_End
```

### Run API Integration Tests
Start the server first:
```bash
go run cmd/all/main.go
```

Then in another terminal:
```bash
go test ./test/integration -v -run TestIntegrationAPI
```

### Run Performance Benchmarks
```bash
go test ./test/integration -bench=. -v
```

## Test Coverage

### 🗄️ TiDB Integration Tests
- ✅ Database connection and schema migration
- ✅ TTL features (30-day expiration on event_vecs table)
- ✅ JSON vector storage and retrieval
- ✅ Multi-dimensional vector operations
- ✅ Event and prediction CRUD operations

### 🤖 Kimi AI Integration Tests
- ✅ API connectivity and authentication
- ✅ Market analysis prediction generation
- ✅ Response validation (direction, confidence, logic)
- ✅ Context handling with news and chain data
- ✅ Error handling and timeout management

### 📈 Binance Integration Tests
- ✅ Testnet API connectivity
- ✅ Authentication validation
- ✅ Market data retrieval
- ✅ Server time synchronization
- ✅ API rate limiting compliance

### 🔄 End-to-End Tests
- ✅ Complete signal generation workflow
- ✅ Data ingestion → AI prediction → database storage
- ✅ Multi-component integration validation
- ✅ Real-time data processing simulation

### 🌐 API Integration Tests
- ✅ REST endpoint validation
- ✅ Health check endpoints
- ✅ Manual ingestion triggers
- ✅ Request/response validation
- ✅ Error handling

### ⚡ Performance Tests
- ✅ Event insertion benchmarks
- ✅ Prediction storage benchmarks
- ✅ Vector data operation benchmarks
- ✅ Database performance validation

## Test Features Validated

### TiDB Advanced Features
1. **TTL (Time-To-Live)**
   - Automatic data expiration after 30 days
   - Verified on `event_vecs` table
   - Proper schema configuration validation

2. **JSON Vector Storage**
   - 5-dimensional vector embeddings
   - JSON validation and operations
   - Vector similarity preparation (extensible)

3. **Composite Primary Keys**
   - Multi-tenant data isolation
   - Bot-specific data partitioning
   - Efficient query patterns

### AI Integration Validation
1. **Kimi AI Predictions**
   - Market sentiment analysis
   - Technical indicator processing
   - News sentiment integration
   - Confidence scoring (1-100)

2. **Response Quality**
   - Direction validation (LONG/SHORT/FLAT)
   - Logic reasoning verification
   - Timeout and error handling

### Trading Infrastructure
1. **Binance Testnet Integration**
   - API key validation
   - Market data access
   - Order management preparation
   - Rate limiting compliance

2. **Real-time Data Processing**
   - Event ingestion simulation
   - Prediction generation pipeline
   - Database persistence validation

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   docker compose up -d
   docker ps  # Verify TiDB is running
   ```

2. **API Key Issues**
   - Verify `.env` file exists and contains valid keys
   - Check Kimi AI account status and billing
   - Verify Binance testnet key permissions

3. **Test Timeouts**
   - Check internet connectivity
   - Verify API service availability
   - Increase timeout in test context if needed

4. **Permission Errors**
   ```bash
   chmod +x docker-compose.yml
   sudo docker compose up -d  # If needed
   ```

### Skip Tests in CI
Tests automatically skip if:
- Running in short mode: `go test -short`
- No `.env` file present
- Database unavailable
- API server not running (for API tests)

## Performance Expectations

### Typical Benchmark Results
- Event insertion: ~1000 ops/sec
- Prediction storage: ~800 ops/sec  
- Vector operations: ~500 ops/sec
- API response time: <200ms
- End-to-end signal: <5 seconds

### Resource Usage
- Memory: ~50MB during tests
- CPU: Moderate during AI calls
- Network: ~1MB for full test suite
- Disk: Minimal (test data cleaned up)

## Integration with CI/CD

### GitHub Actions Example
```yaml
name: Integration Tests
on: [push, pull_request]
jobs:
  integration:
    runs-on: ubuntu-latest
    services:
      tidb:
        image: pingcap/tidb:latest
        ports:
          - 4000:4000
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.25
      - name: Run Integration Tests
        env:
          KIMI_API_KEY: ${{ secrets.KIMI_API_KEY }}
          BINANCE_TEST_KEY: ${{ secrets.BINANCE_TEST_KEY }}
          BINANCE_TEST_SECRET: ${{ secrets.BINANCE_TEST_SECRET }}
        run: go test ./test/integration -v
```

This integration test suite provides comprehensive validation of all system components working together with real external services!
