package onfido

import (
	"context"
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

// DownloadLiveVideo returns the binary data representing the video.
// see https://documentation.onfido.com/#download-live-video
func (c *client) DownloadLiveVideo(ctx context.Context, id string) (*LiveVideo, error) {
	req, err := c.newRequest(http.MethodGet, "/live_videos/"+id+"/download", nil)
	if err != nil {
		return nil, err
	}

	var resp *LiveVideo
	_, err = c.do(ctx, req, resp)

	return resp, err
}
