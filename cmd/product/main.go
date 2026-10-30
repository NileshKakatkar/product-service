package main

import (
	"log"

	"github.com/nilesh/product-service/internal/product"
)

func main() {
	cfg := product.LoadConfig()
	app, err := product.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
