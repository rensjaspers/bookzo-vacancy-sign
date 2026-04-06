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
	setHeaders(request, "abc")
	assertHeader(t, request, "Origin", fallbackOrigin)
	assertHeader(t, request, "Referer", fallbackOrigin+"/")
	assertHeader(t, request, "X-Client-Name", clientName)
	assertHeader(t, request, "x-apikey", "abc")
}

func assertHeader(t *testing.T, request *http.Request, key string, want string) {
	t.Helper()
	if got := request.Header.Get(key); got != want {
		t.Fatalf("unexpected %s header: %q", key, got)
	}
}
