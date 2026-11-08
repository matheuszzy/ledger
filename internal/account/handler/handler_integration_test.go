package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/matheuszzy/ledger/internal/account/repository"
	"github.com/matheuszzy/ledger/internal/account/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newServer starts the real repository, service and routes behind a local HTTP server and returns its URL.
func newServer(t *testing.T) string {
	t.Helper()
	server := httptest.NewServer(New(service.New(repository.NewMemory())).Router())
	t.Cleanup(server.Close)
	return server.URL
}

// send makes one HTTP request and returns the response with its trimmed body; safe to call from goroutines.
func send(t *testing.T, server, method, target, body string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, server+target, strings.NewReader(body))
	if err != nil {
		t.Errorf("build %s %s: %v", method, target, err)
		return &http.Response{}, ""
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("%s %s: %v", method, target, err)
		return &http.Response{}, ""
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("read %s %s: %v", method, target, err)
	}
	return resp, strings.TrimSpace(string(raw))
}

// do makes one HTTP request and returns the status and trimmed body.
func do(t *testing.T, server, method, target, body string) (int, string) {
	t.Helper()
	resp, text := send(t, server, method, target, body)
	return resp.StatusCode, text
}

// step is one request in a workflow and the response it must get.
type step struct {
	name       string
	method     string
	target     string
	body       string
	wantStatus int
	wantBody   string
}

// runSteps sends each step in order and stops at the first unexpected response.
func runSteps(t *testing.T, server string, steps []step) {
	t.Helper()
	for _, s := range steps {
		status, body := do(t, server, s.method, s.target, s.body)
		if status != s.wantStatus || body != s.wantBody {
			t.Fatalf("%s: got %d %q, want %d %q", s.name, status, body, s.wantStatus, s.wantBody)
		}
	}
}

// TestREADMEScenario replays every step of the spec in order against one server.
func TestREADMEScenario(t *testing.T) {
	server := newServer(t)
	runSteps(t, server, []step{
		{"reset state", "POST", "/reset", "", 200, "OK"},
		{"balance of non-existing account", "GET", "/balance?account_id=1234", "", 404, "0"},
		{"create account with initial balance", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`, 201, `{"destination":{"id":"100","balance":10}}`},
		{"deposit into existing account", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`, 201, `{"destination":{"id":"100","balance":20}}`},
		{"balance of existing account", "GET", "/balance?account_id=100", "", 200, "20"},
		{"withdraw from non-existing account", "POST", "/event", `{"type":"withdraw", "origin":"200", "amount":10}`, 404, "0"},
		{"withdraw from existing account", "POST", "/event", `{"type":"withdraw", "origin":"100", "amount":5}`, 201, `{"origin":{"id":"100","balance":15}}`},
		{"transfer from existing account", "POST", "/event", `{"type":"transfer", "origin":"100", "amount":15, "destination":"300"}`, 201, `{"origin":{"id":"100","balance":0},"destination":{"id":"300","balance":15}}`},
		{"transfer from non-existing account", "POST", "/event", `{"type":"transfer", "origin":"200", "amount":15, "destination":"300"}`, 404, "0"},
		// State left behind by the scenario is visible through GET /balance.
		{"final balance of 100", "GET", "/balance?account_id=100", "", 200, "0"},
		{"final balance of 300", "GET", "/balance?account_id=300", "", 200, "15"},
		{"failed events did not create 200", "GET", "/balance?account_id=200", "", 404, "0"},
	})
}

// TestWorkflow_AccountLifecycle follows one account from creation until it is emptied and refilled.
func TestWorkflow_AccountLifecycle(t *testing.T) {
	server := newServer(t)
	runSteps(t, server, []step{
		{"open with a deposit", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":50}`, 201, `{"destination":{"id":"100","balance":50}}`},
		{"partial withdraw", "POST", "/event", `{"type":"withdraw", "origin":"100", "amount":20}`, 201, `{"origin":{"id":"100","balance":30}}`},
		{"withdraw the exact balance", "POST", "/event", `{"type":"withdraw", "origin":"100", "amount":30}`, 201, `{"origin":{"id":"100","balance":0}}`},
		{"emptied account still exists", "GET", "/balance?account_id=100", "", 200, "0"},
		{"overdraw is rejected", "POST", "/event", `{"type":"withdraw", "origin":"100", "amount":1}`, 422, "0"},
		{"rejected overdraw keeps balance", "GET", "/balance?account_id=100", "", 200, "0"},
		{"refill emptied account", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":5}`, 201, `{"destination":{"id":"100","balance":5}}`},
	})
}

// TestWorkflow_TransferChain moves money around a cycle of accounts and checks the total never changes.
func TestWorkflow_TransferChain(t *testing.T) {
	server := newServer(t)
	runSteps(t, server, []step{
		{"fund 100", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":100}`, 201, `{"destination":{"id":"100","balance":100}}`},
		{"100 to new 200", "POST", "/event", `{"type":"transfer", "origin":"100", "destination":"200", "amount":30}`, 201, `{"origin":{"id":"100","balance":70},"destination":{"id":"200","balance":30}}`},
		{"200 to new 300", "POST", "/event", `{"type":"transfer", "origin":"200", "destination":"300", "amount":20}`, 201, `{"origin":{"id":"200","balance":10},"destination":{"id":"300","balance":20}}`},
		{"300 back to 100", "POST", "/event", `{"type":"transfer", "origin":"300", "destination":"100", "amount":5}`, 201, `{"origin":{"id":"300","balance":15},"destination":{"id":"100","balance":75}}`},
		{"200 cannot send more than it holds", "POST", "/event", `{"type":"transfer", "origin":"200", "destination":"300", "amount":11}`, 422, "0"},
		// 75 + 10 + 15 = 100: the failed transfer moved nothing and no money was created or lost.
		{"balance of 100", "GET", "/balance?account_id=100", "", 200, "75"},
		{"balance of 200", "GET", "/balance?account_id=200", "", 200, "10"},
		{"balance of 300", "GET", "/balance?account_id=300", "", 200, "15"},
	})
}

// TestWorkflow_ResetThenReuseIDs checks that reset wipes every account and old IDs start from zero.
func TestWorkflow_ResetThenReuseIDs(t *testing.T) {
	server := newServer(t)
	runSteps(t, server, []step{
		{"deposit before reset", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`, 201, `{"destination":{"id":"100","balance":10}}`},
		{"transfer before reset", "POST", "/event", `{"type":"transfer", "origin":"100", "destination":"300", "amount":4}`, 201, `{"origin":{"id":"100","balance":6},"destination":{"id":"300","balance":4}}`},
		{"reset", "POST", "/reset", "", 200, "OK"},
		{"origin is gone", "GET", "/balance?account_id=100", "", 404, "0"},
		{"destination is gone", "GET", "/balance?account_id=300", "", 404, "0"},
		{"reused id starts from zero", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":5}`, 201, `{"destination":{"id":"100","balance":5}}`},
	})
}

// TestWorkflow_ConcurrentTransfers races transfers out of one account and checks it is never overdrawn.
func TestWorkflow_ConcurrentTransfers(t *testing.T) {
	server := newServer(t)
	do(t, server, "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`)

	const workers = 50
	var succeeded atomic.Int32
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			status, _ := do(t, server, "POST", "/event", `{"type":"transfer", "origin":"100", "destination":"300", "amount":1}`)
			switch status {
			case http.StatusCreated:
				succeeded.Add(1)
			case http.StatusUnprocessableEntity:
			default:
				t.Errorf("transfer status = %d, want 201 or 422", status)
			}
		})
	}
	wg.Wait()

	if got := succeeded.Load(); got != 10 {
		t.Fatalf("successful transfers = %d, want 10 (one per unit of balance)", got)
	}
	runSteps(t, server, []step{
		{"origin is emptied, not negative", "GET", "/balance?account_id=100", "", 200, "0"},
		{"destination received every unit", "GET", "/balance?account_id=300", "", 200, "10"},
	})
}

// TestResponses_ContentType checks the media type clients receive for each kind of response.
func TestResponses_ContentType(t *testing.T) {
	server := newServer(t)
	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   string
	}{
		{"event", "POST", "/event", `{"type":"deposit", "destination":"100", "amount":10}`, "application/json; charset=utf-8"},
		{"balance", "GET", "/balance?account_id=100", "", "text/plain; charset=utf-8"},
		{"error", "GET", "/balance?account_id=999", "", "text/plain; charset=utf-8"},
	}
	for _, tt := range tests {
		resp, _ := send(t, server, tt.method, tt.target, tt.body)
		if got := resp.Header.Get("Content-Type"); got != tt.want {
			t.Fatalf("%s: Content-Type = %q, want %q", tt.name, got, tt.want)
		}
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
			server := newServer(t)
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
			status, body := do(t, newServer(t), "GET", tt.target, "")

			if status != tt.wantStatus || body != tt.wantBody {
				t.Fatalf("got %d %q, want %d %q", status, body, tt.wantStatus, tt.wantBody)
			}
		})
	}
}

// TestBalance_HasNoSideEffects checks that reading a missing account never creates it.
func TestBalance_HasNoSideEffects(t *testing.T) {
	server := newServer(t)

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

func TestRoutes_WrongMethod(t *testing.T) {
	server := newServer(t)

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
