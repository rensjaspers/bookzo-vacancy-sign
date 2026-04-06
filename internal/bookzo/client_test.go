package bookzo

import (
	"net/http"
	"testing"
)

func TestSetHeadersMatchesProxyStyleRequest(t *testing.T) {
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	setHeaders(request, "abc", "")
	assertHeader(t, request, "Origin", defaultOrigin)
	assertHeader(t, request, "Referer", defaultOrigin+"/")
	assertHeader(t, request, "X-Client-Name", clientName)
	assertHeader(t, request, "x-apikey", "abc")
}

func TestSetHeadersUsesConfiguredOrigin(t *testing.T) {
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	setHeaders(request, "abc", "https://hotel.example")
	assertHeader(t, request, "Origin", "https://hotel.example")
	assertHeader(t, request, "Referer", "https://hotel.example/")
}

func assertHeader(t *testing.T, request *http.Request, key string, want string) {
	t.Helper()
	if got := request.Header.Get(key); got != want {
		t.Fatalf("unexpected %s header: %q", key, got)
	}
}
