package main

import (
	"log"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
	"app/internal/server"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("gagal migrasi: %v", err)
	}
	if err := database.Seed(db, cfg); err != nil {
		log.Fatalf("gagal seeding: %v", err)
	}
	chat := ai.NewClient(cfg.OpenAIEndpoint, cfg.OpenAIKey, cfg.OpenAIModel)
	if err := server.NewRouter(cfg, db, chat, ai.NewSearchClient(cfg.SearchEndpoint)).Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
