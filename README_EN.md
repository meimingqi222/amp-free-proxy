# Amp Free Proxy

A lightweight proxy that enables free tier features for Amp CLI in smart mode.

## Features

- **Free Web Search**: Forces `web_search` and `read_web_page` tools to use free tier
- **Model Mapping**: Redirect expensive models to free tier `claude-haiku-4-5-20251001`
- **Docker Support**: Easy deployment with Docker/docker-compose

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the binary for your platform from [Releases](https://github.com/aftely1337/amp-free-proxy/releases):

- Windows: `amp-free-proxy-windows-amd64.exe`
- Linux: `amp-free-proxy-linux-amd64`
- macOS: `amp-free-proxy-darwin-amd64`

### Option 2: Build from Source

Requires Go 1.21+

```bash
git clone https://github.com/aftely1337/amp-free-proxy.git
cd amp-free-proxy
go build -o amp-free-proxy
```

Cross-compile for other platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o amp-free-proxy-linux-amd64

# macOS
GOOS=darwin GOARCH=amd64 go build -o amp-free-proxy-darwin-amd64

# Windows
GOOS=windows GOARCH=amd64 go build -o amp-free-proxy-windows-amd64.exe
```

### Option 3: Docker

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
