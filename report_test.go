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

func TestGetReport_NonOKResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, wErr := w.Write([]byte("{\"error\": \"things went bad\"}"))
		assert.NoError(t, wErr)
	}))
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	_, err := client.GetReport(context.Background(), "")
	if err == nil {
		t.Fatal("expected server to return non ok response, got successful response")
	}
}

func TestGetReport_ReportRetrieved_Clear(t *testing.T) {
	expected := Report{
		ID:        "ce62d838-56f8-4ea5-98be-e7166d1dc33d",
		Name:      ReportNameDocument,
		Status:    "complete",
		Result:    ReportResultClear,
		SubResult: ReportSubResultClear,
		Href:      "/v2/live_photos/7410A943-8F00-43D8-98DE-36A774196D86",
		Properties: Properties{
			"document_type":   "passport",
			"issuing_country": "GBR",
		},
		CheckID: "541d040b-89f8-444b-8921-16b1333bf1c6",
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/reports/{reportId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		assert.Equal(t, expected.ID, vars["reportId"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, wErr := w.Write(expectedJSON)
		assert.NoError(t, wErr)
	}).Methods("GET")
	srv := httptest.NewServer(m)
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	r, err := client.GetReport(context.Background(), expected.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected.ID, r.ID)
	assert.Equal(t, expected.Name, r.Name)
	assert.Equal(t, expected.Status, r.Status)
	assert.Equal(t, expected.Result, r.Result)
	assert.Equal(t, expected.SubResult, r.SubResult)
	assert.Equal(t, expected.Href, r.Href)
	assert.Equal(t, expected.CheckID, r.CheckID)
	assert.Zero(t, r.Breakdown)
	assert.NotZero(t, r.Properties)
	assert.Contains(t, r.Properties, "document_type")
	assert.Equal(t, "passport", r.Properties["document_type"])
	assert.Contains(t, r.Properties, "issuing_country")
	assert.Equal(t, "GBR", r.Properties["issuing_country"])
}

func TestGetReport_ReportRetrieved_Consider(t *testing.T) {
	breakdownResultConsider := BreakdownResult(ReportResultConsider)
	breakdownSubResultConsider := BreakdownSubResult(ReportResultConsider)
	breakdownSubResultClear := BreakdownSubResult(ReportResultClear)
	expected := Report{
		ID:        "ce62d838-56f8-4ea5-98be-e7166d1dc33d",
		Name:      ReportNameDocument,
		Status:    "complete",
		Result:    ReportResultConsider,
		SubResult: ReportSubResultRejected,
		Href:      "/v2/live_photos/7410A943-8F00-43D8-98DE-36A774196D86",
		CheckID:   "541d040b-89f8-444b-8921-16b1333bf1c6",
		Breakdown: Breakdowns{
			"image_integrity": Breakdown{
				Result: &breakdownResultConsider,
				SubBreakdowns: SubBreakdowns{
					"image_quality": SubBreakdown{
						Result: &breakdownSubResultConsider,
						Properties: Properties{
							"alpha": "one",
							"beta":  "two",
						},
					},
					"supported_document": {
						Result: &breakdownSubResultClear,
					},
				},
			},
		},
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/reports/{reportId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		assert.Equal(t, expected.ID, vars["reportId"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, wErr := w.Write(expectedJSON)
		assert.NoError(t, wErr)
	}).Methods("GET")
	srv := httptest.NewServer(m)
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	r, err := client.GetReport(context.Background(), expected.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected.ID, r.ID)
	assert.Equal(t, expected.Name, r.Name)
	assert.Equal(t, expected.Status, r.Status)
	assert.Equal(t, expected.Result, r.Result)
	assert.Equal(t, expected.SubResult, r.SubResult)
	assert.Equal(t, expected.Href, r.Href)
	assert.Equal(t, expected.CheckID, r.CheckID)
	assert.Len(t, r.Breakdown, 1)
	assert.Contains(t, r.Breakdown, "image_integrity")
	assert.Contains(t, r.Breakdown["image_integrity"].SubBreakdowns, "image_quality")
	assert.Equal(t, breakdownSubResultConsider, *r.Breakdown["image_integrity"].SubBreakdowns["image_quality"].Result)
	assert.NotZero(t, r.Breakdown["image_integrity"].SubBreakdowns["image_quality"].Properties)
	assert.Contains(t, r.Breakdown["image_integrity"].SubBreakdowns["image_quality"].Properties, "alpha")
	assert.Contains(t, r.Breakdown["image_integrity"].SubBreakdowns["image_quality"].Properties["alpha"], "one")
	assert.Contains(t, r.Breakdown["image_integrity"].SubBreakdowns["image_quality"].Properties, "beta")
	assert.Contains(t, r.Breakdown["image_integrity"].SubBreakdowns["image_quality"].Properties["beta"], "two")
	assert.Contains(t, r.Breakdown["image_integrity"].SubBreakdowns, "supported_document")
	assert.Equal(t, breakdownSubResultClear, *r.Breakdown["image_integrity"].SubBreakdowns["supported_document"].Result)
	assert.Zero(t, r.Breakdown["image_integrity"].SubBreakdowns["supported_document"].Properties)
}

func TestResumeReport_NonOKResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, wErr := w.Write([]byte("{\"error\": \"things went bad\"}"))
		assert.NoError(t, wErr)
	}))
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	err := client.ResumeReport(context.Background(), "")
	if err == nil {
		t.Fatal("expected server to return non ok response, got successful response")
	}
}

func TestResumeReport_ReportResumed(t *testing.T) {
	reportID := "ce62d838-56f8-4ea5-98be-e7166d1dc33d"

	m := mux.NewRouter()
	m.HandleFunc("/reports/{reportId}/resume", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		assert.Equal(t, reportID, vars["reportId"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")
	srv := httptest.NewServer(m)
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	err := client.ResumeReport(context.Background(), reportID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCancelReport_NonOKResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, wErr := w.Write([]byte("{\"error\": \"things went bad\"}"))
		assert.NoError(t, wErr)
	}))
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	err := client.CancelReport(context.Background(), "")
	if err == nil {
		t.Fatal("expected server to return non ok response, got successful response")
	}
}

func TestCancelReport_ReportResumed(t *testing.T) {
	reportID := "ce62d838-56f8-4ea5-98be-e7166d1dc33d"

	m := mux.NewRouter()
	m.HandleFunc("/reports/{reportId}/cancel", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		assert.Equal(t, reportID, vars["reportId"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")
	srv := httptest.NewServer(m)
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	err := client.CancelReport(context.Background(), reportID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestListReports_NonOKResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, wErr := w.Write([]byte("{\"error\": \"things went bad\"}"))
		assert.NoError(t, wErr)
	}))
	defer srv.Close()

	client := NewClient("123").(*client)
	client.endpoint = srv.URL

	it := client.ListReports("")
	if it.Next(context.Background()) == true {
		t.Fatal("expected iterator not to return next item, got next item")
	}
	if it.Err() == nil {
		t.Fatal("expected iterator to return error message, got nil")
	}
}

func TestListReports_ReportsRetrieved(t *testing.T) {
	checkID := "541d040b-89f8-444b-8921-16b1333bf1c6"
	expected := Report{
		ID:        "ce62d838-56f8-4ea5-98be-e7166d1dc33d",
		Name:      ReportNameDocument,
		Status:    "complete",
		Result:    ReportResultClear,
		SubResult: ReportSubResultClear,
		Href:      "/v2/live_photos/7410A943-8F00-43D8-98DE-36A774196D86",
		CheckID:   checkID,
	}
	expectedJSON, err := json.Marshal(Reports{
		Reports: []*Report{&expected},
	})
	if err != nil {
		t.Fatal(err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("check_id") != checkID {
			t.Fatal("expected checkID id was not in the request")
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

	it := client.ListReports(checkID)
	for it.Next(context.Background()) {
		r := it.Report()

		assert.Equal(t, expected.ID, r.ID)
		assert.Equal(t, expected.Name, r.Name)
		assert.Equal(t, expected.Status, r.Status)
		assert.Equal(t, expected.Result, r.Result)
		assert.Equal(t, expected.SubResult, r.SubResult)
		assert.Equal(t, expected.Href, r.Href)
		assert.Equal(t, expected.CheckID, r.CheckID)
	}
	if it.Err() != nil {
		t.Fatal(it.Err())
	}
}
