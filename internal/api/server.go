package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/llem0on/ai-data-extractor/internal/document"
	"github.com/llem0on/ai-data-extractor/internal/extract"
)

const maxUploadBytes = 50 << 20

type Server struct {
	Extractor *extract.Extractor
	Options   document.Options
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /schema", s.schema)
	mux.HandleFunc("POST /extract", s.extract)
	return logRequests(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "llm": s.Extractor.LLM.Enabled()})
}

func (s *Server) schema(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.Extractor.Schema)
}

func (s *Server) extract(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "multipart field 'file' (.pdf or .docx) is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" && ext != ".docx" {
		writeError(w, http.StatusBadRequest, "only .pdf and .docx files are supported")
		return
	}
	tmp, err := os.CreateTemp("", "extract-*"+ext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tmp.Close()

	doc, err := document.ReadFile(r.Context(), tmp.Name(), s.Options)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	doc.Name = header.Filename

	res := s.Extractor.Run(r.Context(), doc)
	if r.URL.Query().Get("include_text") == "true" {
		res.Text = doc.Text
	}
	log.Printf("extract file=%q fields=%d missing=%d llm=%v took_ms=%d", doc.Name, len(res.Fields), len(res.Missing), res.LLMUsed, res.TookMS)
	writeJSON(w, http.StatusOK, res)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
