package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
)

// ScrapeResult represents the result of a web scraping operation
type ScrapeResult struct {
	URL     string
	Content string
}

// Scraper handles web scraping operations
type ScraperTool struct {
	client      *http.Client
	id          string
	name        string
	description string
}

func (s ScraperTool) Call(ctx context.Context, input string) (string, error) {
	return s.ExecuteScrapeWebsite(input).Content, nil
}

func (s ScraperTool) Name() string {
	return s.name
}

func (s ScraperTool) Description() string {
	return s.description
}

// executeScrapeWebsite executes the scraping tool to extract content from a given URL.
// It marshals the tool call to extract the URL, validates the URL, and then uses the scraper to scrape the content.
// The function returns a ToolResult containing the tool name, content, or an error if the operation fails.
func (s ScraperTool) ExecuteScrapeWebsite(url string) llms.ToolCallResponse {
	// Validate URL is not empty
	if url == "" {
		return llms.ToolCallResponse{
			ToolCallID: s.id,
			Name:       s.name,
			Content:    "url is empty",
		}
	}

	// Validate URL is a valid URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Scrape the website
	scraper := NewScraperTool()
	result, err := scraper.ScrapeWebsite(url)
	if err != nil {
		return llms.ToolCallResponse{
			ToolCallID: s.id,
			Name:       s.name,
			Content:    fmt.Sprintf("failed to scrape %s: %v", url, err),
		}
	}

	return llms.ToolCallResponse{
		ToolCallID: s.id,
		Name:       s.name,
		Content:    fmt.Sprintf("Successfully scraped content from %s:\n\n%s", url, result.Content),
	}
}

// NewScraper creates a new scraper instance with default configuration
func NewScraperTool() *ScraperTool {
	return &ScraperTool{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		id:          uuid.New().String(),
		name:        "scrape_website",
		description: "Scrape content from a website URL to extract recipe information. Use this when you need to get recipe details from a specific website.",
	}
}

// ScrapeWebsite scrapes content from a given URL
func (s *ScraperTool) ScrapeWebsite(url string) (*ScrapeResult, error) {
	// Make HTTP request
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Remove script and style elements to get only visible text
	doc.Find("script, style, nav, footer, .ad, .advertisement, .sidebar, .menu, .navigation").Remove()

	// Extract all visible text content
	var content strings.Builder

	// Get the page title
	title := doc.Find("title").Text()
	if title != "" {
		content.WriteString("Page Title: " + title + "\n\n")
	}

	// Get all text content from the body
	bodyText := doc.Find("body").Text()

	// Clean up the text by removing excessive whitespace and formatting
	lines := strings.Split(bodyText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && len(line) > 2 {
			// Skip very short lines that are likely navigation or ads
			if len(line) > 5 {
				content.WriteString(line + "\n")
			}
		}
	}

	// Limit content length to avoid overwhelming the LLM
	scrapedText := content.String()
	if len(scrapedText) > 6000 {
		scrapedText = scrapedText[:6000] + "... (content truncated)"
	}

	return &ScrapeResult{
		URL:     url,
		Content: scrapedText,
	}, nil
}
