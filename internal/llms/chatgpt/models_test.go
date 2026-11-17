package chatgpt

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFetchModelsFiltersListed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("client_version") == "" {
			t.Errorf("missing client_version")
		}
		if r.Header.Get("Authorization") != "Bearer tok" || r.Header.Get("chatgpt-account-id") != "acc_1" {
			t.Errorf("headers = %v", r.Header)
		}
		_, _ = io.WriteString(w, `{"models":[
			{"slug":"gpt-a","visibility":"list"},
			{"slug":"gpt-hidden","visibility":"hide"},
			{"slug":"","visibility":"list"},
			{"slug":"gpt-b","visibility":"list"}]}`)
	}))
	defer srv.Close()
	got, err := FetchModels(context.Background(), &fakeTokens{access: "tok"}, WithModelsURL(srv.URL), WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("FetchModels: %v", err)
	}
	if want := []string{"gpt-a", "gpt-b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("models = %v, want %v", got, want)
	}
}

func TestFetchModelsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	if _, err := FetchModels(context.Background(), &fakeTokens{access: "tok"}, WithModelsURL(srv.URL)); err == nil {
		t.Fatal("want error on 403")
	}
}
