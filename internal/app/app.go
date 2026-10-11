package app

import (
	"net/http"

	"github.com/llem0on/ai-data-extractor/internal/config"
	"github.com/llem0on/ai-data-extractor/internal/document"
	"github.com/llem0on/ai-data-extractor/internal/extract"
)

func New(cfg config.Config) (*extract.Extractor, document.Options, error) {
	schema, err := extract.LoadSchema(cfg.SchemaPath)
	if err != nil {
		return nil, document.Options{}, err
	}
	ex := &extract.Extractor{
		Schema: schema,
		LLM: &extract.LLM{
			BaseURL: cfg.LLMBaseURL,
			APIKey:  cfg.LLMAPIKey,
			Model:   cfg.LLMModel,
			HTTP:    &http.Client{Timeout: cfg.LLMTimeout},
		},
	}
	return ex, document.Options{PDFPages: cfg.PDFPages, MaxChars: cfg.MaxChars}, nil
}
