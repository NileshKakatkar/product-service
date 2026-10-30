package product

import "os"

type Config struct{ Port, DatabaseURL, PublicKeyPath, InternalServiceKey string }

func LoadConfig() Config {
	get := func(k, d string) string {
		if v := os.Getenv(k); v != "" {
			return v
		}
		return d
	}
	return Config{Port: get("PRODUCT_SERVICE_PORT", "8082"), DatabaseURL: "postgres://" + get("POSTGRES_USER", "shopcore") + ":" + get("POSTGRES_PASSWORD", "shopcore") + "@" + get("POSTGRES_HOST", "localhost") + ":" + get("POSTGRES_PORT", "5432") + "/shopcore_products?sslmode=" + get("POSTGRES_SSLMODE", "disable"), PublicKeyPath: get("JWT_PUBLIC_KEY_PATH", "./secrets/public.pem"), InternalServiceKey: get("INTERNAL_SERVICE_KEY", "")}
}
