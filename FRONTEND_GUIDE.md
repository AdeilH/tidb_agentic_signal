# Frontend Development Guide - Crypto Signals API

This guide provides everything you need to build a frontend for the TiDB-powered crypto signals bot.

## 🚀 Quick Start

### Base Configuration
```javascript
const API_BASE_URL = 'http://localhost:3333';
const WS_URL = 'ws://localhost:3333/ws';
```

### Authentication
Currently no authentication required (development mode).

## 📡 REST API Endpoints

### Health Check
```http
GET /healthz
```
**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-09-14T18:30:00Z",
  "version": "v0.1.0"
}
```

### Bot Management

#### Create Bot
```http
POST /bot/create
Content-Type: application/json

{
  "name": "My Trading Bot",
  "symbols": ["BTCUSDT", "ETHUSDT"],
  "risk_params": {
    "account_balance": 10000,
    "risk_per_trade": 0.02,
    "max_position_size": 0.10,
    "stop_loss_percent": 0.05
  }
}
```

#### Get Bot Details
```http
GET /bot/{bot_id}
```

#### Get Trading Signals
```http
GET /bot/{bot_id}/signals
```
**Key Response Fields:**
- `signals[]` - Array of active trading signals
- `symbol` - Trading pair (e.g., "BTCUSDT")
- `prediction` - "bullish", "bearish", or "neutral"
- `confidence` - Percentage (0-100)
- `entry_price` - Suggested entry price
- `stop_loss` - Risk management stop loss
- `take_profit` - Target profit level

### Data & Analytics

#### Manual Data Ingestion
```http
POST /ingest/manual
Content-Type: application/json

{
  "bot_id": "bot_123456789"
}
```

#### Get Predictions History
```http
GET /predictions/{bot_id}/{symbol}
```

#### Get Trade History
```http
GET /trades/{bot_id}
```

## 🔌 WebSocket Real-Time Updates

### Connection
```javascript
const ws = new WebSocket('ws://localhost:3333/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleRealtimeUpdate(data);
};
```

### Event Types

#### Ingestion Events
```json
{
  "type": "ingestion",
  "bot_id": "bot_123456789",
  "timestamp": "2025-09-14T18:45:00Z",
  "news_count": 12,
  "metrics": {
    "active_addresses": 526152,
    "tx_count": 533108,
    "price": 65500.00
  }
}
```

#### Prediction Events
```json
{
  "type": "prediction",
  "bot_id": "bot_123456789",
  "symbol": "BTCUSDT",
  "prediction": "bullish",
  "confidence": 85,
  "timestamp": "2025-09-14T18:44:30Z"
}
```

#### Trade Events
```json
{
  "type": "trade",
  "bot_id": "bot_123456789",
  "symbol": "BTCUSDT",
  "side": "long",
  "quantity": 0.015,
  "price": 65500.00,
  "status": "filled",
  "timestamp": "2025-09-14T18:40:00Z"
}
```

## 🎨 Frontend Component Suggestions

### 1. Signal Dashboard
```javascript
// Real-time signals display
const SignalCard = ({ signal }) => (
  <div className={`signal-card ${signal.prediction}`}>
    <h3>{signal.symbol}</h3>
    <div className="prediction">
      <span className="direction">{signal.direction}</span>
      <span className="confidence">{signal.confidence}%</span>
    </div>
    <div className="price-levels">
      <div>Entry: ${signal.entry_price}</div>
      <div>Stop Loss: ${signal.stop_loss}</div>
      <div>Take Profit: ${signal.take_profit}</div>
    </div>
    <div className="risk-info">
      <div>Size: {signal.position_size}</div>
      <div>Risk: ${signal.risk_amount}</div>
      <div>R:R {signal.risk_reward_ratio}</div>
    </div>
  </div>
);
```

### 2. Bot Status Monitor
```javascript
const BotStatus = ({ bot }) => (
  <div className="bot-status">
    <h2>{bot.name}</h2>
    <div className="metrics">
      <div>Status: <span className={bot.status}>{bot.status}</span></div>
      <div>Total Trades: {bot.total_trades}</div>
      <div>Success Rate: {bot.success_rate}%</div>
      <div>Last Activity: {formatTime(bot.last_activity)}</div>
    </div>
  </div>
);
```

### 3. Real-Time Chart Integration
```javascript
// Using Chart.js or similar
const PriceChart = ({ symbol, signals }) => {
  const addSignalToChart = (signal) => {
    // Add signal markers to price chart
    chart.data.annotations.push({
      type: 'point',
      x: signal.timestamp,
      y: signal.entry_price,
      backgroundColor: signal.prediction === 'bullish' ? 'green' : 'red',
      borderColor: 'white',
      label: `${signal.prediction} (${signal.confidence}%)`
    });
    chart.update();
  };
};
```

### 4. Trade History Table
```javascript
const TradeHistory = ({ trades }) => (
  <table className="trade-history">
    <thead>
      <tr>
        <th>Time</th>
        <th>Symbol</th>
        <th>Side</th>
        <th>Quantity</th>
        <th>Price</th>
        <th>PnL</th>
        <th>Status</th>
      </tr>
    </thead>
    <tbody>
      {trades.map(trade => (
        <tr key={trade.id} className={trade.pnl >= 0 ? 'profit' : 'loss'}>
          <td>{formatTime(trade.timestamp)}</td>
          <td>{trade.symbol}</td>
          <td className={trade.side}>{trade.side}</td>
          <td>{trade.quantity}</td>
          <td>${trade.price}</td>
          <td className={trade.pnl >= 0 ? 'positive' : 'negative'}>
            ${trade.pnl} ({trade.pnl_percent}%)
          </td>
          <td>{trade.status}</td>
        </tr>
      ))}
    </tbody>
  </table>
);
```

## 📱 State Management

### React Example with Context
```javascript
const SignalsContext = createContext();

const SignalsProvider = ({ children }) => {
  const [bots, setBots] = useState([]);
  const [signals, setSignals] = useState([]);
  const [trades, setTrades] = useState([]);
  const [wsConnected, setWsConnected] = useState(false);

  useEffect(() => {
    const ws = new WebSocket(WS_URL);
    
    ws.onopen = () => setWsConnected(true);
    ws.onclose = () => setWsConnected(false);
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      switch(data.type) {
        case 'prediction':
          updateSignals(data);
          break;
        case 'trade':
          addTrade(data);
          break;
        case 'ingestion':
          updateIngestionStatus(data);
          break;
      }
    };
    
    return () => ws.close();
  }, []);

  return (
    <SignalsContext.Provider value={{
      bots, signals, trades, wsConnected,
      setBots, setSignals, setTrades
    }}>
      {children}
    </SignalsContext.Provider>
  );
};
```

## 🎯 Key Features to Implement

### 1. Real-Time Signal Updates
- WebSocket connection for live updates
- Signal cards with color-coded predictions
- Confidence level indicators
- Entry/exit price levels

### 2. Trading Dashboard
- Active signals overview
- Bot performance metrics
- Real-time PnL tracking
- Risk management display

### 3. Historical Analysis
- Prediction accuracy over time
- Trade performance analytics
- Success rate by symbol
- Risk-adjusted returns

### 4. Bot Management
- Create/configure trading bots
- Adjust risk parameters
- Enable/disable bots
- Monitor bot status

## 🔧 Error Handling

```javascript
const apiCall = async (endpoint, options = {}) => {
  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      },
      ...options
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'API call failed');
    }
    
    return await response.json();
  } catch (error) {
    console.error('API Error:', error);
    // Handle error appropriately
    throw error;
  }
};
```

## 🎨 CSS Styling Suggestions

```css
/* Signal Cards */
.signal-card {
  border-radius: 8px;
  padding: 16px;
  margin: 8px;
  border-left: 4px solid;
}

.signal-card.bullish {
  border-left-color: #10b981;
  background: #ecfdf5;
}

.signal-card.bearish {
  border-left-color: #ef4444;
  background: #fef2f2;
}

.signal-card.neutral {
  border-left-color: #6b7280;
  background: #f9fafb;
}

/* Confidence Indicator */
.confidence {
  font-weight: bold;
  padding: 4px 8px;
  border-radius: 4px;
  background: rgba(0,0,0,0.1);
}

/* Trade Status */
.trade-history .profit {
  background-color: #f0fdf4;
}

.trade-history .loss {
  background-color: #fef2f2;
}

.positive {
  color: #059669;
}

.negative {
  color: #dc2626;
}
```

## 🚀 Deployment Notes

- API runs on port 3333 by default
- WebSocket uses same port with `/ws` path
- No CORS restrictions in development mode
- Ensure TiDB cluster is running via docker compose

## 📊 Sample Data Flow

1. **Initial Load**: Fetch bots and current signals
2. **WebSocket Connect**: Establish real-time connection
3. **Live Updates**: Handle prediction and trade events
4. **User Actions**: Create bots, trigger manual ingestion
5. **Historical Data**: Load trade history and analytics

This specification provides everything needed to build a comprehensive frontend for the crypto signals trading bot!
