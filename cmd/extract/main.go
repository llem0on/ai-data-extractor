package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/llem0on/ai-data-extractor/internal/app"
	"github.com/llem0on/ai-data-extractor/internal/config"
	"github.com/llem0on/ai-data-extractor/internal/document"
	"github.com/llem0on/ai-data-extractor/internal/extract"
)

func main() {
	withText := flag.Bool("text", false, "include the extracted text in the output")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: extract [-text] file.pdf|file.docx ...")
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	ex, opts, err := app.New(config.Load())
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	results := make([]extract.Result, 0, flag.NArg())
	failed := false
	for _, path := range flag.Args() {
		doc, err := document.ReadFile(ctx, path, opts)
		if err != nil {
			log.Printf("%s: %v", path, err)
			failed = true
			continue
		}
		res := ex.Run(ctx, doc)
		if *withText {
			res.Text = doc.Text
		}
		results = append(results, res)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		log.Fatal(err)
	}
	if failed {
		os.Exit(1)
	}
}
