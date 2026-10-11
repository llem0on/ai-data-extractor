# API

Base URL: `http://localhost:8090`

## `POST /extract`

Extract fields from one document.

| Field | Type | Required | Description |
|---|---|---|---|
| `file` | file | yes | `.pdf` or `.docx`, max 50 MB |

| Query | Description |
|---|---|
| `include_text=true` | Add the marked text that rules and LLM saw |

```bash
curl -s -F file=@procedure.pdf "localhost:8090/extract?include_text=true"
```

```json
{
  "document": { "name": "procedure.pdf", "format": "pdf" },
  "fields": {
    "document_number": { "value": "DOC-OPS 07.002", "source": "rule", "verified": true },
    "effective_date":  { "value": "2025-08-12", "source": "rule", "verified": true },
    "title":           { "value": "Data Backup and Restore", "source": "llm", "verified": true }
  },
  "missing": ["prepared_by"],
  "llm_used": true,
  "took_ms": 1840,
  "text": "[HEADING] STANDARD OPERATING PROCEDURE\n[FIELD] ..."
}
```

| Status | Meaning |
|---|---|
| `200` | Extracted; check `missing` and `llm_error` |
| `400` | No file, or not a `.pdf` / `.docx` |
| `422` | File could not be read (corrupt, encrypted, no text) |

## `GET /schema`

Returns the active schema.

## `GET /health`

```json
{ "status": "ok", "llm": true }
```

## Command line

```bash
go run ./cmd/extract [-text] file.pdf file.docx ...
```

Prints a JSON array with one result per file. Exits with code `1` when any file could not be read.
