package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultUpstream            = "https://ampcode.com"
	defaultPort                = 8318
	internalAPIPath            = "/api/internal"
	webSearchQuery             = "webSearch2"
	extractWebPageContentQuery = "extractWebPageContent"
	anthropicMessagesPath      = "/api/provider/anthropic/v1/messages"
)

var isFreeTierRequestRegex = regexp.MustCompile(`"isFreeTierRequest"\s*:\s*false`)

// Config represents the configuration file structure
type Config struct {
	Port               int            `yaml:"port"`
	Upstream           string         `yaml:"upstream"`
	EnableFreeSearch   bool           `yaml:"enable-free-search"`
	EnableModelMapping bool           `yaml:"enable-model-mapping"`
	ModelMappings      []ModelMapping `yaml:"model-mappings"`
}

// ModelMapping represents a model redirection rule
type ModelMapping struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	port := flag.Int("port", defaultPort, "Port to listen on")
	upstream := flag.String("upstream", defaultUpstream, "Upstream URL")
	configFile := flag.String("config", "config.yaml", "Path to config file (YAML)")
	flag.Parse()

	// Load config file (default: config.yaml in current directory)
	var modelMappings []ModelMapping
	enableFreeSearch := true
	enableModelMapping := true
	if _, err := os.Stat(*configFile); err == nil {
		cfg, err := loadConfig(*configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		if cfg.Port != 0 {
			*port = cfg.Port
		}
		if cfg.Upstream != "" {
			*upstream = cfg.Upstream
		}
		enableFreeSearch = cfg.EnableFreeSearch
		enableModelMapping = cfg.EnableModelMapping
		modelMappings = cfg.ModelMappings
	}

	upstreamURL, err := url.Parse(*upstream)
	if err != nil {
		log.Fatalf("Invalid upstream URL: %v", err)
	}

	// Build model mapping lookup
	modelMap := make(map[string]string)
	for _, m := range modelMappings {
		modelMap[m.From] = m.To
		log.Printf("Model mapping: %s -> %s", m.From, m.To)
	}

	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
	originalDirector := proxy.Director

	// Suppress context canceled errors (normal for streaming/SSE)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if err.Error() != "context canceled" {
			log.Printf("Proxy error: %v", err)
		}
	}

	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = upstreamURL.Host

		query := req.URL.RawQuery
		path := req.URL.Path

		// Intercept webSearch2 and extractWebPageContent requests
		if enableFreeSearch && path == internalAPIPath && (query == webSearchQuery || query == extractWebPageContentQuery) && req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				log.Printf("Warning: could not read request body for %s: %v", query, err)
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else if isFreeTierRequestRegex.Match(bodyBytes) {
				modifiedBody := isFreeTierRequestRegex.ReplaceAll(bodyBytes, []byte(`"isFreeTierRequest":true`))
				req.ContentLength = int64(len(modifiedBody))
				req.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))
				req.Body = io.NopCloser(bytes.NewBuffer(modifiedBody))
				log.Printf("Modified %s request to use free tier", query)
			} else {
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Intercept Anthropic messages and apply model mapping
		if enableModelMapping && path == anthropicMessagesPath && req.Body != nil && len(modelMap) > 0 {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				log.Printf("Warning: could not read request body for anthropic messages: %v", err)
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else {
				modifiedBody := bodyBytes
				modified := false

				// Check each model mapping
				for fromModel, toModel := range modelMap {
					pattern := regexp.MustCompile(`"model"\s*:\s*"` + regexp.QuoteMeta(fromModel) + `"`)
					if pattern.Match(modifiedBody) {
						modifiedBody = pattern.ReplaceAll(modifiedBody, []byte(`"model":"`+toModel+`"`))
						modified = true
						log.Printf("Mapped model: %s -> %s", fromModel, toModel)
						break
					}
				}

				if modified {
					req.ContentLength = int64(len(modifiedBody))
					req.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))
					req.Body = io.NopCloser(bytes.NewBuffer(modifiedBody))
					// Set X-Amp-Mode header to free for mapped models
					req.Header.Set("X-Amp-Mode", "free")
				} else {
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}
	}

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Free Proxy listening on %s, forwarding to %s", addr, *upstream)
	log.Printf("Configure upstream to: http://127.0.0.1:%d", *port)
	log.Printf("Free search enabled: %v", enableFreeSearch)
	log.Printf("Model mapping enabled: %v", enableModelMapping)
	if enableModelMapping && len(modelMappings) > 0 {
		log.Printf("Loaded %d model mapping(s)", len(modelMappings))
	}

	if err := http.ListenAndServe(addr, proxy); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Helper to check if string contains any of the patterns
func containsAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
