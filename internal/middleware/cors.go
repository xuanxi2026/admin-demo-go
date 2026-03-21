package middleware

import "github.com/gin-contrib/cors"

func CORS() cors.Config {
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "accessToken"}
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	return cfg
}
