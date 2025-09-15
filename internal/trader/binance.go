package trader

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	apiKey    string
	apiSecret string
	baseURL   string
	wsURL     string
	client    *http.Client
}

// Market Data Structures
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type Ticker24hr struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	Count              int64  `json:"count"`
}

type OrderBook struct {
	LastUpdateId int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type Kline struct {
	OpenTime                 int64  `json:"openTime"`
	Open                     string `json:"open"`
	High                     string `json:"high"`
	Low                      string `json:"low"`
	Close                    string `json:"close"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"closeTime"`
	QuoteAssetVolume         string `json:"quoteAssetVolume"`
	NumberOfTrades           int64  `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
}

type Trade struct {
	ID           int64  `json:"id"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	QuoteQty     string `json:"quoteQty"`
	Time         int64  `json:"time"`
	IsBuyerMaker bool   `json:"isBuyerMaker"`
	IsBestMatch  bool   `json:"isBestMatch"`
}

// WebSocket Event Structures
type WSTickerEvent struct {
	EventType          string `json:"e"`
	EventTime          int64  `json:"E"`
	Symbol             string `json:"s"`
	PriceChange        string `json:"p"`
	PriceChangePercent string `json:"P"`
	WeightedAvgPrice   string `json:"w"`
	FirstTradePrice    string `json:"x"`
	LastPrice          string `json:"c"`
	LastQty            string `json:"Q"`
	BestBidPrice       string `json:"b"`
	BestBidQty         string `json:"B"`
	BestAskPrice       string `json:"a"`
	BestAskQty         string `json:"A"`
	OpenPrice          string `json:"o"`
	HighPrice          string `json:"h"`
	LowPrice           string `json:"l"`
	TotalTradedVolume  string `json:"v"`
	TotalTradedQuote   string `json:"q"`
	OpenTime           int64  `json:"O"`
	CloseTime          int64  `json:"C"`
	FirstTradeId       int64  `json:"F"`
	LastTradeId        int64  `json:"L"`
	TotalTrades        int64  `json:"n"`
}

type WSTradeEvent struct {
	EventType        string `json:"e"`
	EventTime        int64  `json:"E"`
	Symbol           string `json:"s"`
	TradeId          int64  `json:"t"`
	Price            string `json:"p"`
	Quantity         string `json:"q"`
	BuyerOrderId     int64  `json:"b"`
	SellerOrderId    int64  `json:"a"`
	TradeTime        int64  `json:"T"`
	IsBuyerMaker     bool   `json:"m"`
	IsBestPriceMatch bool   `json:"M"`
}

type WSDepthEvent struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	Symbol        string     `json:"s"`
	FirstUpdateId int64      `json:"U"`
	FinalUpdateId int64      `json:"u"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

type WSKlineEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		StartTime           int64  `json:"t"`
		EndTime             int64  `json:"T"`
		Symbol              string `json:"s"`
		Interval            string `json:"i"`
		FirstTradeId        int64  `json:"f"`
		LastTradeId         int64  `json:"L"`
		Open                string `json:"o"`
		Close               string `json:"c"`
		High                string `json:"h"`
		Low                 string `json:"l"`
		Volume              string `json:"v"`
		NumberOfTrades      int64  `json:"n"`
		IsClosed            bool   `json:"x"`
		QuoteVolume         string `json:"q"`
		TakerBuyBaseVolume  string `json:"V"`
		TakerBuyQuoteVolume string `json:"Q"`
	} `json:"k"`
}

// Order structures
type Order struct {
	Symbol   string `json:"symbol"`
	OrderID  int64  `json:"orderId"`
	ClientID string `json:"clientOrderId"`
	Side     string `json:"side"`
	Type     string `json:"type"`
	Quantity string `json:"origQty"`
	Price    string `json:"price"`
	Status   string `json:"status"`
	Time     int64  `json:"transactTime"`
}

type ErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// WebSocket Handler Function Types
type TickerHandler func(WSTickerEvent)
type TradeHandler func(WSTradeEvent)
type DepthHandler func(WSDepthEvent)
type KlineHandler func(WSKlineEvent)

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   "https://testnet.binance.vision",
		wsURL:     "wss://testnet.binance.vision/ws",
		client:    &http.Client{Timeout: 120 * time.Second},
	}
}

// NewProductionClient creates a client for Binance production environment
func NewProductionClient(apiKey, apiSecret string) *Client {
	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   "https://api.binance.com",
		wsURL:     "wss://stream.binance.com:9443",
		client:    &http.Client{Timeout: 120 * time.Second},
	}
}

// NewClientWithConfig creates a client with custom configuration
func NewClientWithConfig(apiKey, apiSecret string, isProduction bool) *Client {
	if isProduction {
		return NewProductionClient(apiKey, apiSecret)
	}
	return NewClient(apiKey, apiSecret)
}

func (c *Client) sign(query string) string {
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(query))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) TestConnection() error {
	url := c.baseURL + "/api/v3/time"

	resp, err := c.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) PlaceOrder(symbol, side string, qty float64) (Order, error) {
	endpoint := "/api/v3/order"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", strings.ToUpper(side))
	params.Set("type", "MARKET")
	params.Set("quantity", strconv.FormatFloat(qty, 'f', 8, 64))
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	query := params.Encode()
	signature := c.sign(query)
	query += "&signature=" + signature

	req, err := http.NewRequest("POST", c.baseURL+endpoint+"?"+query, nil)
	if err != nil {
		return Order{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return Order{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return Order{}, fmt.Errorf("API error: %d - %s", errResp.Code, errResp.Msg)
	}

	var order Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return Order{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return order, nil
}

func (c *Client) GetAccountInfo() (map[string]interface{}, error) {
	endpoint := "/api/v3/account"

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	query := params.Encode()
	signature := c.sign(query)
	query += "&signature=" + signature

	req, err := http.NewRequest("GET", c.baseURL+endpoint+"?"+query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// Market Data Methods

// GetTickerPrice gets the latest price for a symbol
func (c *Client) GetTickerPrice(symbol string) (TickerPrice, error) {
	endpoint := "/api/v3/ticker/price"

	var url string
	if symbol != "" {
		url = fmt.Sprintf("%s%s?symbol=%s", c.baseURL, endpoint, symbol)
	} else {
		url = c.baseURL + endpoint
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to get ticker price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return TickerPrice{}, fmt.Errorf("API error: %d - %s", errResp.Code, errResp.Msg)
	}

	var ticker TickerPrice
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return TickerPrice{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return ticker, nil
}

// GetTicker24hr gets 24hr ticker price change statistics
func (c *Client) GetTicker24hr(symbol string) ([]Ticker24hr, error) {
	endpoint := "/api/v3/ticker/24hr"

	var url string
	if symbol != "" {
		url = fmt.Sprintf("%s%s?symbol=%s", c.baseURL, endpoint, symbol)
	} else {
		url = c.baseURL + endpoint
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get 24hr ticker: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API error: %d - %s", errResp.Code, errResp.Msg)
	}

	if symbol != "" {
		var ticker Ticker24hr
		if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return []Ticker24hr{ticker}, nil
	}

	var tickers []Ticker24hr
	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tickers, nil
}

// GetOrderBook gets the order book for a symbol
func (c *Client) GetOrderBook(symbol string, limit int) (OrderBook, error) {
	endpoint := "/api/v3/depth"

	url := fmt.Sprintf("%s%s?symbol=%s", c.baseURL, endpoint, symbol)
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return OrderBook{}, fmt.Errorf("failed to get order book: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return OrderBook{}, fmt.Errorf("API error: %d - %s", errResp.Code, errResp.Msg)
	}

	var orderBook OrderBook
	if err := json.NewDecoder(resp.Body).Decode(&orderBook); err != nil {
		return OrderBook{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return orderBook, nil
}

// GetKlines gets kline/candlestick bars for a symbol
func (c *Client) GetKlines(symbol, interval string, limit int, startTime, endTime int64) ([][]interface{}, error) {
	endpoint := "/api/v3/klines"

	url := fmt.Sprintf("%s%s?symbol=%s&interval=%s", c.baseURL, endpoint, symbol, interval)

	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}
	if startTime > 0 {
		url += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API error: %d - %s", errResp.Code, errResp.Msg)
	}

	var klines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return klines, nil
}

// GetRecentTrades gets recent trades for a symbol
func (c *Client) GetRecentTrades(symbol string, limit int) ([]Trade, error) {
	endpoint := "/api/v3/trades"

	url := fmt.Sprintf("%s%s?symbol=%s", c.baseURL, endpoint, symbol)
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent trades: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API error: %d - %s", errResp.Code, errResp.Msg)
	}

	var trades []Trade
	if err := json.NewDecoder(resp.Body).Decode(&trades); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return trades, nil
}

// GetExchangeInfo gets current exchange trading rules and symbol information
func (c *Client) GetExchangeInfo() (map[string]interface{}, error) {
	endpoint := "/api/v3/exchangeInfo"

	resp, err := c.client.Get(c.baseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API error: %d - %s", errResp.Code, errResp.Msg)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return info, nil
}

// WebSocket Methods for Real-time Data

// SubscribeTickerStream subscribes to 24hr ticker statistics stream
func (c *Client) SubscribeTickerStream(ctx context.Context, symbol string, handler TickerHandler) error {
	streamName := fmt.Sprintf("%s@ticker", strings.ToLower(symbol))
	return c.subscribeWebSocket(ctx, streamName, func(data []byte) {
		var event WSTickerEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("Error unmarshaling ticker event: %v", err)
			return
		}
		handler(event)
	})
}

// SubscribeTradeStream subscribes to trade stream
func (c *Client) SubscribeTradeStream(ctx context.Context, symbol string, handler TradeHandler) error {
	streamName := fmt.Sprintf("%s@trade", strings.ToLower(symbol))
	return c.subscribeWebSocket(ctx, streamName, func(data []byte) {
		var event WSTradeEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("Error unmarshaling trade event: %v", err)
			return
		}
		handler(event)
	})
}

// SubscribeDepthStream subscribes to partial book depth stream
func (c *Client) SubscribeDepthStream(ctx context.Context, symbol string, levels int, handler DepthHandler) error {
	var streamName string
	if levels > 0 {
		streamName = fmt.Sprintf("%s@depth%d", strings.ToLower(symbol), levels)
	} else {
		streamName = fmt.Sprintf("%s@depth", strings.ToLower(symbol))
	}

	return c.subscribeWebSocket(ctx, streamName, func(data []byte) {
		var event WSDepthEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("Error unmarshaling depth event: %v", err)
			return
		}
		handler(event)
	})
}

// SubscribeKlineStream subscribes to kline/candlestick stream
func (c *Client) SubscribeKlineStream(ctx context.Context, symbol, interval string, handler KlineHandler) error {
	streamName := fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval)
	return c.subscribeWebSocket(ctx, streamName, func(data []byte) {
		var event WSKlineEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("Error unmarshaling kline event: %v", err)
			return
		}
		handler(event)
	})
}

// SubscribeAllTickerStream subscribes to all symbols ticker stream
func (c *Client) SubscribeAllTickerStream(ctx context.Context, handler func([]WSTickerEvent)) error {
	streamName := "!ticker@arr"
	return c.subscribeWebSocket(ctx, streamName, func(data []byte) {
		var events []WSTickerEvent
		if err := json.Unmarshal(data, &events); err != nil {
			log.Printf("Error unmarshaling all ticker events: %v", err)
			return
		}
		handler(events)
	})
}

// SubscribeMultiStream subscribes to multiple streams at once
func (c *Client) SubscribeMultiStream(ctx context.Context, streams []string, handler func([]byte)) error {
	streamStr := strings.Join(streams, "/")
	wsURL := fmt.Sprintf("%s/%s", c.wsURL, streamStr)

	return c.connectWebSocket(ctx, wsURL, handler)
}

// subscribeWebSocket is a helper method for single stream subscriptions
func (c *Client) subscribeWebSocket(ctx context.Context, stream string, handler func([]byte)) error {
	wsURL := fmt.Sprintf("%s/%s", c.wsURL, stream)
	return c.connectWebSocket(ctx, wsURL, handler)
}

// connectWebSocket establishes WebSocket connection and handles messages
func (c *Client) connectWebSocket(ctx context.Context, wsURL string, handler func([]byte)) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}
	defer conn.Close()

	log.Printf("Connected to WebSocket: %s", wsURL)

	// Handle WebSocket messages in a goroutine
	go func() {
		defer conn.Close()
		for {
			select {
			case <-ctx.Done():
				log.Println("WebSocket context cancelled")
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("WebSocket read error: %v", err)
					return
				}

				// Call the handler with the raw message
				handler(message)
			}
		}
	}()

	// Keep connection alive until context is cancelled
	<-ctx.Done()
	return nil
}

// WebSocket Hub for Broadcasting to Frontend

type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
}

type WSClient struct {
	hub  *WSHub
	conn *websocket.Conn
	send chan []byte
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *WSHub) Broadcast(messageType string, data interface{}) {
	message := WSMessage{
		Type: messageType,
		Data: data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	h.broadcast <- jsonData
}

func (h *WSHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WSClient{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
