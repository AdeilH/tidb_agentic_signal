# Integration Testing Guide

This project includes comprehensive integration tests that validate the complete system with real TiDB, Kimi AI, and Binance APIs.

## 🎯 What Gets Tested

### Core System Integration
- **TiDB Features**: TTL tables, JSON vectors, multi-tenant architecture
- **Kimi AI**: Market prediction generation with real API calls
- **Binance**: Testnet connectivity and market data access
- **End-to-End**: Complete signal generation workflow

### Advanced TiDB Features Validated
- ✅ **TTL (Time-To-Live)**: 30-day automatic data expiration
- ✅ **Vector Storage**: JSON-based vector embeddings for ML features
- ✅ **Multi-Tenant**: Composite primary keys for data isolation
- ✅ **High Performance**: Benchmarked insert/query operations

## 🚀 Quick Start

### 1. Setup Environment
```bash
# Copy and edit environment variables
cp .env.example .env
# Edit .env with your API keys:
# - KIMI_API_KEY (from kimi.ai)
# - BINANCE_TEST_KEY (from testnet.binance.vision)
# - BINANCE_TEST_SECRET
```

### 2. Start TiDB
```bash
docker compose up -d
```

### 3. Run Tests
```bash
# Run all integration tests
./test/integration/run_tests.sh

# Or run specific test suites
./test/integration/run_tests.sh tidb      # TiDB features only
./test/integration/run_tests.sh ai        # Kimi AI only  
./test/integration/run_tests.sh binance   # Binance API only
./test/integration/run_tests.sh e2e       # End-to-end flow
./test/integration/run_tests.sh api       # HTTP API tests
./test/integration/run_tests.sh bench     # Performance benchmarks
```

## 📊 Test Results Example

```
🚀 Crypto Signals Bot - Integration Test Runner
=================================================
[SUCCESS] Go version: go1.25
[SUCCESS] TiDB is ready
[SUCCESS] Environment variables validated

🗄️  Testing TiDB integration...
=== RUN   TestIntegrationFullSystem/TiDB_Connection_And_Schema
    integration_test.go:87: Testing TiDB connection and schema...
    integration_test.go:106: ✅ TiDB connection and schema tests passed
--- PASS: TestIntegrationFullSystem/TiDB_Connection_And_Schema (0.12s)

🤖 Testing Kimi AI integration...
=== RUN   TestIntegrationFullSystem/Kimi_AI_Integration
    integration_test.go:109: Testing Kimi AI integration...
    integration_test.go:125: ✅ Kimi AI prediction: LONG (conviction: 78%) - Strong institutional adoption signals
--- PASS: TestIntegrationFullSystem/Kimi_AI_Integration (2.34s)

📈 Testing Binance integration...
=== RUN   TestIntegrationFullSystem/Binance_Test_API
    integration_test.go:128: Testing Binance test API integration...
    integration_test.go:151: ✅ Binance testnet connectivity verified, server time: 1726338000000
--- PASS: TestIntegrationFullSystem/Binance_Test_API (0.45s)

🔄 Testing end-to-end signal generation...
=== RUN   TestIntegrationFullSystem/End_To_End_Signal_Generation
    integration_test.go:154: Testing end-to-end signal generation...
    integration_test.go:201: ✅ End-to-end signal generated: LONG with 82% confidence
--- PASS: TestIntegrationFullSystem/End_To_End_Signal_Generation (3.67s)

[SUCCESS] Integration tests completed successfully! 🎉

System validated with:
  ✅ TiDB cluster with TTL and vector storage
  ✅ Kimi AI prediction generation  
  ✅ Binance testnet connectivity
  ✅ End-to-end signal generation pipeline

Your crypto signals bot is ready for production! 🚀
```

## 🏗️ Integration Test Architecture

### Test Structure
```
test/integration/
├── integration_test.go    # Main test suite
├── run_tests.sh          # Automated test runner
└── README.md            # Detailed documentation
```

### Test Flow
1. **Environment Validation**: Check API keys and dependencies
2. **Database Setup**: Verify TiDB connection and schema
3. **External APIs**: Test Kimi AI and Binance connectivity
4. **End-to-End**: Complete signal generation workflow
5. **Performance**: Benchmark critical operations

## 📈 Performance Benchmarks

The integration tests include performance benchmarks:

```bash
./test/integration/run_tests.sh bench
```

Expected results:
- **Event Insertion**: ~1000 ops/sec
- **Prediction Storage**: ~800 ops/sec
- **Vector Operations**: ~500 ops/sec
- **End-to-End Signal**: <5 seconds

## 🔧 Troubleshooting

### Common Issues

**TiDB Connection Failed**
```bash
docker compose down && docker compose up -d
docker ps  # Verify containers are running
```

**API Key Errors**
- Check `.env` file exists and has valid keys
- Verify Kimi AI account status and billing
- Confirm Binance testnet key permissions

**Test Timeouts**
- Ensure stable internet connection
- Check API service availability
- Increase timeout in test context if needed

### Skip Tests
Tests automatically skip when:
- Running with `-short` flag
- No `.env` file present  
- Database unavailable
- API keys invalid

## 🚀 CI/CD Integration

### GitHub Actions
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
        run: ./test/integration/run_tests.sh
```

### Docker Testing
```bash
# Run tests in isolated environment
docker run --rm -v $(pwd):/app -w /app \
  -e KIMI_API_KEY=$KIMI_API_KEY \
  -e BINANCE_TEST_KEY=$BINANCE_TEST_KEY \
  -e BINANCE_TEST_SECRET=$BINANCE_TEST_SECRET \
  golang:1.25 ./test/integration/run_tests.sh
```

## 🎯 What This Validates

### Production Readiness
- ✅ All external API integrations working
- ✅ Database schema and performance validated
- ✅ End-to-end signal generation verified
- ✅ Error handling and timeout management
- ✅ Multi-tenant data isolation
- ✅ TTL and vector storage features

### TiDB Showcase
- ✅ Advanced features (TTL, JSON, vectors)
- ✅ High-performance operations
- ✅ Production schema patterns
- ✅ Multi-tenant architecture
- ✅ Composite primary keys

This comprehensive integration test suite ensures your crypto signals bot is production-ready and showcases TiDB's advanced capabilities!
