# TiDB-Powered Crypto Signals Bot

A sophisticated, multi-tenant cryptocurrency trading signals bot that maximizes TiDB's advanced features including vector storage, TTL (Time-To-Live), and distributed architecture.

## 🚀 TiDB Features Showcased

### 1. **TTL (## 🧪 Testing

### Unit Tests
```bash
# Run all unit tests
go test ./... -v

# Run tests for specific package
go test ./internal/db -v
go test ./internal/predictor -v
```

### Integration Tests
Comprehensive integration tests validate the complete system with real TiDB, Kimi AI, and Binance APIs.

```bash
# Run all integration tests
./test/integration/run_tests.sh

# Run specific test suites
./test/integration/run_tests.sh tidb      # TiDB features only
./test/integration/run_tests.sh ai        # Kimi AI integration
./test/integration/run_tests.sh binance   # Binance API testing
./test/integration/run_tests.sh e2e       # End-to-end signal generation
./test/integration/run_tests.sh api       # HTTP API testing
./test/integration/run_tests.sh bench     # Performance benchmarks
```

#### What Gets Tested
- ✅ **TiDB Features**: TTL, vector storage, multi-tenant isolation
- ✅ **AI Integration**: Real Kimi AI prediction generation
- ✅ **External APIs**: Binance testnet connectivity and data
- ✅ **End-to-End Flow**: Complete signal generation pipeline
- ✅ **Performance**: Database operation benchmarks
- ✅ **API Endpoints**: REST API and WebSocket functionality

For detailed testing documentation, see [INTEGRATION_TESTING.md](INTEGRATION_TESTING.md).

## 🔧 Development

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- Valid API keys (Kimi AI, Binance TestNet)

### Quick Development Setup
```bash
# 1. Clone and setup
git clone <repository-url>
cd agentic_go_signals
cp .env.example .env
# Edit .env with your API keys

# 2. Start TiDB cluster
docker-compose up -d

# 3. Run tests to validate setup
./test/integration/run_tests.sh

# 4. Start development server
go run cmd/all/main.go
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (if golangci-lint installed)
golangci-lint run
```

## 🔐 Security Featuresime-To-Live) Data Management**
- **Automatic Data Expiration**: Events and event vectors automatically expire after 30 days
- **Storage Optimization**: Prevents database bloat with automatic cleanup
- **Implementation**: Raw SQL with `TTL = 30 DAY` on `event_vecs` table

```sql
CREATE TABLE event_vecs (
    -- columns...
) TTL = ts + INTERVAL 30 DAY;
```

### 2. **Vector Storage with JSON**
- **High-Dimensional Data**: Store 128-dimensional vectors for semantic search
- **Flexible Schema**: JSON column allows varying vector dimensions
- **Implementation**: Custom vector generation and storage pipeline

### 3. **Multi-Tenant Architecture**
- **Composite Primary Keys**: `(bot_id, id)` ensures tenant isolation
- **Horizontal Scaling**: Each bot operates independently
- **Data Separation**: Clean separation between different trading bots

### 4. **Distributed Database Support**
- **TiKV Integration**: Leverages TiDB's distributed storage layer
- **PD Cluster**: Placement Driver for metadata management
- **Docker Compose**: Local TiDB cluster with PD, TiKV, and TiDB components

### 5. **Real-Time WebSocket Updates**
- **Live Data Streaming**: Real-time updates via WebSocket connections
- **Event Broadcasting**: Ingestion, prediction, and trade events
- **Scalable Architecture**: Hub pattern for managing multiple connections

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Data Sources  │    │   TiDB Cluster   │    │   API Gateway   │
│                 │    │                  │    │                 │
│ • CryptoCompare │───▶│ • TTL Tables     │◀───│ • Fiber v2 API  │
│ • Blockchain.io │    │ • Vector Storage │    │ • WebSocket Hub │
│ • Market Data   │    │ • Multi-Tenant   │    │ • Real-time     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ Ingest Pipeline │    │ Predictor AI     │    │ Risk Manager    │
│                 │    │                  │    │                 │
│ • News Fetcher  │    │ • Kimi AI        │    │ • Position Size │
│ • Chain Metrics │    │ • Context Build  │    │ • Stop Loss     │
│ • Vector Gen    │    │ • DB Persistence │    │ • Risk Limits   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌──────────────────┐
                    │ Trading Engine   │
                    │                  │
                    │ • Binance TestNet│
                    │ • Paper Trading  │
                    │ • Orchestrator   │
                    └──────────────────┘
```

## 📦 Project Structure

```
agentic_go_signals/
├── cmd/all/                    # Main application entry point
├── internal/
│   ├── api/                    # Fiber v2 REST API + WebSocket
│   ├── chain/                  # Blockchain metrics (blockchain.info)
│   ├── config/                 # Environment configuration
│   ├── db/                     # TiDB models and migrations
│   ├── ingest/                 # Data ingestion pipeline
│   ├── kimi/                   # Kimi AI client (Moonshot)
│   ├── news/                   # News fetcher (CryptoCompare)
│   ├── notifications/          # Slack notifications (configurable)
│   ├── predictor/              # AI prediction service
│   ├── risk/                   # Risk management calculator
│   ├── svc/                    # Service initialization
│   ├── trader/                 # Binance TestNet integration
│   └── worker/                 # Orchestrator for pipeline
├── docker-compose.yml          # TiDB cluster setup
├── .env.example               # Environment variables template
└── README.md                  # This file
```

## 🛠️ Installation & Setup

### 1. Clone Repository
```bash
git clone <repository-url>
cd agentic_go_signals
```

### 2. Start TiDB Cluster
```bash
docker-compose up -d
```

This starts:
- **TiDB Server** (port 4000): SQL interface
- **PD Server** (port 2379): Placement Driver
- **TiKV Server** (port 20160): Distributed storage

### 3. Configure Environment
```bash
cp .env.example .env
# Edit .env with your API keys:
# - KIMI_API_KEY=your_moonshot_api_key
# - BINANCE_TEST_KEY=your_testnet_key
# - BINANCE_TEST_SECRET=your_testnet_secret
# - SLACK_WEBHOOK_URL=optional_slack_webhook (leave empty to disable)
```

### 4. Build & Run
```bash
go build -o bin/sigforge ./cmd/all
./bin/sigforge
```

## 🧪 Testing

Run all tests across the project:
```bash
go test ./... -v
```

Individual package tests:
```bash
go test ./internal/db/ -v      # Database models
go test ./internal/api/ -v     # API endpoints
go test ./internal/risk/ -v    # Risk calculations
go test ./internal/worker/ -v  # Orchestrator
```

## 🔗 API Endpoints

### Core Endpoints
- `GET /healthz` - Health check
- `POST /bot/create` - Create new trading bot
- `GET /bot/:id` - Get bot details
- `GET /bot/:id/signals` - Get trading signals

### Data Endpoints
- `POST /ingest/manual` - Manual data ingestion
- `GET /predictions/:bot_id/:symbol` - Get predictions
- `GET /trades/:bot_id` - Get trade history

### WebSocket
- `WS /ws` - Real-time updates stream

## 📊 TiDB Schema Design

### Events Table (TTL Enabled)
```sql
CREATE TABLE events (
    id BIGINT AUTO_INCREMENT,
    bot_id VARCHAR(50),
    ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    symbol VARCHAR(20),
    source VARCHAR(50),
    usd_val DECIMAL(20,8),
    text TEXT,
    PRIMARY KEY (bot_id, id)
);

CREATE TABLE event_vecs (
    id BIGINT AUTO_INCREMENT,
    bot_id VARCHAR(50),
    ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sym VARCHAR(20),
    vec JSON,
    text TEXT,
    PRIMARY KEY (bot_id, id)
) TTL = ts + INTERVAL 30 DAY;
```

### Multi-Tenant Design
- **Composite Keys**: `(bot_id, id)` ensures data isolation
- **Partitioning**: Each bot's data is logically separated
- **Scalability**: Add new bots without schema changes

## 🤖 AI Integration

### Kimi AI (Moonshot)
- **Context Building**: Combines news and chain metrics
- **Prediction Generation**: AI-driven market analysis
- **Confidence Scoring**: Quantified prediction confidence

### Vector Search (Future)
- **Semantic Similarity**: Find similar market conditions
- **Pattern Recognition**: Historical pattern matching
- **Contextual Retrieval**: Enhanced AI context

## ⚖️ Risk Management

### Position Sizing
- **Account Balance Based**: Percentage of total capital
- **Risk Per Trade**: Configurable risk tolerance (e.g., 2%)
- **Stop Loss Calculation**: Automatic stop loss prices

### Risk Limits
- **Maximum Position Size**: Per-trade position limits
- **Total Exposure**: Portfolio-wide exposure limits
- **Validation**: Pre-trade risk checks

## 🔄 Pipeline Orchestration

### Worker Coordination
1. **Ingestion Worker**: Fetches news and chain metrics (5min interval)
2. **Prediction Worker**: Generates AI predictions (10min interval)
3. **Execution Worker**: Evaluates and executes trades (1min interval)

### Data Flow
```
News/Chain Data → Ingestion → Vector Storage → AI Analysis → Risk Check → Trade Execution
```

## 🚀 Deployment Considerations

### Scaling TiDB
- **Horizontal Scaling**: Add TiKV nodes for storage
- **Read Replicas**: TiDB nodes for read scaling
- **Regional Deployment**: Multi-region setup for latency

### Application Scaling
- **Stateless Design**: Easy horizontal scaling
- **Bot Isolation**: Independent bot operations
- **WebSocket Clustering**: Load balancer with sticky sessions

## 📈 Performance Features

### TiDB Optimizations
- **Clustered Index**: Primary key clustering for performance
- **TTL Automation**: Automatic cleanup reduces maintenance
- **JSON Indexing**: Efficient vector column operations

### Application Optimizations
- **Connection Pooling**: Efficient database connections
- **Async Processing**: Non-blocking pipeline operations
- **Concurrent Workers**: Parallel processing capabilities

## � Configurable Notifications

### Slack Integration
- **Optional Configuration**: Set `SLACK_WEBHOOK_URL` environment variable to enable
- **Smart Fallback**: Gracefully skips notifications when not configured
- **Rich Messages**: Trading signals, predictions, and error alerts
- **Formatted Updates**: Color-coded messages with timestamps

### Notification Types
- **Trade Execution**: Real-time trading signal notifications
- **AI Predictions**: Market prediction alerts with confidence scores
- **System Errors**: Error alerts for debugging and monitoring
- **Custom Messages**: Flexible message structure for future extensions

## �🔐 Security Features

### Data Protection
- **Environment Variables**: Secure API key management
- **TestNet Only**: Safe paper trading environment
- **Input Validation**: SQL injection prevention

### Multi-Tenancy Security
- **Bot Isolation**: Complete data separation
- **Access Control**: Bot-specific data access
- **Audit Trail**: Complete trade history tracking

## 🎯 Next Steps

1. **Enhanced Vector Search**: Implement semantic similarity queries
2. **Real Market Data**: Integrate live price feeds
3. **Advanced AI Models**: Multi-model ensemble predictions
4. **Performance Monitoring**: TiDB cluster monitoring
5. **Production Deployment**: Kubernetes orchestration

---

## 📄 License

This project is part of a TiDB hackathon submission showcasing advanced database features in a real-world application.

## 🤝 Contributing

This is a hackathon project. For production use, consider:
- Enhanced error handling
- Comprehensive logging
- Production-grade security
- Performance optimization
- Monitoring and alerting
