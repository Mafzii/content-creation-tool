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

type geminiClient struct {
	model  string
	apiKey string
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

func (c *geminiClient) call(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.model, c.apiKey,
	)
	body, _ := json.Marshal(geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API %d: %s", resp.StatusCode, raw)
	}

	var gr geminiResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return "", err
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content")
	}
	return gr.Candidates[0].Content.Parts[0].Text, nil
}

func (c *geminiClient) Generate(ctx context.Context, prompt string) ([]string, error) {
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

func (c *geminiClient) Tweak(ctx context.Context, content, instruction string) (string, error) {
	prompt := fmt.Sprintf("Here is a draft:\n\n%s\n\nApply the following instruction and return only the revised draft:\n%s", content, instruction)
	return c.call(ctx, prompt)
}
