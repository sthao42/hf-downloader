package hfapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	ErrGatedRepo    = errors.New("repository requires authorization or accepted agreement (gated/private)")
	ErrNotFound     = errors.New("repository, folder, or revision not found")
	linkHeaderRegex = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)
)

// Client coordinates requests to the Hugging Face REST API.
type Client struct {
	httpClient *http.Client
	token      string
	userAgent  string
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithToken sets the Hugging Face personal access token.
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = strings.TrimSpace(token)
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// NewClient initializes a new Hugging Face API client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "hf-downloader/1.0 (+https://github.com/leaanthony/wails)",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetToken updates the client token at runtime.
func (c *Client) SetToken(token string) {
	c.token = strings.TrimSpace(token)
}

// Token returns the currently configured token.
func (c *Client) Token() string {
	return c.token
}

func (c *Client) prepareRequest(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// FetchTree retrieves the file tree from Hugging Face for the given target,
// traversing pages if the result is paginated.
func (c *Client) FetchTree(target *ParsedTarget, recursive bool) ([]FileNode, error) {
	if target == nil {
		return nil, fmt.Errorf("target cannot be nil")
	}

	apiURL := BuildTreeAPIURL(target.RepoID, target.Revision, target.Subpath, recursive)
	var allFiles []FileNode

	nextURL := apiURL
	for nextURL != "" {
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		c.prepareRequest(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to query Hugging Face API: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("failed to read response body: %w", readErr)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			// Success, proceed to decode
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("%w (status %d): %s", ErrGatedRepo, resp.StatusCode, string(body))
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w: repository %s at %s not found", ErrNotFound, target.RepoID, target.Revision)
		default:
			return nil, fmt.Errorf("hugging face API returned unexpected status %d: %s", resp.StatusCode, string(body))
		}

		var pageItems []FileNode
		if err := json.Unmarshal(body, &pageItems); err != nil {
			return nil, fmt.Errorf("failed to decode Hugging Face tree response: %w", err)
		}

		for i := range pageItems {
			// Populate SHA256 and true size from LFS metadata if present
			if pageItems[i].LFS != nil {
				if pageItems[i].LFS.OID != "" {
					pageItems[i].SHA256 = pageItems[i].LFS.OID
				}
				if pageItems[i].LFS.Size > 0 {
					pageItems[i].Size = pageItems[i].LFS.Size
				}
			}
			pageItems[i].DownloadURL = BuildDownloadURL(target.RepoID, target.Revision, pageItems[i].Path)
			allFiles = append(allFiles, pageItems[i])
		}

		// Check pagination Link header: <url>; rel="next"
		linkHeader := resp.Header.Get("Link")
		if linkHeader != "" {
			match := linkHeaderRegex.FindStringSubmatch(linkHeader)
			if len(match) > 1 {
				nextURL = match[1]
				continue
			}
		}
		nextURL = ""
	}

	return allFiles, nil
}
