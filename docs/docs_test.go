package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRegisterDocs(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":   {Data: []byte("<html>ui</html>")},
		"swagger.yaml": {Data: []byte("swagger: \"2.0\"")},
	}
	mux := http.NewServeMux()
	RegisterDocs(mux, fsys)

	tests := []struct {
		name     string
		path     string
		wantCode int
		wantBody string
	}{
		{"index at docs root", "/docs/", http.StatusOK, "ui"},
		{"spec file", "/docs/swagger.yaml", http.StatusOK, "swagger"},
		{"missing file", "/docs/nope.yaml", http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}
