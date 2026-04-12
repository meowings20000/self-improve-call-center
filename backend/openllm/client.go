package openllm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	OpenLLM "github.com/golang-io/OpenLLM"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// Client wraps the OpenAI SDK client with proper timeout configuration
type Client struct {
	client  openai.Client
	BaseURL string
	APIKey  string
	Model   string
}

// Message represents a chat message (alias to OpenLLM.Message)
type Message = OpenLLM.Message

// MessageRole type alias for message role
type MessageRole = OpenLLM.MessageRole

// Role constants for message roles
var (
	RoleSystem    = OpenLLM.RoleSystem
	RoleUser      = OpenLLM.RoleUser
	RoleAssistant = OpenLLM.RoleAssistant
	RoleTool      = OpenLLM.RoleTool
)

// ChatResponse represents a simplified chat response
type ChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content string `json:"content"`
	Usage   struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
		TotalTokens      int64 `json:"total_tokens"`
	} `json:"usage"`
}

// NewClient creates a new OpenAI client with proper HTTP timeout (180s)
// We bypass OpenLLM's CreateOpenAI because it has a bug:
// requests.New() uses default 30s timeout and HTTPClientOptions only affects RoundTripper, not http.Client.Timeout
func NewClient(baseURL string, apiKey string, model string) *Client {
	httpClient := &http.Client{
		Timeout: 600 * time.Second, // 10 minutes - LLM calls can take very long for Chinese content
	}

	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(httpClient),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := openai.NewClient(opts...)

	return &Client{
		client:  client,
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	}
}

// Chat sends a chat completion request to the LLM
func (c *Client) Chat(model string, messages []Message) (*ChatResponse, error) {
	effectiveModel := model
	if effectiveModel == "" {
		effectiveModel = c.Model
	}

	// Convert messages to OpenAI SDK format
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case OpenLLM.RoleSystem:
			openaiMessages[i] = openai.SystemMessage(msg.Content)
		case OpenLLM.RoleUser:
			openaiMessages[i] = openai.UserMessage(msg.Content)
		case OpenLLM.RoleAssistant:
			openaiMessages[i] = openai.AssistantMessage(msg.Content)
		default:
			openaiMessages[i] = openai.UserMessage(msg.Content)
		}
	}

	params := openai.ChatCompletionNewParams{
		Messages: openaiMessages,
		Model:    effectiveModel,
	}

	// Context timeout as a safety net (HTTP client already has 600s)
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("LLM API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from LLM")
	}

	resp := &ChatResponse{
		ID:      completion.ID,
		Model:   effectiveModel,
		Content: completion.Choices[0].Message.Content,
	}
	resp.Usage.PromptTokens = int64(completion.Usage.PromptTokens)
	resp.Usage.CompletionTokens = int64(completion.Usage.CompletionTokens)
	resp.Usage.TotalTokens = int64(completion.Usage.PromptTokens + completion.Usage.CompletionTokens)

	return resp, nil
}

// SetBaseURL updates the base URL and recreates the client
func (c *Client) SetBaseURL(baseURL string) {
	c.BaseURL = baseURL
	c.recreateClient()
}

// SetAPIKey updates the API key and recreates the client
func (c *Client) SetAPIKey(apiKey string) {
	c.APIKey = apiKey
	c.recreateClient()
}

func (c *Client) recreateClient() {
	newClient := NewClient(c.BaseURL, c.APIKey, c.Model)
	c.client = newClient.client
}
