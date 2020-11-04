package onfido

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestCreateWebhook_NonOKResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, wErr := w.Write([]byte("{\"error\": \"things went bad\"}"))
		assert.NoError(t, wErr)
	}))
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	_, err := client.CreateWebhook(context.Background(), WebhookRefRequest{})
	if err == nil {
		t.Fatal("expected server to return non ok response, got successful response")
	}
}

func TestCreateWebhook_WebhookCreated(t *testing.T) {
	expected := WebhookRef{
		ID:           "fcb73186-0733-4f6f-9c57-d9d5ef979443",
		URL:          "https://webhookendpoint.url",
		Enabled:      true,
		Href:         "/v2/webhooks/fcb73186-0733-4f6f-9c57-d9d5ef979443",
		Token:        "ExampleToken",
		Environments: []WebhookEnvironment{WebhookEnvironmentSandbox, WebhookEnvironmentLive},
		Events:       []WebhookEvent{WebhookEventCheckStarted, WebhookEventCheckCompleted},
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, wErr := w.Write(expectedJSON)
		assert.NoError(t, wErr)
	}).Methods("POST")
	srv := httptest.NewServer(m)
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	wh, err := client.CreateWebhook(context.Background(), WebhookRefRequest{
		URL:          "https://webhookendpoint.url",
		Enabled:      true,
		Environments: []WebhookEnvironment{WebhookEnvironmentSandbox, WebhookEnvironmentLive},
		Events:       []WebhookEvent{WebhookEventCheckStarted, WebhookEventCheckCompleted},
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected.ID, wh.ID)
	assert.Equal(t, expected.URL, wh.URL)
	assert.Equal(t, expected.Href, wh.Href)
	assert.Equal(t, expected.Token, wh.Token)
	assert.Equal(t, expected.Enabled, wh.Enabled)
	assert.Equal(t, expected.Environments, wh.Environments)
	assert.Equal(t, expected.Events, wh.Events)
}

func TestListWebhooks_NonOKResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, wErr := w.Write([]byte("{\"error\": \"things went bad\"}"))
		assert.NoError(t, wErr)
	}))
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	it := client.ListWebhooks()
	if it.Next(context.Background()) == true {
		t.Fatal("expected iterator not to return next item, got next item")
	}
	if it.Err() == nil {
		t.Fatal("expected iterator to return error message, got nil")
	}
}

func TestListWebhooks_WebhooksRetrieved(t *testing.T) {
	expected := WebhookRef{
		ID:           "fcb73186-0733-4f6f-9c57-d9d5ef979443",
		URL:          "https://webhookendpoint.url",
		Enabled:      true,
		Href:         "/v2/webhooks/fcb73186-0733-4f6f-9c57-d9d5ef979443",
		Token:        "ExampleToken",
		Environments: []WebhookEnvironment{WebhookEnvironmentSandbox, WebhookEnvironmentLive},
		Events:       []WebhookEvent{WebhookEventCheckStarted, WebhookEventCheckCompleted},
	}
	expectedJSON, err := json.Marshal(WebhookRefs{
		WebhookRefs: []*WebhookRef{&expected},
	})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, wErr := w.Write(expectedJSON)
		assert.NoError(t, wErr)
	}))
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	it := client.ListWebhooks()
	for it.Next(context.Background()) {
		wh := it.WebhookRef()

		assert.Equal(t, expected.ID, wh.ID)
		assert.Equal(t, expected.URL, wh.URL)
		assert.Equal(t, expected.Href, wh.Href)
		assert.Equal(t, expected.Token, wh.Token)
		assert.Equal(t, expected.Enabled, wh.Enabled)
		assert.Equal(t, expected.Environments, wh.Environments)
		assert.Equal(t, expected.Events, wh.Events)
	}
	if it.Err() != nil {
		t.Fatal(it.Err())
	}
}
