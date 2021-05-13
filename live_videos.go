package onfido

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

// LiveVideo represents a live video object in Onfido API
// https://documentation.onfido.com/#live-video-object
type LiveVideo struct {
	ID           string     `json:"id,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	Href         string     `json:"href,omitempty"`
	DownloadHref string     `json:"download_href,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	FileType     string     `json:"file_type,omitempty"`
	FileSize     int        `json:"file_size,omitempty"`
}

type LiveVideoDownload struct {
	// Data is the binary data of the video encoded as a Base64 string
	Data string
}

// DownloadLiveVideo returns the binary data representing the video.
// see https://documentation.onfido.com/#download-live-video
func (c *client) DownloadLiveVideo(ctx context.Context, id string) (*LiveVideoDownload, error) {
	req, err := c.newRequest(http.MethodGet, "/live_videos/"+id+"/download", nil)
	if err != nil {
		return nil, err
	}

	var resp bytes.Buffer
	_, err = c.do(ctx, req, &resp)

	var encodedBytes bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &encodedBytes)
	defer encoder.Close()

	_, err = encoder.Write(resp.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write to encoded byte stream: %w", err)
	}

	return &LiveVideoDownload{
		Data: encodedBytes.String(),
	}, err
}
