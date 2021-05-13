package onfido

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestListLiveVideos(t *testing.T) {
	applicantID := "541d040b-89f8-444b-8921-16b1333bf1c6"
	createdAt := time.Now()

	expected := LiveVideo{
		ID:           "541d040b-89f8-444b-8921-16b1333bf1c7",
		CreatedAt:    &createdAt,
		Href:         "/v3.1/live_videos/7410A943-8F00-43D8-98DE-36A774196D86",
		DownloadHref: "https://com/videos/pdf/1234",
		FileName:     "something.mp4",
		FileSize:     1234,
		FileType:     "video/mp4",
	}
	expectedJSON, err := json.Marshal(struct {
		LiveVideos []*LiveVideo `json:"live_videos"`
	}{
		LiveVideos: []*LiveVideo{&expected},
	})
	if err != nil {
		t.Fatal(err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/live_videos", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("applicant_id") != applicantID {
			t.Fatal("expected applicant id was not in the request")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, wErr := w.Write(expectedJSON)
		assert.NoError(t, wErr)
	}).Methods("GET")
	srv := httptest.NewServer(m)
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	it := client.ListLiveVideos(applicantID)
	for it.Next(context.Background()) {
		c := it.LiveVideo()

		assert.Equal(t, expected.ID, c.ID)
		assert.True(t, expected.CreatedAt.Equal(*c.CreatedAt))
		assert.Equal(t, expected.Href, c.Href)
		assert.Equal(t, expected.DownloadHref, c.DownloadHref)
		assert.Equal(t, expected.FileName, c.FileName)
		assert.Equal(t, expected.FileSize, c.FileSize)
		assert.Equal(t, expected.FileType, c.FileType)
	}
	if it.Err() != nil {
		t.Fatal(it.Err())
	}
}
