# Product Service

Fiber service for product and category management. It owns the product database, exposes inventory reservation and release endpoints to Order, and serves product details through the shared gRPC contract on port `50052`.

## Run

```sh
go mod download
go run ./cmd/product
```

Default HTTP port: `8082`. Keep `shared/grpc` beside this service because its Go module uses the shared contracts repository through a local replace directive.
