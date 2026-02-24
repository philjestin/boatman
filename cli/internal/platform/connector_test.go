package platform_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/philjestin/boatmanmode/internal/platform"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

func TestTryConnectSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	conn := platform.TryConnect(context.Background(), srv.URL, storage.Scope{OrgID: "org1"})
	if !conn.IsConnected() {
		t.Error("expected connected")
	}
	if conn.Client() == nil {
		t.Error("expected non-nil client")
	}
}

func TestTryConnectFail(t *testing.T) {
	conn := platform.TryConnect(context.Background(), "http://localhost:1", storage.Scope{})
	if conn.IsConnected() {
		t.Error("expected not connected to unreachable server")
	}
}

func TestTryConnectEmpty(t *testing.T) {
	conn := platform.TryConnect(context.Background(), "", storage.Scope{})
	if conn.IsConnected() {
		t.Error("expected not connected with empty URL")
	}
	if conn.Client() != nil {
		t.Error("expected nil client with empty URL")
	}
}
