package main

import (
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"ocean-strategy/internal/handlers"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Ocean Strategy Game Server",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Use(logger.New())

	api := app.Group("/api/v1")

	api.Get("/games", handlers.ListGames)
	api.Post("/games", handlers.CreateGame)
	api.Get("/games/:id", handlers.GetGame)
	api.Post("/games/:id/join", handlers.JoinGame)
	api.Post("/games/:id/start", handlers.StartGame)
	api.Post("/games/:id/next-phase", handlers.NextPhase)

	api.Post("/games/:id/facilities", handlers.BuildFacility)
	api.Post("/games/:id/ships", handlers.BuildShip)
	api.Post("/games/:id/ships/move", handlers.MoveShip)
	api.Post("/games/:id/ships/explore", handlers.Explore)
	api.Post("/games/:id/research", handlers.StartResearch)

	api.Get("/technologies", handlers.GetTechnologies)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:game_id/:player_id", websocket.New(handlers.WebSocketHandler))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
