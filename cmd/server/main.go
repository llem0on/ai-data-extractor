package main

import (
	"log"
	"net/http"
	"time"

	"github.com/llem0on/ai-data-extractor/internal/api"
	"github.com/llem0on/ai-data-extractor/internal/app"
	"github.com/llem0on/ai-data-extractor/internal/config"
)

func main() {
	cfg := config.Load()
	ex, opts, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if !ex.LLM.Enabled() {
		log.Print("LLM_BASE_URL or LLM_MODEL not set: running with rules only")
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           (&api.Server{Extractor: ex, Options: opts}).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("listening on %s (%d fields)", cfg.Addr, len(ex.Schema.Fields))
	log.Fatal(srv.ListenAndServe())
}
