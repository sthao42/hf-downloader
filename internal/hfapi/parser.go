package hfapi

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Regex patterns for matching Hugging Face URLs
var (
	// Matches: (http(s)://)?(www\.)?huggingface.co/(models/|datasets/)?([^/?#]+/[^/?#]+|[^/?#]+)
	// along with optional /blob/, /resolve/, /raw/, or /tree/
	actionRegex = regexp.MustCompile(`^(?:https?://)?(?:www\.)?huggingface\.co/(?:models/)?([^/?#]+(?:/[^/?#]+)?)/(blob|resolve|raw|tree)/([^/?#]+)(?:/(.+))?$`)
	repoRegex   = regexp.MustCompile(`^(?:https?://)?(?:www\.)?huggingface\.co/(?:models/)?([^/?#]+/[^/?#]+|[^/?#]+)/?$`)
	shortRegex  = regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+/[a-zA-Z0-9_\-\.]+)$`)
)

// ParseHFURL analyzes an input string and returns a structured ParsedTarget.
func ParseHFURL(input string) (*ParsedTarget, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, fmt.Errorf("input URL or repo ID cannot be empty")
	}

	// Remove trailing query params and hashes for parsing
	cleanInput := trimmed
	if idx := strings.IndexAny(cleanInput, "?#"); idx != -1 {
		cleanInput = cleanInput[:idx]
	}
	cleanInput = strings.TrimRight(cleanInput, "/")

	// 1. Check for action URLs (/blob/, /resolve/, /raw/, /tree/)
	if matches := actionRegex.FindStringSubmatch(cleanInput); len(matches) > 0 {
		repoID := matches[1]
		action := matches[2]
		revision := matches[3]
		subpath := ""
		if len(matches) > 4 {
			subpath = matches[4]
		}

		if revision == "" {
			revision = "main"
		}

		var targetType URLType
		switch action {
		case "blob", "resolve", "raw":
			targetType = TargetFile
		case "tree":
			if subpath == "" {
				targetType = TargetRepo
			} else {
				targetType = TargetFolder
			}
		}

		// URL-decode subpath
		if decoded, err := url.PathUnescape(subpath); err == nil {
			subpath = decoded
		}

		return &ParsedTarget{
			RawInput: trimmed,
			RepoID:   repoID,
			Revision: revision,
			Type:     targetType,
			Subpath:  subpath,
		}, nil
	}

	// 2. Check for full repo URL: huggingface.co/org/repo
	if matches := repoRegex.FindStringSubmatch(cleanInput); len(matches) > 0 {
		repoID := matches[1]
		return &ParsedTarget{
			RawInput: trimmed,
			RepoID:   repoID,
			Revision: "main",
			Type:     TargetRepo,
			Subpath:  "",
		}, nil
	}

	// 3. Check for short repo identifier: org/repo
	if matches := shortRegex.FindStringSubmatch(cleanInput); len(matches) > 0 {
		repoID := matches[1]
		return &ParsedTarget{
			RawInput: trimmed,
			RepoID:   repoID,
			Revision: "main",
			Type:     TargetRepo,
			Subpath:  "",
		}, nil
	}

	return nil, fmt.Errorf("unable to parse '%s' as a valid Hugging Face URL or repository ID", trimmed)
}

// BuildDownloadURL returns the direct file download URL for a file in a repo.
func BuildDownloadURL(repoID, revision, filePath string) string {
	if revision == "" {
		revision = "main"
	}
	// Clean leading slashes
	filePath = strings.TrimLeft(filePath, "/")
	return fmt.Sprintf("https://huggingface.co/%s/resolve/%s/%s", repoID, revision, filePath)
}

// BuildTreeAPIURL constructs the Hugging Face API URL to fetch the file tree.
func BuildTreeAPIURL(repoID, revision, subpath string, recursive bool) string {
	if revision == "" {
		revision = "main"
	}
	subpath = strings.Trim(subpath, "/")
	var base string
	if subpath == "" {
		base = fmt.Sprintf("https://huggingface.co/api/models/%s/tree/%s", repoID, revision)
	} else {
		base = fmt.Sprintf("https://huggingface.co/api/models/%s/tree/%s/%s", repoID, revision, subpath)
	}
	if recursive {
		return base + "?recursive=true"
	}
	return base
}
