package onfido

import (
	"context"
	"encoding/json"
	"time"
)

// Supported report names, results, subresults, and variants
const (
	ReportNameDocument                       ReportName = "document"
	ReportNameDocumentWithAddress            ReportName = "document_with_address_information"
	ReportNameDocumentWithDrivingLicense     ReportName = "document_with_driving_licence_information"
	ReportNameFacialSimilarityPhoto          ReportName = "facial_similarity_photo"
	ReportNameFacialSimilarityPhotoFullyAuto ReportName = "facial_similarity_photo_fully_auto"
	ReportNameFacialSimilarityVideo          ReportName = "facial_similarity_video"
	ReportNameKnownFaces                     ReportName = "known_faces"
	ReportNameIdentityEnhanced               ReportName = "identity_enhanced"
	ReportNameWatchlistEnhanced              ReportName = "watchlist_enhanced"
	ReportNameWatchlistStandard              ReportName = "watchlist_standard"
	ReportNameWatchlistPepsOnly              ReportName = "watchlist_peps_only"
	ReportNameWatchlistSanctionsOnly         ReportName = "watchlist_sanctions_only"
	ReportNameProofOfAddress                 ReportName = "proof_of_address"
	ReportNameRightToWork                    ReportName = "right_to_work"

	ReportResultClear        ReportResult = "clear"
	ReportResultConsider     ReportResult = "consider"
	ReportResultUnidentified ReportResult = "unidentified"

	ReportSubResultClear     ReportSubResult = "clear"
	ReportSubResultRejected  ReportSubResult = "rejected"
	ReportSubResultSuspected ReportSubResult = "suspected"
	ReportSubResultCaution   ReportSubResult = "caution"
)

// ReportName represents a report type name
type ReportName string

// ReportResult represents a report result
type ReportResult string

// ReportSubResult represents a report sub result
type ReportSubResult string

// DocumentProcessed contains metadata about the document that has been processed
type DocumentProcessed map[string]interface{}

// Report represents a report from the Onfido API
type Report struct {
	ID         string                 `json:"id,omitempty"`
	Name       ReportName             `json:"name,omitempty"`
	CreatedAt  *time.Time             `json:"created_at,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Result     ReportResult           `json:"result,omitempty"`
	SubResult  ReportSubResult        `json:"sub_result,omitempty"`
	Href       string                 `json:"href,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
	Breakdown  Breakdowns             `json:"breakdown,omitempty"`
	Properties Properties             `json:"properties,omitempty"`
	CheckID    string                 `json:"check_id,omitempty"`
	Documents  []DocumentProcessed    `json:"documents,omitempty"`
}

// Reports represents a list of reports from the Onfido API
type Reports struct {
	Reports []*Report `json:"reports"`
}

// GetReport retrieves a report for the provided check by its ID.
// see https://documentation.onfido.com/?shell#retrieve-report
func (c *client) GetReport(ctx context.Context, id string) (*Report, error) {
	req, err := c.newRequest("GET", "/reports/"+id, nil)
	if err != nil {
		return nil, err
	}

	var resp Report
	_, err = c.do(ctx, req, &resp)
	return &resp, err
}

// ResumeReport resumes a paused report by its ID.
// see https://documentation.onfido.com/?shell#resume-report
func (c *client) ResumeReport(ctx context.Context, id string) error {
	req, err := c.newRequest("POST", "/reports/"+id+"/resume", nil)
	if err != nil {
		return err
	}

	_, err = c.do(ctx, req, nil)
	return err
}

// CancelReport cancels a report by its ID.
// see https://documentation.onfido.com/?shell#cancel-report
func (c *client) CancelReport(ctx context.Context, id string) error {
	req, err := c.newRequest("POST", "/reports/"+id+"/cancel", nil)
	if err != nil {
		return err
	}

	_, err = c.do(ctx, req, nil)
	return err
}

// ReportIter represents a document iterator
type ReportIter struct {
	*iter
}

// Report returns the current item in the iterator as a Report.
func (i *ReportIter) Report() *Report {
	return i.Current().(*Report)
}

// ListReports retrieves the list of reports for the provided check.
// see https://documentation.onfido.com/?shell#list-reports
func (c *client) ListReports(checkID string) *ReportIter {
	handler := func(body []byte) ([]interface{}, error) {
		var r Reports
		if err := json.Unmarshal(body, &r); err != nil {
			return nil, err
		}

		values := make([]interface{}, len(r.Reports))
		for i, v := range r.Reports {
			values[i] = v
		}
		return values, nil
	}

	return &ReportIter{&iter{
		c:       c,
		nextURL: "/reports?check_id=" + checkID,
		handler: handler,
	}}
}
