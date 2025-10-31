package opensearch

import (
	"encoding/base64"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type capturedRequest struct {
	method string
	path   string
	header nethttp.Header
	body   []byte
}

func newTestServer(responseStatus int, body string, captures chan<- capturedRequest) *httptest.Server {
	handler := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		payload, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		captures <- capturedRequest{
			method: r.Method,
			path:   r.URL.Path,
			header: r.Header.Clone(),
			body:   payload,
		}

		w.WriteHeader(responseStatus)
		_, _ = w.Write([]byte(body))
	})

	return httptest.NewServer(handler)
}

func TestOpenSearchAdapterWriteUsesEventID(t *testing.T) {
	t.Parallel()

	captures := make(chan capturedRequest, 1)
	server := newTestServer(nethttp.StatusCreated, `{"result":"created"}`, captures)
	defer server.Close()

	adapter, err := NewOpenSearchAdapter(Config{
		Endpoint:   server.URL,
		Index:      "thunder-events",
		UseEventID: true,
	})
	require.NoError(t, err)

	payload := []byte(`{"event_id":"1234","type":"test"}`)

	err = adapter.Write(payload)
	require.NoError(t, err)

	req := <-captures
	require.Equal(t, nethttp.MethodPut, req.method)
	require.Equal(t, "/thunder-events/_doc/1234", req.path)
	require.Equal(t, "application/json", req.header.Get("Content-Type"))
	require.Equal(t, string(payload), string(req.body))
}

func TestOpenSearchAdapterWriteWithoutEventID(t *testing.T) {
	t.Parallel()

	captures := make(chan capturedRequest, 1)
	server := newTestServer(nethttp.StatusCreated, `{"result":"created"}`, captures)
	defer server.Close()

	adapter, err := NewOpenSearchAdapter(Config{
		Endpoint:   server.URL,
		Index:      "thunder-events",
		UseEventID: true,
	})
	require.NoError(t, err)

	payload := []byte(`{"type":"test"}`)

	err = adapter.Write(payload)
	require.NoError(t, err)

	req := <-captures
	require.Equal(t, nethttp.MethodPost, req.method)
	require.Equal(t, "/thunder-events/_doc", req.path)
}

func TestOpenSearchAdapterHandlesErrorStatus(t *testing.T) {
	t.Parallel()

	captures := make(chan capturedRequest, 1)
	server := newTestServer(nethttp.StatusInternalServerError, `{"error":"boom"}`, captures)
	defer server.Close()

	adapter, err := NewOpenSearchAdapter(Config{
		Endpoint: server.URL,
		Index:    "thunder-events",
	})
	require.NoError(t, err)

	err = adapter.Write([]byte(`{"event_id":"1"}`))
	require.Error(t, err)
}

func TestOpenSearchAdapterClosed(t *testing.T) {
	t.Parallel()

	captures := make(chan capturedRequest, 1)
	server := newTestServer(nethttp.StatusCreated, `{"result":"created"}`, captures)
	defer server.Close()

	adapter, err := NewOpenSearchAdapter(Config{
		Endpoint: server.URL,
		Index:    "thunder-events",
	})
	require.NoError(t, err)

	require.NoError(t, adapter.Close())
	err = adapter.Write([]byte(`{"event_id":"1"}`))
	require.Error(t, err)
}

func TestOpenSearchAdapterAuthHeaders(t *testing.T) {
	t.Parallel()

	t.Run("api key takes precedence", func(t *testing.T) {
		captures := make(chan capturedRequest, 1)
		server := newTestServer(nethttp.StatusCreated, "{}", captures)
		defer server.Close()

		adapter, err := NewOpenSearchAdapter(Config{
			Endpoint: server.URL,
			Index:    "thunder-events",
			Username: "user",
			Password: "pass",
			APIKey:   "abc123",
		})
		require.NoError(t, err)

		require.NoError(t, adapter.Write([]byte(`{"event_id":"1"}`)))
		req := <-captures
		require.Equal(t, "ApiKey abc123", req.header.Get("Authorization"))
	})

	t.Run("basic auth fallback", func(t *testing.T) {
		captures := make(chan capturedRequest, 1)
		server := newTestServer(nethttp.StatusCreated, "{}", captures)
		defer server.Close()

		adapter, err := NewOpenSearchAdapter(Config{
			Endpoint: server.URL,
			Index:    "thunder-events",
			Username: "user",
			Password: "pass",
		})
		require.NoError(t, err)

		require.NoError(t, adapter.Write([]byte(`{"event_id":"1"}`)))
		req := <-captures

		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
		require.Equal(t, expected, req.header.Get("Authorization"))
	})
}
