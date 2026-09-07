package hfapi

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var contentRangeRegex = regexp.MustCompile(`bytes\s+\d+-\d+/(\d+)`)

// Resolver handles inspecting remote file capabilities (Content-Length, Accept-Ranges, Redirects).
type Resolver struct {
	httpClient *http.Client
}

// NewResolver creates a new file capability resolver.
func NewResolver(httpClient *http.Client) *Resolver {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
			// Follow redirects automatically
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		}
	}
	return &Resolver{httpClient: httpClient}
}

// InspectFile checks whether a remote URL supports byte ranges, its total size, and auth requirements.
func (r *Resolver) InspectFile(downloadURL, token string) (*InspectionResult, error) {
	result := &InspectionResult{
		URL: downloadURL,
	}

	// First attempt: HEAD request
	req, err := http.NewRequest("HEAD", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HEAD request: %w", err)
	}
	req.Header.Set("User-Agent", "hf-downloader/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := r.httpClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		result.StatusCode = resp.StatusCode
		result.FinalURL = resp.Request.URL.String()
		result.ETag = resp.Header.Get("ETag")

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			result.IsGated = true
			return result, nil
		}

		if resp.StatusCode == http.StatusOK {
			result.ContentLength = resp.ContentLength
			if strings.EqualFold(resp.Header.Get("Accept-Ranges"), "bytes") {
				result.AcceptRanges = true
			}
		}
	}

	// If HEAD didn't confirm range support or content length was 0/unknown, probe with Range: bytes=0-0
	if !result.AcceptRanges || result.ContentLength <= 0 {
		rangeReq, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create range probe request: %w", err)
		}
		rangeReq.Header.Set("User-Agent", "hf-downloader/1.0")
		rangeReq.Header.Set("Range", "bytes=0-0")
		if token != "" {
			rangeReq.Header.Set("Authorization", "Bearer "+token)
		}

		rangeResp, err := r.httpClient.Do(rangeReq)
		if err != nil {
			// If HEAD succeeded with a positive length, return that; otherwise return error
			if result.ContentLength > 0 {
				return result, nil
			}
			return nil, fmt.Errorf("failed to probe byte ranges: %w", err)
		}
		defer rangeResp.Body.Close()

		result.StatusCode = rangeResp.StatusCode
		result.FinalURL = rangeResp.Request.URL.String()
		if result.ETag == "" {
			result.ETag = rangeResp.Header.Get("ETag")
		}

		if rangeResp.StatusCode == http.StatusUnauthorized || rangeResp.StatusCode == http.StatusForbidden {
			result.IsGated = true
			return result, nil
		}

		if rangeResp.StatusCode == http.StatusPartialContent {
			result.AcceptRanges = true
			cr := rangeResp.Header.Get("Content-Range")
			if match := contentRangeRegex.FindStringSubmatch(cr); len(match) > 1 {
				if total, parseErr := strconv.ParseInt(match[1], 10, 64); parseErr == nil {
					result.ContentLength = total
				}
			}
		} else if rangeResp.StatusCode == http.StatusOK {
			// Server ignored Range header and sent the whole file or 200 OK
			result.AcceptRanges = false
			result.ContentLength = rangeResp.ContentLength
		}
	}

	return result, nil
}
