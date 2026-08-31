package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	SupportAiAPI = "https://api.5systems.ru/support-ai/v1"

	MaxResults = 5
)

type TipsInput struct {
	Query string `json:"query" jsonschema:"User's question"`
}

type Tip struct {
	Title    string `json:"title" jsonschema:"The tip's title"`
	FullText string `json:"full_text" jsonschema:"The full text of tips in markdown format"`
}

type TipsOutput struct {
	Tips []Tip `json:"tips" jsonschema:"Relevant tips, empty if none found"`
}

type ArticlesInput struct {
	Query string `json:"query" jsonschema:"User's question"`
}

type Article struct {
	Title string `json:"title" jsonschema:"The article's title"`
	// FullText string `json:"full_text" jsonschema:"The full text of the article in markdown format"`
	Url string `json:"url" jsonschema:"The article's URL"`
}

type ArticlesOutput struct {
	Articles []Article `json:"articles" jsonschema:"Relevant articles, empty if none found"`
}

func makeNWSRequest[T any](ctx context.Context, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s, url: %s", resp.StatusCode, string(body), url)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to decode response: %w, body res: %s", err, string(body))
	}

	return &result, nil
}

func formatTip(tip Tip) string {
	return fmt.Sprintf(`
Title: %s
Full text: %s
`, tip.Title, tip.FullText)
}

func formatArticle(article Article) string {
	return fmt.Sprintf(`
Title: %s
Url: %s
`, article.Title, article.Url)
}

// searchTips wraps the tips in an object: MCP requires structuredContent to be
// a JSON object, so "no tips" is an empty array under "tips".
func searchTips(ctx context.Context, req *mcp.CallToolRequest, input TipsInput) (
	*mcp.CallToolResult, TipsOutput, error,
) {

	query := input.Query

	params := url.Values{}
	params.Add("collection_name", "5s-tips")
	params.Add("query", query)

	tipsURL := fmt.Sprintf("%s/search?%s", SupportAiAPI, params.Encode())

	rawTips, err := makeNWSRequest[[]Tip](ctx, tipsURL)
	if err != nil {
		return nil, TipsOutput{}, fmt.Errorf("unable to fetch tips for query %s: %w", query, err)
	}

	// An empty result is an empty array, not an error.
	if rawTips == nil || len(*rawTips) == 0 {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No relevant tips for this question."}},
		}

		return result, TipsOutput{Tips: []Tip{}}, nil
	}

	allTips := *rawTips
	lenResList := min(len(allTips), MaxResults)

	tips := make([]Tip, 0, lenResList)
	formatted := make([]string, 0, lenResList)
	for _, tip := range allTips[:lenResList] {
		tips = append(tips, Tip{
			Title:    tip.Title,
			FullText: tip.FullText,
		})

		formatted = append(formatted, formatTip(tip))
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: strings.Join(formatted, "\n---\n")}},
	}

	return result, TipsOutput{Tips: tips}, nil
}

// searchArticles wraps the articles in an object: MCP requires structuredContent
// to be a JSON object, so "no articles" is an empty array under "articles".
func searchArticles(ctx context.Context, req *mcp.CallToolRequest, input ArticlesInput) (
	*mcp.CallToolResult, ArticlesOutput, error,
) {

	query := input.Query

	params := url.Values{}
	params.Add("collection_name", "5s-doc-ollama")
	params.Add("query", query)

	articlesURL := fmt.Sprintf("%s/search?%s", SupportAiAPI, params.Encode())

	rawArticles, err := makeNWSRequest[[]Article](ctx, articlesURL)
	if err != nil {
		return nil, ArticlesOutput{}, fmt.Errorf("unable to fetch articles for query %s: %w", query, err)
	}

	// An empty result is an empty array, not an error.
	if rawArticles == nil || len(*rawArticles) == 0 {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No relevant articles for this question."}},
		}

		return result, ArticlesOutput{Articles: []Article{}}, nil
	}

	allArticles := *rawArticles
	lenResList := min(len(allArticles), MaxResults)

	articles := make([]Article, 0, lenResList)
	formatted := make([]string, 0, lenResList)
	for _, article := range allArticles[:lenResList] {
		articles = append(articles, Article{
			Title: article.Title,
			Url:   article.Url,
		})

		formatted = append(formatted, formatArticle(article))
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: strings.Join(formatted, "\n---\n")}},
	}

	return result, ArticlesOutput{Articles: articles}, nil
}

func main() {
	// Create MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "support-ai-mcp",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Capabilities: &mcp.ServerCapabilities{},
	})

	// Add search_tips tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_tips",
		Description: "Search for relevant tips",
	}, searchTips)

	// Add search_articles tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_articles",
		Description: "Search for relevant articles",
	}, searchArticles)

	// Run server on stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
