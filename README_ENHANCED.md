# 🚀 TiDB Agentic Crypto Signals - Advanced Trading Analytics Platform

A sophisticated crypto trading signals platform leveraging **TiDB's advanced OLAP capabilities**, **Kimi AI**, and **real-time market data** to provide intelligent trading recommendations with comprehensive market analytics.

## 🌟 Key Features

### 🔥 **TiDB Advanced Analytics**
- **Window Functions** for price trend analysis with SMA/EMA calculations
- **Real-time Volume Surge Detection** using statistical aggregations
- **Support/Resistance Level Calculations** with price action analysis
- **Momentum Indicators** with 5-minute and 15-minute timeframes
- **Advanced SQL Queries** leveraging TiDB's OLAP capabilities

### 🤖 **Enhanced Kimi AI Integration**
- **Sophisticated Prompts** incorporating comprehensive TiDB analytics
- **Dynamic Trading Recommendations** (BUY/SELL/HOLD) with confidence scores
- **Entry Timing Analysis** with market condition assessment
- **Risk Management Signals** based on volatility and volume patterns
- **Real-time AI Analysis** using live market data and historical patterns

### 📊 **Real-time Market Data**
- **Binance Production WebSocket Streams** (wss://stream.binance.com:9443)
- **Live Order Book Display** with bid/ask depth visualization
- **Real-time Price Updates** with 24h change tracking
- **Volume Analysis** with surge detection and flow patterns
- **Multi-symbol Support** for major crypto pairs

### 🎨 **Advanced Frontend Dashboard**
- **Two-column Layout** with AI recommendations and order book panels
- **Enhanced Visualizations** with real-time charts and signals
- **Interactive Controls** for TiDB analytics and AI signal generation
- **Professional Styling** with dark theme and gradient effects
- **Responsive Design** optimized for trading workflows

## 📋 Quick Start

### Prerequisites
- **Go 1.25+**
- **TiDB Cluster** (local or cloud)
- **Kimi AI API Key**
- **Binance API Credentials** (optional, for enhanced features)

### 1. Clone and Setup
```bash
git clone <repository-url>
cd agentic_go_signals
cp .env.example .env
```

### 2. Configure Environment
```env
# TiDB Configuration
TIDB_HOST=127.0.0.1
TIDB_PORT=4000
TIDB_USER=root
TIDB_PASSWORD=
TIDB_DATABASE=crypto_signals

# Kimi AI Configuration
KIMI_API_KEY=your_kimi_api_key
KIMI_BASE_URL=https://api.moonshot.cn/v1

# Binance API (Optional)
BINANCE_API_KEY=your_binance_api_key
BINANCE_API_SECRET=your_binance_secret
```

### 3. Build and Run
```bash
# Build the application
make build

# Run with all features
./crypto-signals

# Or run in development mode
make run
```

### 4. Access the Dashboard
Open your browser to: **http://localhost:3333**

## 🏗️ Architecture Overview

### Core Components

#### **TiDB Analytics Engine** (`internal/db/market_data.go`)
- **GetAdvancedSignals()**: Complex window function queries for market analysis
- **GetRealTimeMarketState()**: Real-time volume and volatility detection
- **Sophisticated SQL**: Leverages TiDB's analytical capabilities with STDDEV, LAG, AVG
- **Performance Optimized**: Efficient queries designed for real-time trading

#### **Enhanced API Layer** (`internal/api/api.go`)
- **Advanced Analytics Endpoints**: `/analytics/advanced`, `/analytics/realtime`
- **Enhanced Kimi Integration**: Sophisticated prompt engineering with TiDB data
- **Market Analysis Helpers**: Trend analysis, risk assessment, order flow calculations
- **RESTful Design**: Clean API structure for frontend integration

#### **Intelligent Frontend** (`web/index.html`)
- **AI Recommendations Panel**: Real-time display of Kimi AI signals with confidence scores
- **Order Book Visualization**: Live bid/ask depth with color-coded pricing
- **Advanced Controls**: Dedicated buttons for TiDB analytics and enhanced AI signals
- **Professional Styling**: Modern dark theme with animated elements

### Data Flow
```
Binance WebSocket → TiDB Storage → Analytics Engine → Kimi AI → Frontend Dashboard
                                        ↓
                              Real-time Signals & Recommendations
```

## 🎯 Advanced Features

### **TiDB Window Functions**
```sql
-- SMA Cross Detection
SELECT symbol, price, 
       AVG(price) OVER (PARTITION BY symbol ORDER BY timestamp 
                        ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as sma_20,
       LAG(price, 1) OVER (PARTITION BY symbol ORDER BY timestamp) as prev_price
FROM market_data 
WHERE timestamp >= NOW() - INTERVAL 1 HOUR
```

### **Volume Surge Detection**
```sql
-- Real-time Volume Analysis
SELECT symbol, volume,
       AVG(volume) OVER (PARTITION BY symbol ORDER BY timestamp 
                         ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as avg_volume,
       STDDEV(volume) OVER (PARTITION BY symbol ORDER BY timestamp 
                            ROWS BETWEEN 19 PRECEDING AND CURRENT ROW) as vol_stddev
```

### **Enhanced Kimi AI Prompts**
```go
prompt := fmt.Sprintf(`
Advanced Crypto Trading Analysis for %s

MARKET DATA (Last 5 minutes):
- Current Price: $%.2f
- Volume Ratio: %.2f (vs 20-period avg)
- SMA Cross: %s
- Momentum (5min): %.2f%%
- Support: $%.2f | Resistance: $%.2f

REAL-TIME SIGNALS:
- Volume Surge: %.1fx normal
- Buy Pressure: %.1f%%
- Volatility Spike: %v

Provide trading recommendation with:
1. Action: BUY/SELL/HOLD
2. Confidence: 0-100%%
3. Entry timing score
4. Risk assessment
`, symbol, price, volumeRatio, smaSignal, momentum, support, resistance, 
   volumeSurge, buyPressure, volatilitySpike)
```

## 🔧 API Reference

### TiDB Analytics Endpoints

#### **GET /analytics/advanced**
Returns sophisticated market signals using TiDB window functions
```json
{
  "success": true,
  "signals": [
    {
      "symbol": "BTCUSDT",
      "sma_cross": true,
      "volume_ratio": 0.78,
      "momentum_5min": 2.34,
      "support_level": 42150.50,
      "resistance_level": 43250.75,
      "timestamp": "2025-09-14T22:57:00Z"
    }
  ]
}
```

#### **GET /analytics/realtime**
Real-time market state with volume surge and volatility detection
```json
{
  "success": true,
  "market_state": [
    {
      "symbol": "BTCUSDT",
      "volume_surge": 1.84,
      "buy_pressure": 0.67,
      "volatility_spike": false,
      "order_flow_score": 0.72,
      "timestamp": "2025-09-14T22:57:00Z"
    }
  ]
}
```

#### **GET /kimi/enhanced/{symbol}**
Enhanced Kimi AI recommendations with comprehensive TiDB analytics
```json
{
  "success": true,
  "data": {
    "recommendation": "BUY",
    "confidence": "78%",
    "reasoning": "Strong volume surge with golden cross formation",
    "enhanced_data": {
      "urgency": "HIGH",
      "entry_timing_score": 0.82
    },
    "tidb_analytics": {
      "sma_cross": true,
      "volume_ratio": 0.78,
      "momentum_5min": 2.34
    },
    "realtime_state": {
      "volume_surge": 1.84,
      "buy_pressure": 0.67
    }
  }
}
```

## 🎨 Frontend Features

### **AI Recommendations Panel**
- **Enhanced Display**: Shows recommendations with confidence scores and timing analysis
- **Advanced Signals**: Displays TiDB analytics including SMA cross, volume ratio, momentum
- **Real-time Updates**: Live data with urgency badges and color-coded signals
- **Professional Styling**: Gradient backgrounds, hover effects, and animated elements

### **TiDB Analytics Controls**
- **🔍 Advanced Signals**: Trigger complex TiDB window function analysis
- **⚡ Real-Time State**: Get current market volatility and volume surge data  
- **🤖 Enhanced Kimi AI**: Generate sophisticated AI recommendations with full analytics

### **Order Book Visualization**
- **Live Updates**: Real-time bid/ask depth display
- **Color Coding**: Green for bids, red for asks with intensity gradients
- **Interactive Controls**: Symbol selection and refresh capabilities

## 🚀 Performance & Scalability

### **TiDB Optimizations**
- **Efficient Indexing**: Optimized indexes on (symbol, timestamp) for fast queries
- **Window Function Performance**: Leverages TiDB's columnar storage for analytics
- **Connection Pooling**: Efficient database connection management
- **Query Optimization**: Carefully crafted SQL for minimal resource usage

### **Real-time Capabilities**
- **WebSocket Performance**: Efficient binary message handling
- **Memory Management**: Optimized data structures for live updates
- **Concurrent Processing**: Go routines for parallel market data processing
- **Rate Limiting**: Intelligent throttling to prevent API overload

## 🔮 Advanced Use Cases

### **Algorithmic Trading Integration**
The platform provides sophisticated signals that can be integrated into:
- **Automated Trading Bots** using confidence-based position sizing
- **Risk Management Systems** with volatility and volume surge alerts
- **Portfolio Optimization** using momentum and trend analysis
- **Market Making Strategies** with order book depth analysis

### **Research & Analytics**
- **Backtesting Framework** using historical TiDB data
- **Strategy Development** with comprehensive market indicators
- **Risk Analysis** using volatility and correlation metrics
- **Performance Attribution** with detailed trade analysis

## 🛠️ Development

### **Testing**
```bash
# Run all tests
make test

# Run integration tests
make test-integration

# Run specific test suites
go test ./internal/db -v
go test ./internal/api -v
```

### **Development Mode**
```bash
# Run with auto-reload
make dev

# Run with debug logging
DEBUG=true make run

# Run tests with coverage
make coverage
```

### **Building for Production**
```bash
# Build optimized binary
make build-prod

# Build Docker image
make docker-build

# Deploy with Docker Compose
make deploy
```

## 📊 Monitoring & Observability

### **Metrics Endpoint**
- **GET /metrics**: Prometheus-compatible metrics
- **Performance Monitoring**: Request latency, database query times
- **Business Metrics**: Signal accuracy, recommendation performance
- **System Health**: Memory usage, connection pool status

### **Logging**
- **Structured Logging**: JSON format with correlation IDs
- **Debug Levels**: Configurable logging for development vs production
- **Performance Tracking**: Query execution times and API response metrics
- **Error Tracking**: Comprehensive error logging with stack traces

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **TiDB Team** for the powerful distributed SQL database with advanced analytics
- **Kimi AI** for sophisticated language model capabilities
- **Binance** for comprehensive crypto market data APIs
- **Go Community** for excellent libraries and ecosystem support

---

**Built with ❤️ using TiDB's advanced OLAP capabilities and cutting-edge AI integration**
