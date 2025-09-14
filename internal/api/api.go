package api

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"github.com/adeilh/agentic_go_signals/internal/db"
)

type App struct {
	db  *db.DB
	app *fiber.App
	hub *Hub
}

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
			log.Printf("WebSocket client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Printf("WebSocket client disconnected. Total: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("WebSocket write error: %v", err)
					h.unregister <- client
				}
			}
		}
	}
}

func New(database *db.DB) *App {
	app := fiber.New(fiber.Config{
		AppName:      "SigForge API v1.0",
		ServerHeader: "SigForge",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	hub := newHub()
	go hub.run()

	apiApp := &App{
		db:  database,
		app: app,
		hub: hub,
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

	// WebSocket
	a.app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	a.app.Get("/ws", websocket.New(a.handleWebSocket))

	// OpenAPI
	a.app.Get("/openapi.json", a.getOpenAPI)
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

func (a *App) handleWebSocket(c *websocket.Conn) {
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

func (a *App) Listen(addr string) error {
	return a.app.Listen(addr)
}
