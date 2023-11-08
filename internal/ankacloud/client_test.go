package ankacloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomHeaders(t *testing.T) {
	const (
		headerName = "fake-header"
		headerVal  = "fake-value"
		path       = "/fakepath"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headerName) != headerVal {
			t.Errorf("expected header %q to be %q, got %q", headerName, headerVal, r.Header.Get(headerName))
		}

		json.NewEncoder(w).Encode(response{Status: statusOK})
	}))
	defer server.Close()

	client, err := NewAPIClient(APIClientConfig{
		BaseURL: server.URL,
		CustomHttpHeaders: map[string]string{
			headerName: headerVal,
		},
	})
	if err != nil {
		t.Error(err)
	}

	if _, err = client.Get(context.Background(), path, nil); err != nil {
		t.Error(err)
	}

	if _, err = client.Post(context.Background(), path, nil); err != nil {
		t.Error(err)
	}

	if _, err = client.Delete(context.Background(), path, nil); err != nil {
		t.Error(err)
	}
}
