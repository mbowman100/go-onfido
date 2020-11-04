package onfido

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestLivePhotos_List(t *testing.T) {
	applicantID := "541d040b-89f8-444b-8921-16b1333bf1c6"
	createdAt := time.Now()

	expected := LivePhoto{
		ID:           "541d040b-89f8-444b-8921-16b1333bf1c7",
		CreatedAt:    &createdAt,
		Href:         "/v2/live_photos/7410A943-8F00-43D8-98DE-36A774196D86",
		DownloadHref: "https://com/photo/pdf/1234",
		FileName:     "something.png",
		FileSize:     1234,
		FileType:     "image/png",
	}
	expectedJSON, err := json.Marshal(struct {
		LivePhotos []*LivePhoto `json:"live_photos"`
	}{
		LivePhotos: []*LivePhoto{&expected},
	})
	if err != nil {
		t.Fatal(err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/live_photos", func(w http.ResponseWriter, r *http.Request) {
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

	it := client.ListLivePhotos(applicantID)
	for it.Next(context.Background()) {
		c := it.LivePhoto()

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
