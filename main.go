package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

const (
	maxEmojis       = 50
	columnsPerRow   = 5
	defaultEmojiURL = "https://emojidb.org"
)

// EmojiClient represents our scraper client.
type EmojiClient struct {
	BaseURL string
}

// NewClient initializes a new EmojiDB client.
func NewClient() *EmojiClient {
	return &EmojiClient{
		BaseURL: defaultEmojiURL,
	}
}

func (c *EmojiClient) buildSearchURL(query string) string {
	formattedQuery := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(query)), " ", "-")
	return fmt.Sprintf("%s/%s-emojis?utm_source=user_search", c.BaseURL, url.PathEscape(formattedQuery))
}

// Search queries EmojiDB and returns a list of emojis.
func (c *EmojiClient) Search(query string) ([]string, error) {
	searchURL := c.buildSearchURL(query)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, searchURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch emojis: status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var emojis []string
	// Targeting the specific structure: .emoji-ctn > .emoji
	doc.Find(".emoji-ctn .emoji").Each(func(_ int, s *goquery.Selection) {
		if len(emojis) >= maxEmojis {
			return
		}
		emoji := strings.TrimSpace(s.Text())
		if emoji != "" && len(emoji) < 5 {
			emojis = append(emojis, emoji)
		}
	})
	return emojis, nil
}

func renderEmojiTable(emojis []string) {
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignCenter}},
			Row:    tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignLeft}},
		}),
	)

	table.Header([]string{"#", "EMOJIS", "EMOJIS", "EMOJIS", "EMOJIS", "EMOJIS"})

	var data [][]string
	for i := 0; i < len(emojis); i += columnsPerRow {
		end := i + columnsPerRow
		if end > len(emojis) {
			end = len(emojis)
		}

		rowItems := emojis[i:end]
		row := []string{color.HiBlackString("%02d", (i/columnsPerRow)+1)}
		row = append(row, rowItems...)

		// Fill empty slots for consistent table width
		for j := len(rowItems); j < columnsPerRow; j++ {
			row = append(row, "")
		}
		data = append(data, row)
	}

	_ = table.Bulk(data)
	_ = table.Render()
}

func main() {
	if len(os.Args) < 2 {
		color.Red("Error: Please provide a search term.")
		fmt.Println("Usage: go run main.go <search-term>")
		return
	}

	client := NewClient()
	query := strings.Join(os.Args[1:], " ")

	headerColor := color.New(color.FgCyan, color.Bold)
	headerColor.Printf("🔎 Searching EmojiDB for: '%s'...\n", query)

	emojis, err := client.Search(query)
	if err != nil {
		color.Red("❌ Error: %v", err)
		return
	}

	if len(emojis) == 0 {
		color.Yellow("⚠️  No emojis found for '%s'.", query)
		return
	}

	fmt.Println()
	renderEmojiTable(emojis)
	color.HiGreen("\n✅ Done! Found %d emojis.", len(emojis))
}
