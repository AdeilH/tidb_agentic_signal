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

// getKimiSignals handles requests for Kimi AI trading signals
func (a *App) getKimiSignals(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{
			"status": "error",
			"error":  "symbol parameter is required",
		})
	}

	// Get market data from TiDB for analysis
	tidbData := a.marketDataService.GetTradingSignalsFromTiDB(symbol)
	
	// Create a prompt for Kimi AI based on market data
	prompt := fmt.Sprintf(`Analyze the following market data for %s and provide a trading recommendation:

Price Data: %v

Please provide:
1. Trading recommendation (BUY/SELL/HOLD)
2. Confidence level (High/Medium/Low)
3. Brief reasoning (max 50 words)
4. Risk assessment

Respond in JSON format with fields: recommendation, confidence, reasoning, risk_level`, symbol, tidbData)

	// Call Kimi AI
	ctx := context.Background()
	prediction, err := a.kimiClient.Ask(ctx, 
		"You are a professional crypto trading analyst. Analyze market data and provide clear trading recommendations.",
		prompt)
	if err != nil {
		log.Printf("Error getting Kimi AI response for %s: %v", symbol, err)
		return c.Status(500).JSON(fiber.Map{
			"status": "error",
			"error":  "Failed to get AI analysis",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"symbol":         symbol,
			"recommendation": prediction.Dir,    // LONG, SHORT, FLAT
			"confidence":     prediction.Conv,   // 1-100
			"reasoning":      prediction.Logic,  // Reasoning
			"timestamp":      time.Now(),
			"source":         "Kimi AI",
		},
	})
}

func (a *App) Listen(addr string) error {
	return a.app.Listen(addr)
}
