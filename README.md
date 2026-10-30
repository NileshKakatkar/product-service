# Product Service

Fiber service for product and category management. It owns the product database and exposes internal inventory reservation and release endpoints for the Order Service.

## Run

```sh
go mod download
go run ./cmd/product
```

Default port: `8082`.
