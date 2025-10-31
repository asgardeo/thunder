// Package opensearch provides an OpenSearch output adapter for analytics events.
package opensearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"net/url"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/observability/adapter"
	httpclient "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "OpenSearchOutputAdapter"

// Config contains the configuration required to connect to OpenSearch.
type Config struct {
	Endpoint   string
	Index      string
	Username   string
	Password   string
	APIKey     string
	Headers    map[string]string
	UseEventID bool
	Timeout    time.Duration
}

// OpenSearchAdapter delivers events to an OpenSearch cluster.
type OpenSearchAdapter struct {
	endpoint   string
	index      string
	username   string
	password   string
	apiKey     string
	headers    map[string]string
	useEventID bool
	client     httpclient.HTTPClientInterface
	logger     *log.Logger

	mu     sync.RWMutex
	closed bool
}

var _ adapter.OutputAdapter = (*OpenSearchAdapter)(nil)

// NewOpenSearchAdapter creates a new adapter for OpenSearch destinations.
func NewOpenSearchAdapter(cfg Config) (*OpenSearchAdapter, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("opensearch endpoint is required")
	}
	if cfg.Index == "" {
		return nil, fmt.Errorf("opensearch index is required")
	}

	client := selectHTTPClient(cfg.Timeout)
	headersCopy := make(map[string]string, len(cfg.Headers))
	for k, v := range cfg.Headers {
		headersCopy[k] = v
	}

	adapter := &OpenSearchAdapter{
		endpoint:   cfg.Endpoint,
		index:      cfg.Index,
		username:   cfg.Username,
		password:   cfg.Password,
		apiKey:     cfg.APIKey,
		headers:    headersCopy,
		useEventID: cfg.UseEventID,
		client:     client,
		logger:     log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}

	if adapter.useEventID {
		adapter.logger.Info(
			"OpenSearch adapter initialized with event id indexing",
			log.String("endpoint", cfg.Endpoint),
			log.String("index", cfg.Index),
		)
	} else {
		adapter.logger.Info(
			"OpenSearch adapter initialized",
			log.String("endpoint", cfg.Endpoint),
			log.String("index", cfg.Index),
		)
	}

	return adapter, nil
}

func selectHTTPClient(timeout time.Duration) httpclient.HTTPClientInterface {
	if timeout <= 0 {
		return httpclient.NewHTTPClient()
	}
	return httpclient.NewHTTPClientWithTimeout(timeout)
}

// Write sends the provided event payload to OpenSearch.
func (oa *OpenSearchAdapter) Write(data []byte) error {
	oa.mu.RLock()
	if oa.closed {
		oa.mu.RUnlock()
		return fmt.Errorf("opensearch adapter is closed")
	}
	oa.mu.RUnlock()

	docID := ""
	if oa.useEventID {
		docID = extractEventID(data)
	}

	targetURL, err := oa.buildEndpoint(docID)
	if err != nil {
		oa.logger.Error("Failed to build OpenSearch request URL", log.Error(err))
		return fmt.Errorf("failed to build opensearch request url: %w", err)
	}

	req, err := nethttp.NewRequest(oa.httpMethod(docID), targetURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create opensearch request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	oa.applyAuth(req)
	oa.applyHeaders(req)

	resp, err := oa.client.Do(req)
	if err != nil {
		oa.logger.Error("Failed to deliver event to OpenSearch", log.Error(err))
		return fmt.Errorf("failed to deliver event to opensearch: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			oa.logger.Error("Failed to close response body", log.Error(closeErr))
		}
	}()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		oa.logger.Error(
			"OpenSearch responded with error",
			log.Int("statusCode", resp.StatusCode),
			log.String("response", string(body)),
		)
		return fmt.Errorf("opensearch returned status %d", resp.StatusCode)
	}

	if oa.logger.IsDebugEnabled() {
		oa.logger.Debug("Event indexed in OpenSearch", log.String("index", oa.index), log.String("documentID", docID))
	}

	return nil
}

func (oa *OpenSearchAdapter) buildEndpoint(docID string) (string, error) {
	if docID == "" {
		return url.JoinPath(oa.endpoint, oa.index, "_doc")
	}
	return url.JoinPath(oa.endpoint, oa.index, "_doc", docID)
}

func (oa *OpenSearchAdapter) httpMethod(docID string) string {
	if docID == "" {
		return nethttp.MethodPost
	}
	return nethttp.MethodPut
}

func (oa *OpenSearchAdapter) applyAuth(req *nethttp.Request) {
	if oa.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", oa.apiKey))
		return
	}
	if oa.username != "" {
		req.SetBasicAuth(oa.username, oa.password)
	}
}

func (oa *OpenSearchAdapter) applyHeaders(req *nethttp.Request) {
	for k, v := range oa.headers {
		if k == "" || v == "" {
			continue
		}
		req.Header.Set(k, v)
	}
}

// Flush is a no-op because requests are delivered synchronously.
func (oa *OpenSearchAdapter) Flush() error {
	return nil
}

// Close marks the adapter as closed.
func (oa *OpenSearchAdapter) Close() error {
	oa.mu.Lock()
	defer oa.mu.Unlock()

	oa.closed = true
	return nil
}

// GetName returns the adapter name.
func (oa *OpenSearchAdapter) GetName() string {
	return "OpenSearchAdapter"
}

func extractEventID(data []byte) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}

	value, ok := payload["event_id"]
	if !ok {
		return ""
	}

	id, ok := value.(string)
	if !ok {
		return ""
	}

	return id
}
