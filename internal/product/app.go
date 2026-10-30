package product

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/nilesh/product-service/internal/product/rpc"
	"log"
)

func NewApp(cfg Config) (*fiber.App, error) {
	db, err := openDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	app := fiber.New(fiber.Config{AppName: "ShopCore Product Service", BodyLimit: 1 << 20})
	app.Use(recover.New())
	h := &Handler{DB: db}
	go func() {
		if err := rpc.Listen(":"+cfg.GRPCPort, db); err != nil {
			log.Printf("product gRPC server stopped: %v", err)
		}
	}()
	categories := &CategoryHandler{DB: db}
	app.Get("/api/v1/categories", categories.List)
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok", "service": "product"}) })
	app.Get("/api/v1/products", h.List)
	app.Get("/api/v1/products/:id", h.Get)
	secured := app.Group("/api/v1", authenticate(cfg.PublicKeyPath))
	secured.Post("/categories", requireAdmin, categories.Create)
	secured.Post("/products", requireAdmin, h.Create)
	secured.Patch("/products/:id", requireAdmin, h.Update)
	secured.Delete("/products/:id", requireAdmin, h.Delete)
	internal := app.Group("/internal/v1", serviceAuth(cfg.InternalServiceKey))
	internal.Post("/products/:id/reserve", h.Reserve)
	internal.Post("/products/:id/release", h.Release)
	return app, nil
}
