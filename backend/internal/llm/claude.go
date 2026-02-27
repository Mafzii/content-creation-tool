package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type claudeClient struct {
	model  string
	apiKey string
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (c *claudeClient) call(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(claudeRequest{
		Model:     c.model,
		MaxTokens: 4096,
		Messages:  []claudeMessage{{Role: "user", Content: prompt}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude API %d: %s", resp.StatusCode, raw)
	}

	var cr claudeResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return "", err
	}
	if len(cr.Content) == 0 {
		return "", fmt.Errorf("claude returned no content")
	}
	return cr.Content[0].Text, nil
}

func (c *claudeClient) Generate(ctx context.Context, prompt string) ([]string, error) {
	variants := make([]string, 3)
	errs := make([]error, 3)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			variants[idx], errs[idx] = c.call(ctx, prompt)
		}(i)
	}
	wg.Wait()
	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}
	return variants, nil
}

func (c *claudeClient) Tweak(ctx context.Context, content, instruction string) (string, error) {
	prompt := fmt.Sprintf("Here is a draft:\n\n%s\n\nApply the following instruction and return only the revised draft:\n%s", content, instruction)
	return c.call(ctx, prompt)
}
