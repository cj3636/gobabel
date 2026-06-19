# GoBabel

GoBabel is a Go API inspired by the Library of Babel. It has deterministic corpus browsing and exact constructive text locations.

## Theory note

GoBabel does **not** brute-force an infinite random library. Exact lookup of arbitrary user text is impossible without brute force, storage, or a constructive address. GoBabel uses constructive sealed coordinates: `/v1/locate` packs submitted code-ascii-v1 text with `pack98-v1`, seals it with authenticated encryption, and returns a Library-style address that can reconstruct a 5000-byte page containing that text.

## Quick start

```sh
go run ./cmd/gobabel serve
```

Defaults: `gobabel_ADDR=:3000`, `gobabel_SEAL_MODE=public`, `gobabel_LOG_LEVEL=info`.

## API examples

```sh
curl -s localhost:3000/v1/locate \
  -H 'content-type: application/json' \
  -d '{"text":"hello world","placement":"hash"}'
```

Fetch the returned `range_url`:

```sh
curl -s localhost:3000/v1/book/bf1/.../100:111
```

Browse corpus pages:

```sh
curl -s localhost:3000/v1/corpus/hex/zk4n2/wall/3/shelf/1/book/k9Lm2/page/412
```

## Address format

Sealed anchor addresses use `/v1/book/bf1/{base64url-segments}` and optional exclusive byte range suffix `{start}:{end}`. The encrypted payload contains flags, alphabet and codec IDs, page size, placement, start, text length, filler seed, and packed text. Corpus addresses use visible coordinates and SHA-256/ChaCha20 deterministic generation.

## Public vs private sealing

Public mode derives a portable deterministic project key and provides obfuscation plus integrity. Private mode requires `gobabel_SEAL_KEY` as base64url raw 32 bytes and makes addresses decodable only by servers with that key.

## Development

```sh
make fmt
make test
make vet
make bench
```

CLI:

```sh
go run ./cmd/gobabel locate "hello world"
go run ./cmd/gobabel validate "hello"
go run ./cmd/gobabel page "/v1/book/bf1/..."
```
