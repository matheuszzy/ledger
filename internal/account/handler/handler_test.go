package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/matheuszzy/ledger/internal/account/repository"
	"github.com/matheuszzy/ledger/internal/account/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newServer wires the real repository and service behind the handler routes.
func newServer() http.Handler {
	return New(service.New(repository.NewMemory())).Router()
}

// do sends one request to the server and returns the status and body.
func do(t *testing.T, server http.Handler, method, target, body string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec.Code, strings.TrimSpace(rec.Body.String())
}

// TestREADMEScenario replays every step of the spec in order against one server.
func TestREADMEScenario(t *testing.T) {
	server := newServer()
	steps := []struct {
		name       string
		method     string
		target     string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"reset state", "POST", "/reset", "", 200, "OK"},
		{"balance of non-existing account", "GET", "/balance?account_id=1234", "", 404, "0"},
		{"create account with initial balance", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`, 201, `{"destination":{"id":"100","balance":10}}`},
		{"deposit into existing account", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`, 201, `{"destination":{"id":"100","balance":20}}`},
		{"balance of existing account", "GET", "/balance?account_id=100", "", 200, "20"},
		{"withdraw from non-existing account", "POST", "/event", `{"type":"withdraw", "origin":"200", "amount":10}`, 404, "0"},
		{"withdraw from existing account", "POST", "/event", `{"type":"withdraw", "origin":"100", "amount":5}`, 201, `{"origin":{"id":"100","balance":15}}`},
		{"transfer from existing account", "POST", "/event", `{"type":"transfer", "origin":"100", "amount":15, "destination":"300"}`, 201, `{"origin":{"id":"100","balance":0},"destination":{"id":"300","balance":15}}`},
		{"transfer from non-existing account", "POST", "/event", `{"type":"transfer", "origin":"200", "amount":15, "destination":"300"}`, 404, "0"},
	}
	for _, step := range steps {
		status, body := do(t, server, step.method, step.target, step.body)
		if status != step.wantStatus || body != step.wantBody {
			t.Fatalf("%s: got %d %q, want %d %q", step.name, status, body, step.wantStatus, step.wantBody)
		}
	}

	// State left behind by the scenario is visible through GET /balance.
	for id, want := range map[string]string{"100": "0", "300": "15"} {
		if status, body := do(t, server, "GET", "/balance?account_id="+id, ""); status != 200 || body != want {
			t.Fatalf("final balance of %s: got %d %q, want 200 %q", id, status, body, want)
		}
	}
	// The failed requests must not have created account 200.
	if status, _ := do(t, server, "GET", "/balance?account_id=200", ""); status != 404 {
		t.Fatalf("account 200 status = %d, want 404 (failed events must not create it)", status)
	}
}

func TestEvent_Errors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"insufficient funds on withdraw", `{"type":"withdraw", "origin":"100", "amount":11}`, 422, "0"},
		{"insufficient funds on transfer", `{"type":"transfer", "origin":"100", "destination":"300", "amount":11}`, 422, "0"},
		{"malformed json", `{"type":`, 400, errInvalidBody.Error()},
		{"empty body", ``, 400, errInvalidBody.Error()},
		{"amount with wrong type", `{"type":"deposit", "destination":"100", "amount":"ten"}`, 400, errInvalidBody.Error()},
		{"unknown event type", `{"type":"refund", "destination":"100", "amount":1}`, 400, errUnknownType.Error()},
		{"missing event type", `{"destination":"100", "amount":1}`, 400, errUnknownType.Error()},
		{"deposit without destination", `{"type":"deposit", "amount":1}`, 400, "AC002 - The ID must be a valid integer"},
		{"withdraw without origin", `{"type":"withdraw", "amount":1}`, 400, "AC002 - The ID must be a valid integer"},
		{"transfer without destination", `{"type":"transfer", "origin":"100", "amount":1}`, 400, "AC002 - The ID must be a valid integer"},
		{"transfer to the same account", `{"type":"transfer", "origin":"100", "destination":"100", "amount":5}`, 400, "AC004 - Origin and destination must be different accounts"},
		{"negative deposit", `{"type":"deposit", "destination":"100", "amount":-1}`, 400, "AC001 - Amount cannot be negative"},
		{"negative withdraw", `{"type":"withdraw", "origin":"100", "amount":-1}`, 400, "AC001 - Amount cannot be negative"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newServer()
			do(t, server, "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`)

			status, body := do(t, server, "POST", "/event", tt.body)

			if status != tt.wantStatus || body != tt.wantBody {
				t.Fatalf("got %d %q, want %d %q", status, body, tt.wantStatus, tt.wantBody)
			}
			// A rejected event must leave the existing account untouched and create nothing.
			if status, body := do(t, server, "GET", "/balance?account_id=100", ""); status != 200 || body != "10" {
				t.Fatalf("balance after rejected event: got %d %q, want 200 \"10\"", status, body)
			}
			if status, _ := do(t, server, "GET", "/balance?account_id=300", ""); status != 404 {
				t.Fatalf("account 300 status = %d, want 404 (rejected event must not create it)", status)
			}
		})
	}
}

func TestBalance_Errors(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantBody   string
	}{
		{"unknown account", "/balance?account_id=1234", 404, "0"},
		{"missing account_id", "/balance", 400, "AC002 - The ID must be a valid integer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := do(t, newServer(), "GET", tt.target, "")

			if status != tt.wantStatus || body != tt.wantBody {
				t.Fatalf("got %d %q, want %d %q", status, body, tt.wantStatus, tt.wantBody)
			}
		})
	}
}

// TestBalance_HasNoSideEffects checks that reading a missing account never creates it.
func TestBalance_HasNoSideEffects(t *testing.T) {
	server := newServer()

	for range 3 {
		if status, _ := do(t, server, "GET", "/balance?account_id=100", ""); status != 404 {
			t.Fatalf("GET /balance status = %d, want 404", status)
		}
	}

	status, body := do(t, server, "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`)
	if status != 201 || body != `{"destination":{"id":"100","balance":10}}` {
		t.Fatalf("deposit after GETs: got %d %q, want a fresh account with balance 10", status, body)
	}
}

func TestReset_ClearsState(t *testing.T) {
	server := newServer()
	do(t, server, "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`)

	if status, body := do(t, server, "POST", "/reset", ""); status != 200 || body != "OK" {
		t.Fatalf("reset: got %d %q, want 200 \"OK\"", status, body)
	}
	if status, _ := do(t, server, "GET", "/balance?account_id=100", ""); status != 404 {
		t.Fatalf("account 100 status after reset = %d, want 404", status)
	}
}

func TestRoutes_WrongMethod(t *testing.T) {
	server := newServer()

	for _, tt := range []struct{ method, target string }{
		{"GET", "/event"},
		{"POST", "/balance?account_id=100"},
		{"GET", "/reset"},
	} {
		if status, _ := do(t, server, tt.method, tt.target, ""); status != http.StatusMethodNotAllowed {
			t.Fatalf("%s %s status = %d, want 405", tt.method, tt.target, status)
		}
	}
}
