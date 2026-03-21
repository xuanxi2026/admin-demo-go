package main

import (
	"fmt"
	"log"

	"admin-demo-go/internal/bootstrap"
	"admin-demo-go/internal/config"
	"admin-demo-go/internal/router"
)

func main() {
	cfg, err := config.Load(config.ConfigPath())
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		log.Fatalf("bootstrap app failed: %v", err)
	}

	r := router.New(app)
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("admin-demo-go started at %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
