# Schema

The schema decides which fields are extracted. The built-in one lives in [`internal/extract/default_schema.json`](../internal/extract/default_schema.json). Set `SCHEMA_PATH` to use your own.

## Field

```json
{
  "name": "effective_date",
  "description": "Date the document takes effect",
  "type": "date",
  "labels": ["tanggal berlaku", "mulai berlaku(?:\\s*tanggal)?", "effective\\s*date"]
}
```

| Key | Required | Description |
|---|---|---|
| `name` | yes | Key in the output, unique |
| `description` | yes | Sent to the LLM; write it the way you would explain the field to a person |
| `type` | no | `string` (default), `date` or `enum` |
| `labels` | no | Regular expressions for the rule step, matched case-insensitively at the start of a line or cell. Without labels the field is LLM-only |
| `values` | enum only | Allowed values |

## Built-in fields

| Field | Type | Example labels |
|---|---|---|
| `document_type` | enum: SOP, Work Instruction, Policy, Form, Other | `jenis dokumen`, `document type` |
| `document_number` | string | `nomor dokumen`, `no. dok`, `document no` |
| `title` | string | `judul`, `perihal`, `title` |
| `department` | string | `departemen`, `divisi`, `department` |
| `effective_date` | date | `tanggal berlaku`, `mulai berlaku`, `effective date` |
| `revision` | string | `revisi`, `rev.`, `version` |
| `prepared_by` | string | `dibuat oleh`, `disusun oleh`, `prepared by` |
| `approved_by` | string | `disetujui oleh`, `disahkan oleh`, `approved by` |

## Writing good labels

- Put the longest, most specific label first: `nomor dokumen` before `nomor`.
- Do not include the separator; `:` or `-` is handled for you.
- Escape regex characters in JSON: `no\\.?\\s*dok` matches `No. Dok` and `No Dok`.
- A field the rules rarely find (for example a title with no label) can skip `labels` entirely and rely on the LLM.

## Example

[`examples/invoice.schema.json`](../examples/invoice.schema.json) turns the extractor into an invoice reader:

```bash
SCHEMA_PATH=examples/invoice.schema.json make run
```
