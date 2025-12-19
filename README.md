# Amp Free Proxy

A lightweight proxy that enables free tier features for Amp CLI in smart mode.

## Features

- **Free Web Search**: Forces `web_search` and `read_web_page` tools to use free tier
- **Model Mapping**: Redirect expensive models to free tier `claude-haiku-4-5-20251001`
- **Docker Support**: Easy deployment with Docker/docker-compose

## Installation

### Binary

```bash
go build -o amp-free-proxy
```

### Docker

```bash
docker-compose up -d
```

## Configuration

Create `config.yaml` (auto-loaded from current directory):

```yaml
port: 8318
upstream: "http://127.0.0.1:8317"  # CLIProxyAPI or https://ampcode.com

model-mappings:
  - from: "claude-opus-4-5-20251101"
    to: "claude-haiku-4-5-20251001"
  - from: "claude-sonnet-4-20250514"
    to: "claude-haiku-4-5-20251001"
```

## Usage

```bash
# With config file (default: config.yaml)
./amp-free-proxy

# With custom config
./amp-free-proxy -config /path/to/config.yaml

# With command line flags
./amp-free-proxy -port 8318 -upstream https://ampcode.com
```

Then configure Amp CLI upstream to `http://127.0.0.1:8318`.

## Architecture

```
Amp CLI → amp-free-proxy (8318) → CLIProxyAPI (8317) → ampcode.com
```

Or direct:

```
Amp CLI → amp-free-proxy (8318) → ampcode.com
```

## License

MIT
