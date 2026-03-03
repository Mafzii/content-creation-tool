package scraper

import (
	"fmt"
	"net/http"
	"time"

	readability "github.com/go-shiori/go-readability"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// FetchURL fetches the given URL and extracts the main article text using go-readability.
func FetchURL(url string) (title string, content string, err error) {
	article, err := readability.FromURL(url, 30*time.Second)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch URL: %w", err)
	}

	title = article.Title
	content = article.TextContent
	if content == "" {
		content = article.Content
	}
	return title, content, nil
}
