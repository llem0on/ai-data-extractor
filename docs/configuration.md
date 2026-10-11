# Configuration

All settings are environment variables. Copy `.env.example` to `.env`; `make run` and `make extract` load it.

| Variable | Default | Description |
|---|---|---|
| `ADDR` | `:8090` | Listen address |
| `LLM_BASE_URL` | *(empty)* | OpenAI-compatible base URL ending in `/v1`, e.g. `http://localhost:4000/v1` |
| `LLM_API_KEY` | *(empty)* | Sent as `Authorization: Bearer` |
| `LLM_MODEL` | *(empty)* | Chat model name. Empty, or empty `LLM_BASE_URL`, means rules only |
| `LLM_TIMEOUT` | `60s` | Timeout of one LLM request |
| `SCHEMA_PATH` | *(built-in)* | Path to a custom schema, see [schema.md](schema.md) |
| `PDF_PAGES` | `2` | Pages read from the start of a PDF. `0` reads every page |
| `MAX_CHARS` | `6000` | Characters of document text passed to rules and LLM |

Any server that implements `POST /v1/chat/completions` works: OpenAI, LiteLLM, vLLM, Ollama (`/v1`), LM Studio and similar. Small instruction-tuned models handle this task well because the prompt is short and the output is a flat JSON object.

## Tuning

| Situation | Change |
|---|---|
| Metadata sits on a cover page plus a later approval page | raise `PDF_PAGES`, or set it to `0` |
| LLM responses are slow or expensive | lower `MAX_CHARS`; the header is usually within the first 2000 characters |
| A field is often wrong from rules | tighten or remove its broadest label in the schema |
| A field is often missing | add the label wording your documents use |
