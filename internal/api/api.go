package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/adeilh/agentic_go_signals/internal/db"
	"github.com/adeilh/agentic_go_signals/internal/kimi"
	"github.com/adeilh/agentic_go_signals/internal/services"
	"github.com/adeilh/agentic_go_signals/internal/trader"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

type App struct {
	db                *db.DB
	app               *fiber.App
	wsHub             *trader.WSHub
	hub               *Hub // Legacy hub for backward compatibility
	marketDataService *services.MarketDataService
	binanceClient     *trader.Client
	kimiClient        *kimi.Client
}

// Legacy Hub struct for backward compatibility with existing WebSocket implementation
type Hub struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan []byte
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan []byte),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Legacy WebSocket client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Printf("Legacy WebSocket client disconnected. Total: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("Legacy WebSocket write error: %v", err)
					h.unregister <- client
				}
			}
		}
	}
}

func New(database *db.DB, binanceClient *trader.Client, kimiClient *kimi.Client) *App {
	app := fiber.New(fiber.Config{
		AppName:      "SigForge API v1.0",
		ServerHeader: "SigForge",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Create WebSocket hub for market data
	wsHub := trader.NewWSHub()
	go wsHub.Run()

	// Create market data service
	marketDataService := services.NewMarketDataService(binanceClient, wsHub, database)

	// Legacy hub for backward compatibility
	legacyHub := newHub()
	go legacyHub.run()

	apiApp := &App{
		db:                database,
		app:               app,
		wsHub:             wsHub,
		hub:               legacyHub,
		marketDataService: marketDataService,
		binanceClient:     binanceClient,
		kimiClient:        kimiClient,
	}

	apiApp.setupRoutes()
	return apiApp
}

func (a *App) setupRoutes() {
	// Health check
	a.app.Get("/healthz", a.healthCheck)

	// Bot management
	a.app.Post("/bot/create", a.createBot)
	a.app.Get("/bot/:botId", a.getBotInfo)

	// Ingestion
	a.app.Post("/ingest/manual", a.manualIngest)

	// Signals
	a.app.Get("/signals/current", a.getCurrentSignal)
	a.app.Get("/signals/history", a.getSignalHistory)

	// Predictions
	a.app.Get("/predictions/latest", a.getLatestPrediction)
	a.app.Get("/predictions/history", a.getPredictionHistory)

	// Trades
	a.app.Get("/trades/latest", a.getLatestTrades)

	// Market Data Endpoints
	a.app.Get("/market/prices", a.getAllPrices)
	a.app.Get("/market/prices/:symbol", a.getSymbolPrice)
	a.app.Get("/market/ticker/:symbol", a.getSymbolTicker)
	a.app.Get("/market/orderbook/:symbol", a.getOrderBook)
	a.app.Get("/market/trades/:symbol", a.getRecentTrades)
	a.app.Get("/market/klines/:symbol", a.getKlines)
	a.app.Get("/market/summary", a.getMarketSummary)

	// Market Data Service Control
	a.app.Post("/market/start", a.startMarketData)
	a.app.Post("/market/stop", a.stopMarketData)
	a.app.Get("/market/status", a.getMarketDataStatus)
	a.app.Post("/market/symbols", a.addSymbol)

	// TiDB-backed market data endpoints
	a.app.Get("/market/tidb/prices", a.getTiDBPrices)
	a.app.Get("/market/tidb/signals/:symbol", a.getTradingSignals)
	a.app.Get("/market/tidb/history/:symbol", a.getPriceHistory)
	a.app.Get("/market/tidb/volume/:symbol", a.getVolumeAnalysis)

	// Kimi AI signals endpoint
	a.app.Get("/kimi/signals/:symbol", a.getKimiSignals)
	a.app.Get("/kimi/enhanced/:symbol", a.getEnhancedKimiSignals)

	// Advanced TiDB Analytics endpoints
	a.app.Get("/tidb/advanced/:symbol", a.getAdvancedAnalytics)
	a.app.Get("/tidb/realtime/:symbol", a.getRealTimeState)

	// Analytics endpoints for frontend (batch operations)
	a.app.Get("/analytics/advanced", a.getAdvancedAnalyticsBatch)
	a.app.Get("/analytics/realtime", a.getRealTimeStateBatch)

	// WebSocket for real-time market data
	a.app.Get("/ws/market", websocket.New(a.handleMarketDataWebSocket))

	// Legacy WebSocket (backward compatibility)
	a.app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	a.app.Get("/ws", websocket.New(a.handleLegacyWebSocket))

	// OpenAPI
	a.app.Get("/openapi.json", a.getOpenAPI)

	// Static files
	a.app.Static("/", "./web")
}

func (a *App) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "sigforge",
	})
}

func (a *App) createBot(c *fiber.Ctx) error {
	botID := "bot_" + strconv.FormatInt(time.Now().UnixNano(), 36)

	return c.JSON(fiber.Map{
		"bot_id":     botID,
		"created_at": time.Now().Unix(),
		"status":     "active",
	})
}

func (a *App) getBotInfo(c *fiber.Ctx) error {
	botID := c.Params("botId")

	return c.JSON(fiber.Map{
		"bot_id": botID,
		"status": "active",
		"trades": 0,
		"pnl":    0.0,
	})
}

func (a *App) manualIngest(c *fiber.Ctx) error {
	botID := c.Query("bot_id", "default")

	// Trigger ingestion - this would normally call the worker
	go func() {
		// Simulate processing
		time.Sleep(2 * time.Second)

		// Broadcast to WebSocket clients
		message := []byte(`{"type":"ingestion","bot_id":"` + botID + `","status":"completed","timestamp":` + strconv.FormatInt(time.Now().Unix(), 10) + `}`)
		a.hub.broadcast <- message
	}()

	return c.JSON(fiber.Map{
		"status": "triggered",
		"bot_id": botID,
	})
}

func (a *App) getCurrentSignal(c *fiber.Ctx) error {
	botID := c.Query("bot_id", "default")

	return c.JSON(fiber.Map{
		"bot_id":    botID,
		"symbol":    "BTC",
		"side":      "LONG",
		"conv":      85,
		"timestamp": time.Now().Unix(),
		"logic":     "Strong bullish momentum with high volume",
	})
}

func (a *App) getSignalHistory(c *fiber.Ctx) error {
	botID := c.Query("bot_id", "default")

	return c.JSON(fiber.Map{
		"bot_id": botID,
		"signals": []fiber.Map{
			{
				"symbol":    "BTC",
				"side":      "LONG",
				"conv":      85,
				"timestamp": time.Now().Unix() - 3600,
			},
			{
				"symbol":    "ETH",
				"side":      "SHORT",
				"conv":      72,
				"timestamp": time.Now().Unix() - 7200,
			},
		},
	})
}

func (a *App) getLatestPrediction(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"symbol":    "BTC",
		"dir":       "LONG",
		"conv":      85,
		"logic":     "Strong institutional adoption and technical breakout",
		"timestamp": time.Now().Unix(),
	})
}

func (a *App) getPredictionHistory(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"predictions": []fiber.Map{
			{
				"symbol":    "BTC",
				"dir":       "LONG",
				"conv":      85,
				"timestamp": time.Now().Unix() - 3600,
			},
		},
	})
}

func (a *App) getLatestTrades(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"trades": []fiber.Map{
			{
				"symbol":    "BTCUSDT",
				"side":      "BUY",
				"qty":       0.001,
				"price":     45000.00,
				"status":    "FILLED",
				"timestamp": time.Now().Unix(),
			},
		},
	})
}

func (a *App) handleLegacyWebSocket(c *websocket.Conn) {
	defer c.Close()

	a.hub.register <- c

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			a.hub.unregister <- c
			break
		}
	}
}

// Market Data Handler Methods

func (a *App) getAllPrices(c *fiber.Ctx) error {
	prices := a.marketDataService.GetAllPrices()
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   prices,
		"count":  len(prices),
	})
}

func (a *App) getSymbolPrice(c *fiber.Ctx) error {
	symbol := c.Params("symbol")

	price, exists := a.marketDataService.GetPriceData(symbol)
	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"status": "error",
			"error":  "Symbol not found",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   price,
	})
}

func (a *App) getSymbolTicker(c *fiber.Ctx) error {
	symbol := c.Params("symbol")

	tickers, err := a.binanceClient.GetTicker24hr(symbol)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	if len(tickers) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"status": "error",
			"error":  "Symbol not found",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   tickers[0],
	})
}

func (a *App) getOrderBook(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	limit := c.QueryInt("limit", 100)

	orderBook, err := a.binanceClient.GetOrderBook(symbol, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   orderBook,
	})
}

func (a *App) getRecentTrades(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	limit := c.QueryInt("limit", 100)

	trades, err := a.binanceClient.GetRecentTrades(symbol, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   trades,
		"count":  len(trades),
	})
}

func (a *App) getKlines(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	interval := c.Query("interval", "1h")
	limit := c.QueryInt("limit", 100)
	startTime := c.QueryInt("startTime", 0)
	endTime := c.QueryInt("endTime", 0)

	klines, err := a.binanceClient.GetKlines(symbol, interval, limit, int64(startTime), int64(endTime))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   klines,
		"count":  len(klines),
	})
}

func (a *App) getMarketSummary(c *fiber.Ctx) error {
	summary := a.marketDataService.GetMarketSummary()
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   summary,
	})
}

// Market Data Service Control Methods

func (a *App) startMarketData(c *fiber.Ctx) error {
	if a.marketDataService.IsRunning() {
		return c.JSON(fiber.Map{
			"status":  "info",
			"message": "Market data service is already running",
		})
	}

	ctx := context.Background()
	if err := a.marketDataService.StartStreaming(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Market data service started",
	})
}

func (a *App) stopMarketData(c *fiber.Ctx) error {
	if !a.marketDataService.IsRunning() {
		return c.JSON(fiber.Map{
			"status":  "info",
			"message": "Market data service is not running",
		})
	}

	a.marketDataService.Stop()

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Market data service stopped",
	})
}

func (a *App) getMarketDataStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"running":     a.marketDataService.IsRunning(),
			"price_count": len(a.marketDataService.GetAllPrices()),
			"timestamp":   time.Now().Unix(),
		},
	})
}

func (a *App) addSymbol(c *fiber.Ctx) error {
	var req struct {
		Symbol string `json:"symbol"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status": "error",
			"error":  "Invalid request body",
		})
	}

	if req.Symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"status": "error",
			"error":  "Symbol is required",
		})
	}

	a.marketDataService.AddSymbol(req.Symbol)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Symbol added to market data service",
		"symbol":  req.Symbol,
	})
}

func (a *App) getOpenAPI(c *fiber.Ctx) error {
	spec := fiber.Map{
		"openapi": "3.0.0",
		"info": fiber.Map{
			"title":       "SigForge API",
			"version":     "1.0.0",
			"description": "TiDB-powered crypto signal generation API with multi-tenant bot support",
		},
		"servers": []fiber.Map{
			{"url": "http://localhost:3333", "description": "Development server"},
		},
		"paths": fiber.Map{
			"/bot/create": fiber.Map{
				"post": fiber.Map{
					"summary": "Create a new trading bot",
					"responses": fiber.Map{
						"200": fiber.Map{"description": "Bot created successfully"},
					},
				},
			},
			"/ingest/manual": fiber.Map{
				"post": fiber.Map{
					"summary": "Trigger manual data ingestion",
					"parameters": []fiber.Map{
						{
							"name":   "bot_id",
							"in":     "query",
							"schema": fiber.Map{"type": "string"},
						},
					},
				},
			},
			"/signals/current": fiber.Map{
				"get": fiber.Map{
					"summary": "Get current trading signal",
				},
			},
			"/ws": fiber.Map{
				"get": fiber.Map{
					"summary": "WebSocket endpoint for real-time updates",
				},
			},
		},
	}

	return c.JSON(spec)
}

// TiDB-backed market data endpoints

func (a *App) getTiDBPrices(c *fiber.Ctx) error {
	data := a.marketDataService.GetMarketDataFromTiDB()
	if data == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch market data from TiDB",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func (a *App) getTradingSignals(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	signals := a.marketDataService.GetTradingSignalsFromTiDB(symbol)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    signals,
	})
}

func (a *App) getPriceHistory(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	// Create a temporary store to access the data
	store := db.NewMarketDataStore(a.db)
	history, err := store.GetPriceHistory(symbol, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch price history",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"symbol":  symbol,
			"history": history,
			"count":   len(history),
		},
	})
}

func (a *App) getVolumeAnalysis(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	hoursStr := c.Query("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		hours = 24
	}

	// Create a temporary store to access the data
	store := db.NewMarketDataStore(a.db)
	volumeMetrics, err := store.GetTradingVolume(symbol, hours)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch volume analysis",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"symbol":  symbol,
			"hours":   hours,
			"metrics": volumeMetrics,
		},
	})
}

// Market Data WebSocket handler
func (a *App) handleMarketDataWebSocket(c *websocket.Conn) {
	log.Println("Market Data WebSocket client connected")
	defer func() {
		c.Close()
		log.Println("Market Data WebSocket client disconnected")
	}()

	// Send initial market data
	if a.marketDataService.IsRunning() {
		initialData := a.marketDataService.GetMarketDataFromTiDB()
		if initialData != nil {
			if err := c.WriteJSON(fiber.Map{
				"type": "initial_data",
				"data": initialData,
			}); err != nil {
				log.Printf("Error sending initial market data: %v", err)
				return
			}
		}
	}

	// Listen for client messages and handle ping/pong
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Market Data WebSocket error: %v", err)
			}
			break
		}

		// Handle ping messages
		if messageType == websocket.PingMessage {
			if err := c.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Printf("Error sending pong: %v", err)
				break
			}
			continue
		}

		// Handle client requests for specific data
		if messageType == websocket.TextMessage {
			var request map[string]interface{}
			if err := json.Unmarshal(message, &request); err == nil {
				if requestType, ok := request["type"].(string); ok {
					switch requestType {
					case "get_signals":
						if symbol, ok := request["symbol"].(string); ok {
							signals := a.marketDataService.GetTradingSignalsFromTiDB(symbol)
							if err := c.WriteJSON(fiber.Map{
								"type":   "signals",
								"symbol": symbol,
								"data":   signals,
							}); err != nil {
								log.Printf("Error sending signals: %v", err)
								return
							}
						}
					case "get_prices":
						data := a.marketDataService.GetMarketDataFromTiDB()
						if err := c.WriteJSON(fiber.Map{
							"type": "prices",
							"data": data,
						}); err != nil {
							log.Printf("Error sending prices: %v", err)
							return
						}
					case "subscribe_symbol":
						if symbol, ok := request["symbol"].(string); ok {
							// Send real-time updates for this symbol
							go func() {
								ticker := time.NewTicker(5 * time.Second)
								defer ticker.Stop()

								for range ticker.C {
									signals := a.marketDataService.GetTradingSignalsFromTiDB(symbol)
									if err := c.WriteJSON(fiber.Map{
										"type":   "live_update",
										"symbol": symbol,
										"data":   signals,
									}); err != nil {
										return // Client disconnected
									}
								}
							}()
						}
					}
				}
			}
		}
	}
}

// getKimiSignals handles requests for enhanced Kimi AI trading signals using TiDB analytics
func (a *App) getKimiSignals(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"status": "error",
			"error":  "symbol parameter is required",
		})
	}

	// Get advanced TiDB analytics for comprehensive analysis
	advancedSignals, err := a.marketDataService.GetAdvancedTiDBSignals(symbol)
	if err != nil {
		log.Printf("Error getting advanced TiDB signals for %s: %v", symbol, err)
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  "Failed to get market analytics",
		})
	}

	// Get real-time market state
	realTimeState, err := a.marketDataService.GetRealTimeMarketState(symbol)
	if err != nil {
		log.Printf("Error getting real-time state for %s: %v", symbol, err)
		realTimeState = make(map[string]interface{})
	}

	// Create enhanced prompt for Kimi AI with TiDB analytics
	prompt := fmt.Sprintf(`REAL-TIME CRYPTO TRADING ANALYSIS FOR %s

TECHNICAL INDICATORS:
- Current Price: $%.2f
- SMA10: $%.2f | SMA20: $%.2f 
- Price vs SMA10: %.2f%% | Price vs SMA20: %.2f%%
- Golden Cross: %t (SMA10 > SMA20)
- Volatility: %.4f

MOMENTUM ANALYSIS:
- 1min: %.3f%% | 5min: %.3f%% | 15min: %.3f%%
- Price Trend: %s

VOLUME DYNAMICS:
- Buy/Sell Ratio: %.2f (>0.6=Bullish, <0.4=Bearish)
- Volume Surge: %.2fx vs previous period
- Trade Frequency: %.0f trades/30min
- Buy Pressure: %.2f%%

SUPPORT/RESISTANCE:
- Support: $%.2f (%.2f%% away)
- Resistance: $%.2f (%.2f%% away)
- Risk Zone: %s

REAL-TIME SIGNALS:
- Volume Spike: %t
- Volatility Spike: %.2fx
- Order Flow: %s

Provide a JSON response with:
{
  "action": "BUY|SELL|HOLD",
  "confidence": 1-100,
  "timeframe": "1-5min|5-15min|15-60min",
  "entry_price": suggested_price,
  "stop_loss": suggested_stop,
  "take_profit": suggested_target,
  "reasoning": "detailed analysis",
  "risk_level": "LOW|MEDIUM|HIGH",
  "signals": ["signal1", "signal2", "signal3"]
}`,
		symbol,
		getFloat(advancedSignals, "current_price"),
		getFloat(advancedSignals, "sma_10"),
		getFloat(advancedSignals, "sma_20"),
		getFloat(advancedSignals, "price_vs_sma10"),
		getFloat(advancedSignals, "price_vs_sma20"),
		getBool(advancedSignals, "sma_cross"),
		getFloat(advancedSignals, "volatility"),
		getFloat(advancedSignals, "momentum_1min"),
		getFloat(advancedSignals, "momentum_5min"),
		getFloat(advancedSignals, "momentum_15min"),
		determineTrend(advancedSignals),
		getFloat(advancedSignals, "volume_ratio"),
		getFloat(realTimeState, "volume_surge"),
		getFloat(advancedSignals, "trade_frequency"),
		getFloat(realTimeState, "buy_pressure")*100,
		getFloat(advancedSignals, "support_level"),
		getFloat(advancedSignals, "support_distance"),
		getFloat(advancedSignals, "resistance_level"),
		getFloat(advancedSignals, "resistance_distance"),
		determineRiskZone(advancedSignals),
		getFloat(realTimeState, "volume_surge") > 1.5,
		getFloat(realTimeState, "volatility_spike"),
		determineOrderFlow(realTimeState),
	)

	// Call Kimi AI with enhanced analytics
	ctx := context.Background()
	prediction, err := a.kimiClient.Ask(ctx,
		"You are an expert cryptocurrency trader with access to real-time TiDB analytics. Analyze the comprehensive market data and provide precise, actionable trading recommendations with specific entry/exit points.",
		prompt)
	if err != nil {
		log.Printf("Error getting Kimi AI response for %s: %v", symbol, err)
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  "Failed to get AI analysis",
		})
	}

	// Parse Kimi response for enhanced data
	enhancedData := parseKimiResponse(prediction, advancedSignals, realTimeState)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"symbol":         symbol,
			"recommendation": prediction.Dir,
			"confidence":     prediction.Conv,
			"reasoning":      prediction.Logic,
			"enhanced_data":  enhancedData,
			"tidb_analytics": advancedSignals,
			"realtime_state": realTimeState,
			"timestamp":      time.Now(),
			"source":         "Kimi AI + TiDB Analytics",
		},
	})
}

func (a *App) Listen(addr string) error {
	return a.app.Listen(addr)
}

// Helper functions for enhanced Kimi AI analysis
func getFloat(data map[string]interface{}, key string) float64 {
	if val, exists := data[key]; exists {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0.0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, exists := data[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func determineTrend(signals map[string]interface{}) string {
	momentum1 := getFloat(signals, "momentum_1min")
	momentum5 := getFloat(signals, "momentum_5min")
	smaCross := getBool(signals, "sma_cross")

	if momentum1 > 0.5 && momentum5 > 1.0 && smaCross {
		return "STRONG_BULLISH"
	} else if momentum1 > 0.2 && momentum5 > 0.5 {
		return "BULLISH"
	} else if momentum1 < -0.5 && momentum5 < -1.0 && !smaCross {
		return "STRONG_BEARISH"
	} else if momentum1 < -0.2 && momentum5 < -0.5 {
		return "BEARISH"
	}
	return "SIDEWAYS"
}

func determineRiskZone(signals map[string]interface{}) string {
	supportDist := getFloat(signals, "support_distance")
	resistanceDist := getFloat(signals, "resistance_distance")
	volatility := getFloat(signals, "volatility")

	if supportDist < 2.0 || resistanceDist < 2.0 {
		return "HIGH_RISK" // Near support/resistance
	} else if volatility > 0.05 {
		return "HIGH_VOLATILITY"
	} else if supportDist > 5.0 && resistanceDist > 5.0 {
		return "SAFE_ZONE"
	}
	return "MODERATE"
}

func determineOrderFlow(realTimeState map[string]interface{}) string {
	buyPressure := getFloat(realTimeState, "buy_pressure")
	volumeSurge := getFloat(realTimeState, "volume_surge")

	if buyPressure > 0.7 && volumeSurge > 1.5 {
		return "STRONG_BUY_FLOW"
	} else if buyPressure > 0.6 {
		return "BUY_FLOW"
	} else if buyPressure < 0.3 && volumeSurge > 1.5 {
		return "STRONG_SELL_FLOW"
	} else if buyPressure < 0.4 {
		return "SELL_FLOW"
	}
	return "BALANCED"
}

func parseKimiResponse(prediction kimi.Prediction, signals, realTimeState map[string]interface{}) map[string]interface{} {
	enhanced := make(map[string]interface{})

	// Calculate risk-adjusted position size
	volatility := getFloat(signals, "volatility")
	if volatility > 0 {
		enhanced["position_size_multiplier"] = 1.0 / (1.0 + volatility*10) // Lower size for higher volatility
	}

	// Calculate entry timing score
	momentum := getFloat(signals, "momentum_5min")
	volumeRatio := getFloat(signals, "volume_ratio")
	buyPressure := getFloat(realTimeState, "buy_pressure")

	timingScore := (momentum + (volumeRatio-0.5)*2 + (buyPressure-0.5)*2) / 3
	enhanced["entry_timing_score"] = timingScore

	// Determine urgency level
	if getFloat(realTimeState, "volume_surge") > 2.0 {
		enhanced["urgency"] = "HIGH"
	} else if getFloat(realTimeState, "volume_surge") > 1.5 {
		enhanced["urgency"] = "MEDIUM"
	} else {
		enhanced["urgency"] = "LOW"
	}

	return enhanced
}

// getEnhancedKimiSignals provides advanced Kimi AI analysis with comprehensive TiDB data
func (a *App) getEnhancedKimiSignals(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"status": "error",
			"error":  "symbol parameter is required",
		})
	}

	// Get comprehensive TiDB analytics
	advancedSignals, err := a.marketDataService.GetAdvancedTiDBSignals(symbol)
	if err != nil {
		log.Printf("Error getting advanced TiDB signals for %s: %v", symbol, err)
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  "Failed to get market analytics",
		})
	}

	// Get real-time market state
	realTimeState, err := a.marketDataService.GetRealTimeMarketState(symbol)
	if err != nil {
		log.Printf("Error getting real-time state for %s: %v", symbol, err)
		realTimeState = make(map[string]interface{})
	}

	// Enhanced prompt with more sophisticated analysis
	prompt := fmt.Sprintf(`ADVANCED CRYPTO TRADING ANALYSIS FOR %s

TECHNICAL ANALYSIS:
- Current Price: $%.2f
- SMA Cross Signal: %t (Golden: %t, Death: %t)
- Volume Ratio: %.3f (vs 20-period average)
- Momentum Indicators:
  * 5min: %.3f%% 
  * 15min: %.3f%%
  * Trend: %s

REAL-TIME MARKET STATE:
- Volume Surge: %.2fx normal levels
- Buy Pressure: %.1f%% (>60%%=Bullish, <40%%=Bearish)
- Volatility Spike: %t
- Order Flow: %s

RISK ASSESSMENT:
- Support: $%.2f (%.2f%% away)
- Resistance: $%.2f (%.2f%% away)
- Risk Zone: %s
- Volatility: %.4f

Provide a sophisticated trading recommendation with:
1. Clear BUY/SELL/HOLD decision with confidence 1-100%%
2. Specific entry price and timing
3. Stop loss and take profit levels
4. Risk assessment and position sizing
5. Key technical reasons supporting the decision

Format as JSON:
{
  "recommendation": "BUY|SELL|HOLD",
  "confidence": "XX%%",
  "reasoning": "comprehensive analysis",
  "entry_price": price,
  "stop_loss": price,
  "take_profit": price,
  "risk_level": "LOW|MEDIUM|HIGH",
  "position_size": "percentage",
  "timeframe": "minutes",
  "key_signals": ["signal1", "signal2"]
}`,
		symbol,
		getFloat(advancedSignals, "current_price"),
		getBool(advancedSignals, "sma_cross"),
		getBool(advancedSignals, "sma_cross"),
		!getBool(advancedSignals, "sma_cross"),
		getFloat(advancedSignals, "volume_ratio"),
		getFloat(advancedSignals, "momentum_5min"),
		getFloat(advancedSignals, "momentum_15min"),
		determineTrend(advancedSignals),
		getFloat(realTimeState, "volume_surge"),
		getFloat(realTimeState, "buy_pressure")*100,
		getFloat(realTimeState, "volatility_spike") > 1.5,
		determineOrderFlow(realTimeState),
		getFloat(advancedSignals, "support_level"),
		getFloat(advancedSignals, "support_distance"),
		getFloat(advancedSignals, "resistance_level"),
		getFloat(advancedSignals, "resistance_distance"),
		determineRiskZone(advancedSignals),
		getFloat(advancedSignals, "volatility"),
	)

	// Call Kimi AI with enhanced analytics
	ctx := context.Background()
	prediction, err := a.kimiClient.Ask(ctx,
		"You are a professional cryptocurrency trader and technical analyst with access to real-time TiDB market analytics. Provide precise, actionable trading recommendations with specific entry/exit points and risk management.",
		prompt)
	if err != nil {
		log.Printf("Error getting enhanced Kimi AI response for %s: %v", symbol, err)
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  "Failed to get AI analysis",
		})
	}

	// Enhanced response parsing
	enhancedData := parseKimiResponse(prediction, advancedSignals, realTimeState)

	// Additional enhanced metrics
	enhancedData["market_sentiment"] = determineTrend(advancedSignals)
	enhancedData["volatility_adjusted_confidence"] = calculateVolatilityAdjustedConfidence(float64(prediction.Conv), getFloat(advancedSignals, "volatility"))
	enhancedData["optimal_position_size"] = calculateOptimalPositionSize(advancedSignals, realTimeState)

	return c.JSON(fiber.Map{
		"success": true,
		"data": map[string]interface{}{
			"symbol":         symbol,
			"recommendation": prediction.Dir,
			"confidence":     fmt.Sprintf("%.0f%%", prediction.Conv),
			"reasoning":      prediction.Logic,
			"enhanced_data":  enhancedData,
			"tidb_analytics": advancedSignals,
			"realtime_state": realTimeState,
			"timestamp":      time.Now(),
			"source":         "Enhanced Kimi AI + Advanced TiDB Analytics",
		},
	})
}

// Helper functions for enhanced analysis
func calculateVolatilityAdjustedConfidence(baseConfidence, volatility float64) float64 {
	// Reduce confidence in high volatility environments
	volatilityFactor := 1.0 - (volatility * 2)
	if volatilityFactor < 0.5 {
		volatilityFactor = 0.5
	}
	return baseConfidence * volatilityFactor
}

func calculateOptimalPositionSize(signals, realTimeState map[string]interface{}) float64 {
	// Base position size of 10%
	baseSize := 0.10

	// Adjust for volatility
	volatility := getFloat(signals, "volatility")
	volatilityAdjustment := 1.0 - (volatility * 5)
	if volatilityAdjustment < 0.2 {
		volatilityAdjustment = 0.2
	}

	// Adjust for confidence and volume
	volumeRatio := getFloat(signals, "volume_ratio")
	buyPressure := getFloat(realTimeState, "buy_pressure")
	confidenceBoost := (volumeRatio + buyPressure) / 2

	return baseSize * volatilityAdjustment * (1 + confidenceBoost)
}

// getAdvancedAnalytics provides TiDB-powered advanced market analytics
func (a *App) getAdvancedAnalytics(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"status": "error",
			"error":  "symbol parameter is required",
		})
	}

	analytics, err := a.marketDataService.GetAdvancedTiDBSignals(symbol)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  fmt.Sprintf("Failed to get analytics: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":    "success",
		"data":      analytics,
		"symbol":    symbol,
		"timestamp": time.Now(),
		"source":    "TiDB Advanced Analytics",
	})
}

// getRealTimeState provides real-time market state using TiDB
func (a *App) getRealTimeState(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"status": "error",
			"error":  "symbol parameter is required",
		})
	}

	state, err := a.marketDataService.GetRealTimeMarketState(symbol)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  fmt.Sprintf("Failed to get real-time state: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":    "success",
		"data":      state,
		"symbol":    symbol,
		"timestamp": time.Now(),
		"source":    "TiDB Real-time Analytics",
	})
}

// getAdvancedAnalyticsBatch provides TiDB-powered advanced analytics for multiple symbols
func (a *App) getAdvancedAnalyticsBatch(c *fiber.Ctx) error {
	// Default symbols to analyze
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "SOLUSDT"}

	var results []map[string]interface{}

	for _, symbol := range symbols {
		analytics, err := a.marketDataService.GetAdvancedTiDBSignals(symbol)
		if err != nil {
			log.Printf("Failed to get analytics for %s: %v", symbol, err)
			continue
		}

		// Add symbol and timestamp to the analytics result
		analytics["symbol"] = symbol
		analytics["timestamp"] = time.Now()
		results = append(results, analytics)
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"signals":   results,
		"timestamp": time.Now(),
		"source":    "TiDB Advanced Analytics Batch",
	})
}

// getRealTimeStateBatch provides real-time market state for multiple symbols
func (a *App) getRealTimeStateBatch(c *fiber.Ctx) error {
	// Default symbols to analyze
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "SOLUSDT"}

	var results []map[string]interface{}

	for _, symbol := range symbols {
		state, err := a.marketDataService.GetRealTimeMarketState(symbol)
		if err != nil {
			log.Printf("Failed to get real-time state for %s: %v", symbol, err)
			continue
		}

		// Add symbol and timestamp to the state result
		state["symbol"] = symbol
		state["timestamp"] = time.Now()
		results = append(results, state)
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"market_state": results,
		"timestamp":    time.Now(),
		"source":       "TiDB Real-time Analytics Batch",
	})
}
