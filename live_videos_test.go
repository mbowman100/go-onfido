package onfido

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadLiveVideo(t *testing.T) {
	mockVideoID := "93672a37-8223-48b9-a440-3b5cb52a8e4b"
	m := mux.NewRouter()
	m.HandleFunc("/live_videos/{videoId}/download", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		assert.Equal(t, mockVideoID, vars["videoId"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, wErr := w.Write([]byte("this is a video"))
		assert.NoError(t, wErr)
	}).Methods("GET")
	srv := httptest.NewServer(m)
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	videoDownload, err := client.DownloadLiveVideo(context.Background(), mockVideoID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "dGhpcyBpcyBhIHZpZGVv", videoDownload.Data)
}