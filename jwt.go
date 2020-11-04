package onfido

import (
	"bytes"
	"context"
	"encoding/json"
)

// SdkToken represents the response for a request for a JWT token
type SdkToken struct {
	ApplicantID   string `json:"applicant_id,omitempty"`
	Referrer      string `json:"referrer,omitempty"`
	ApplicationID string `json:"application_id,omitempty"`
	Token         string `json:"token,omitempty"`
}

// NewSdkTokenWeb returns a JWT token to used by the Javascript SDK.
func (c *client) NewSdkTokenWeb(ctx context.Context, applicantID, referrer string) (*SdkToken, error) {
	return c.sdkTokenRequest(ctx, &SdkToken{
		ApplicantID: applicantID,
		Referrer:    referrer,
	})
}

// NewSdkTokenMobile returns a JWT token to used by the iOS and Android SDKs.
func (c *client) NewSdkTokenMobile(ctx context.Context, applicantID, applicationID string) (*SdkToken, error) {
	return c.sdkTokenRequest(ctx, &SdkToken{
		ApplicantID:   applicantID,
		ApplicationID: applicationID,
	})
}

func (c *client) sdkTokenRequest(ctx context.Context, t *SdkToken) (*SdkToken, error) {
	jsonStr, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest("POST", "/sdk_token", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	var resp SdkToken
	if _, err := c.do(ctx, req, &resp); err != nil {
		return nil, err
	}

	t.Token = resp.Token

	return t, err
}
