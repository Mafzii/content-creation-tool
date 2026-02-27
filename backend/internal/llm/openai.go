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

type openaiClient struct {
	model  string
	apiKey string
}

type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message openaiMessage `json:"message"`
	} `json:"choices"`
}

func (c *openaiClient) call(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(openaiRequest{
		Model:    c.model,
		Messages: []openaiMessage{{Role: "user", Content: prompt}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai API %d: %s", resp.StatusCode, raw)
	}

	var or openaiResponse
	if err := json.Unmarshal(raw, &or); err != nil {
		return "", err
	}
	if len(or.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}
	return or.Choices[0].Message.Content, nil
}

func (c *openaiClient) Generate(ctx context.Context, prompt string) ([]string, error) {
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

func (c *openaiClient) Tweak(ctx context.Context, content, instruction string) (string, error) {
	prompt := fmt.Sprintf("Here is a draft:\n\n%s\n\nApply the following instruction and return only the revised draft:\n%s", content, instruction)
	return c.call(ctx, prompt)
}
