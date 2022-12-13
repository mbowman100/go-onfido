package onfido

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tomnomnom/linkheader"
)

// Constants
const (
	ClientVersion   = "0.1.0"
	DefaultEndpoint = "https://api.eu.onfido.com/v3.5"
	TokenEnv        = "ONFIDO_TOKEN"
)

type OnfidoClient interface {
	SetHTTPClient(client HTTPRequester)
	NewSdkTokenWeb(ctx context.Context, applicantID, referrer string) (*SdkToken, error)
	NewSdkTokenMobile(ctx context.Context, applicantID, applicationID string) (*SdkToken, error)
	GetReport(ctx context.Context, id string) (*Report, error)
	ResumeReport(ctx context.Context, id string) error
	CancelReport(ctx context.Context, id string) error
	ListReports(checkID string) *ReportIter
	GetDocument(ctx context.Context, id string) (*Document, error)
	ListDocuments(applicantID string) *DocumentIter
	UploadDocument(ctx context.Context, dr DocumentRequest) (*Document, error)
	DownloadDocument(ctx context.Context, id string) (*DocumentDownload, error)
	ListLivePhotos(applicantID string) *LivePhotoIter
	DownloadLiveVideo(ctx context.Context, id string) (*LiveVideoDownload, error)
	ListLiveVideos(applicantID string) LiveVideoIter
	CreateApplicant(ctx context.Context, a Applicant) (*Applicant, error)
	DeleteApplicant(ctx context.Context, id string) error
	GetApplicant(ctx context.Context, id string) (*Applicant, error)
	ListApplicants() *ApplicantIter
	UpdateApplicant(ctx context.Context, a Applicant) (*Applicant, error)
	CreateCheck(ctx context.Context, cr CheckRequest) (*Check, error)
	GetCheck(ctx context.Context, id string) (*CheckRetrieved, error)
	GetCheckExpanded(ctx context.Context, id string) (*Check, error)
	ResumeCheck(ctx context.Context, id string) (*Check, error)
	ListChecks(applicantID string) *CheckIter
	CreateWebhook(ctx context.Context, wr WebhookRefRequest) (*WebhookRef, error)
	UpdateWebhook(ctx context.Context, id string, wr WebhookRefRequest) (*WebhookRef, error)
	DeleteWebhook(ctx context.Context, id string) error
	ListWebhooks() *WebhookRefIter
	PickAddresses(postcode string) *PickerIter
	GetResource(ctx context.Context, href string, v interface{}) error
	Token() Token
}

// Client represents an Onfido API client
type client struct {
	endpoint   string
	httpClient HTTPRequester
	token      Token
}

func (c *client) SetHTTPClient(client HTTPRequester) {
	c.httpClient = client
}

var _ OnfidoClient = &client{}

// HTTPRequester represents an HTTP requester
type HTTPRequester interface {
	Do(*http.Request) (*http.Response, error)
}

// Error represents an Onfido API error response
type Error struct {
	Resp *http.Response
	// see https://documentation.onfido.com/#error-object
	Err struct {
		ID     string      `json:"id"`
		Type   string      `json:"type"`
		Msg    string      `json:"message"`
		Fields ErrorFields `json:"fields"`
	} `json:"error"`
}

// known shapes of the values are []string and map[string][]string for recursive field validation
type ErrorFields map[string]interface{}

func (e *Error) Error() string {
	if e.Err.Msg != "" {
		return e.Err.Msg
	}
	if e.Resp != nil {
		return fmt.Sprintf("http request failed with status code %d", e.Resp.StatusCode)
	}
	return "an unknown error occurred"
}

// Token is an Onfido authentication token
type Token string

// String returns the token as a string.
func (t Token) String() string {
	return string(t)
}

// Prod checks if this is a production token or not.
func (t Token) Prod() bool {
	return !strings.HasPrefix(string(t), "test_") &&
		!strings.HasPrefix(string(t), "api_sandbox.")
}

func (c *client) Token() Token { return c.token }

// NewClientFromEnv creates a new Onfido client using configuration
// from environment variables.
func NewClientFromEnv() (OnfidoClient, error) {
	token := os.Getenv(TokenEnv)
	if token == "" {
		return nil, fmt.Errorf("onfido token not found in environmental variable `%s`", TokenEnv)
	}
	return NewClient(token), nil
}

// NewClient creates a new Onfido client.
func NewClient(token string) OnfidoClient {
	return &client{
		endpoint:   DefaultEndpoint,
		httpClient: http.DefaultClient,
		token:      Token(token),
	}
}

func (c *client) newRequest(method, uri string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(uri, "http") {
		if !strings.HasPrefix(uri, "/") {
			uri = "/" + uri
		}
		uri = c.endpoint + uri
	}

	// Add in query params if they are present
	var q url.Values
	splitUri := strings.Split(uri, "?")
	if len(splitUri) == 2 {
		uri = splitUri[0]

		var err error
		q, err = url.ParseQuery(splitUri[1])
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go-Onfido/"+ClientVersion)
	req.Header.Set("Authorization", "Token token="+c.token.String())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *client) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, err
		}
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if c := resp.StatusCode; c < 200 || c > 299 {
		return nil, handleResponseErr(resp)
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else if isJSONResponse(resp) {
			err = json.NewDecoder(resp.Body).Decode(v)
		} else {
			err = fmt.Errorf("unable to parse respose body into %T", v)
		}
	}

	return resp, err
}

func isJSONResponse(resp *http.Response) bool {
	return strings.Contains(resp.Header.Get("Content-Type"), "application/json")
}

func handleResponseErr(resp *http.Response) error {
	var onfidoErr Error
	if resp.Body != nil && isJSONResponse(resp) {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&onfidoErr); err != nil {
			return err
		}
	} else {
		onfidoErr = Error{}
	}
	onfidoErr.Resp = resp
	return &onfidoErr
}

type Iter interface {
	Current() interface{}
	Err() error
	Next(ctx context.Context) bool
}

type iter struct {
	c       *client
	nextURL string
	handler iterHandler

	values []interface{}
	cur    interface{}
	err    error
}

type iterHandler func(body []byte) ([]interface{}, error)

func (it *iter) Current() interface{} {
	return it.cur
}

func (it *iter) Err() error {
	return it.err
}

func (it *iter) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}
	if len(it.values) == 0 && it.nextURL != "" {
		req, err := it.c.newRequest("GET", it.nextURL, nil)
		if err != nil {
			it.err = err
			return false
		}

		var body bytes.Buffer
		resp, err := it.c.do(ctx, req, &body)
		if err != nil {
			it.err = err
			return false
		}
		if !isJSONResponse(resp) {
			it.err = errors.New("non json response")
			return false
		}

		values, err := it.handler(body.Bytes())
		if err != nil {
			it.err = err
			return false
		}
		it.values = values

		links := linkheader.Parse(resp.Header.Get("Link"))
		links = links.FilterByRel("next")
		if len(links) > 0 {
			it.nextURL = links[0].URL
		} else {
			it.nextURL = ""
		}
	}
	if len(it.values) == 0 {
		return false
	}

	it.cur = it.values[0]
	it.values = it.values[1:]
	return true
}
