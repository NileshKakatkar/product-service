FROM golang:1.23-alpine AS builder
WORKDIR /src/product-service
COPY shared/grpc/ /src/shared/grpc/
COPY product-service/go.mod ./
RUN go mod download
COPY product-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/product

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/service /app/service
USER app
EXPOSE 8082 50052
ENTRYPOINT ["/app/service"]
