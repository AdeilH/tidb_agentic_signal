package trader

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// BinanceWebSocketManager manages real-time Binance data streams
type BinanceWebSocketManager struct {
	client       *Client
	conn         *websocket.Conn
	symbols      []string
	isRunning    bool
	ctx          context.Context
	cancel       context.CancelFunc
	dataHandlers map[string]func(interface{})
}

// CombinedStreamResponse represents the wrapped response from combined streams
type CombinedStreamResponse struct {
	Stream string      `json:"stream"`
	Data   interface{} `json:"data"`
}

// WebSocket Event Handlers
type StreamDataHandler func(symbol, streamType string, data interface{})

// NewBinanceWebSocketManager creates a new WebSocket manager
func NewBinanceWebSocketManager(client *Client, symbols []string) *BinanceWebSocketManager {
	return &BinanceWebSocketManager{
		client:       client,
		symbols:      symbols,
		dataHandlers: make(map[string]func(interface{})),
	}
}

// SetDataHandler sets a handler for specific data types
func (wsm *BinanceWebSocketManager) SetDataHandler(dataType string, handler func(interface{})) {
	wsm.dataHandlers[dataType] = handler
}

// Start begins the WebSocket connection with combined streams
func (wsm *BinanceWebSocketManager) Start(ctx context.Context) error {
	if wsm.isRunning {
		return fmt.Errorf("WebSocket manager is already running")
	}

	wsm.ctx, wsm.cancel = context.WithCancel(ctx)
	wsm.isRunning = true

	// Build combined stream URL using production market data endpoints (real data)
	// Note: We use production WebSocket for real market data, but testnet for trading
	streams := wsm.buildCombinedStreams()
	wsURL := fmt.Sprintf("wss://stream.binance.com:9443/stream?streams=%s", strings.Join(streams, "/"))

	log.Printf("Connecting to Binance Market Data WebSocket (Production Data): %s", wsURL)

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		wsm.isRunning = false
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	wsm.conn = conn

	// Start message handling
	go wsm.handleMessages()
	go wsm.pingHandler()

	log.Printf("✅ Binance WebSocket connected successfully with %d streams", len(streams))
	return nil
}

// Stop closes the WebSocket connection
func (wsm *BinanceWebSocketManager) Stop() {
	if !wsm.isRunning {
		return
	}

	log.Println("Stopping Binance WebSocket manager...")
	wsm.isRunning = false

	if wsm.cancel != nil {
		wsm.cancel()
	}

	if wsm.conn != nil {
		wsm.conn.Close()
	}

	log.Println("✅ Binance WebSocket manager stopped")
}

// buildCombinedStreams creates the stream list for combined connection
func (wsm *BinanceWebSocketManager) buildCombinedStreams() []string {
	var streams []string

	for _, symbol := range wsm.symbols {
		symbolLower := strings.ToLower(symbol)

		// Add essential streams for market data and decision making
		streams = append(streams,
			fmt.Sprintf("%s@ticker", symbolLower),   // 24hr ticker statistics
			fmt.Sprintf("%s@trade", symbolLower),    // Individual trades
			fmt.Sprintf("%s@kline_1m", symbolLower), // 1-minute candlesticks
			fmt.Sprintf("%s@depth20", symbolLower),  // Order book top 20 levels
		)
	}

	return streams
}

// handleMessages processes incoming WebSocket messages
func (wsm *BinanceWebSocketManager) handleMessages() {
	defer wsm.conn.Close()

	for {
		select {
		case <-wsm.ctx.Done():
			return
		default:
			_, message, err := wsm.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}

			wsm.processMessage(message)
		}
	}
}

// processMessage parses and handles individual messages
func (wsm *BinanceWebSocketManager) processMessage(message []byte) {
	var response CombinedStreamResponse
	if err := json.Unmarshal(message, &response); err != nil {
		log.Printf("Error unmarshaling WebSocket message: %v", err)
		return
	}

	// Parse stream name to extract symbol and type
	parts := strings.Split(response.Stream, "@")
	if len(parts) != 2 {
		log.Printf("Invalid stream format: %s", response.Stream)
		return
	}

	symbol := strings.ToUpper(parts[0])
	streamType := parts[1]

	// Route to appropriate handler based on stream type
	switch {
	case strings.HasPrefix(streamType, "ticker"):
		wsm.handleTickerData(symbol, response.Data)
	case strings.HasPrefix(streamType, "trade"):
		wsm.handleTradeData(symbol, response.Data)
	case strings.HasPrefix(streamType, "kline"):
		wsm.handleKlineData(symbol, response.Data)
	case strings.HasPrefix(streamType, "depth"):
		wsm.handleDepthData(symbol, response.Data)
	default:
		log.Printf("Unknown stream type: %s", streamType)
	}
}

// handleTickerData processes 24hr ticker statistics
func (wsm *BinanceWebSocketManager) handleTickerData(symbol string, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling ticker data: %v", err)
		return
	}

	var tickerEvent WSTickerEvent
	if err := json.Unmarshal(dataBytes, &tickerEvent); err != nil {
		log.Printf("Error unmarshaling ticker event: %v", err)
		return
	}

	// Call registered handler
	if handler, exists := wsm.dataHandlers["ticker"]; exists {
		handler(tickerEvent)
	}

	// log.Printf("📊 %s Ticker: Price=%s, Change=%s%%",
	// 	symbol, tickerEvent.LastPrice, tickerEvent.PriceChangePercent)
}

// handleTradeData processes individual trade information
func (wsm *BinanceWebSocketManager) handleTradeData(symbol string, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling trade data: %v", err)
		return
	}

	var tradeEvent WSTradeEvent
	if err := json.Unmarshal(dataBytes, &tradeEvent); err != nil {
		log.Printf("Error unmarshaling trade event: %v", err)
		return
	}

	// Call registered handler
	if handler, exists := wsm.dataHandlers["trade"]; exists {
		handler(tradeEvent)
	}

	// log.Printf("💱 %s Trade: Price=%s, Qty=%s, Side=%s",
	// 	symbol, tradeEvent.Price, tradeEvent.Quantity,
	// 	map[bool]string{true: "SELL", false: "BUY"}[tradeEvent.IsBuyerMaker])
}

// handleKlineData processes candlestick information
func (wsm *BinanceWebSocketManager) handleKlineData(symbol string, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling kline data: %v", err)
		return
	}

	var klineEvent WSKlineEvent
	if err := json.Unmarshal(dataBytes, &klineEvent); err != nil {
		log.Printf("Error unmarshaling kline event: %v", err)
		return
	}

	// Call registered handler
	if handler, exists := wsm.dataHandlers["kline"]; exists {
		handler(klineEvent)
	}

	if klineEvent.Kline.IsClosed {
		log.Printf("📈 %s Kline Closed: O=%s, H=%s, L=%s, C=%s, V=%s",
			symbol, klineEvent.Kline.Open, klineEvent.Kline.High,
			klineEvent.Kline.Low, klineEvent.Kline.Close, klineEvent.Kline.Volume)
	}
}

// handleDepthData processes order book updates
func (wsm *BinanceWebSocketManager) handleDepthData(symbol string, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling depth data: %v", err)
		return
	}

	var depthEvent WSDepthEvent
	if err := json.Unmarshal(dataBytes, &depthEvent); err != nil {
		log.Printf("Error unmarshaling depth event: %v", err)
		return
	}

	// Call registered handler
	if handler, exists := wsm.dataHandlers["depth"]; exists {
		handler(depthEvent)
	}

	// log.Printf("📋 %s Depth: Bids=%d, Asks=%d",
	// 	symbol, len(depthEvent.Bids), len(depthEvent.Asks))
}

// pingHandler sends periodic ping messages to keep connection alive
func (wsm *BinanceWebSocketManager) pingHandler() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wsm.ctx.Done():
			return
		case <-ticker.C:
			if wsm.conn != nil {
				if err := wsm.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("Error sending ping: %v", err)
					return
				}
			}
		}
	}
}

// AddSymbol adds a new symbol to monitor (requires restart)
func (wsm *BinanceWebSocketManager) AddSymbol(symbol string) {
	for _, existing := range wsm.symbols {
		if strings.EqualFold(existing, symbol) {
			return // Already exists
		}
	}
	wsm.symbols = append(wsm.symbols, strings.ToUpper(symbol))
	log.Printf("Added symbol %s (restart required)", symbol)
}

// GetSymbols returns current monitored symbols
func (wsm *BinanceWebSocketManager) GetSymbols() []string {
	return wsm.symbols
}

// IsRunning returns the connection status
func (wsm *BinanceWebSocketManager) IsRunning() bool {
	return wsm.isRunning
}
