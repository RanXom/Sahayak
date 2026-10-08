package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BhashiniNMT implements translation via the Bhashini (Dhruva) pipeline API.
type BhashiniNMT struct {
	APIKey    string
	UserID    string
	BaseURL   string
	ServiceID string
	Client    *http.Client
}

// NewBhashiniNMT constructs a Bhashini NMT translator.
func NewBhashiniNMT(apiKey, userID, baseURL, serviceID string) *BhashiniNMT {
	if baseURL == "" {
		baseURL = "https://dhruva-api.bhashini.gov.in"
	}
	return &BhashiniNMT{
		APIKey:    apiKey,
		UserID:    userID,
		BaseURL:   baseURL,
		ServiceID: serviceID,
		Client:    &http.Client{Timeout: 30 * time.Second},
	}
}

type nmtRequest struct {
	PipelineTasks []struct {
		TaskType string         `json:"taskType"`
		Config   map[string]any `json:"config"`
	} `json:"pipelineTasks"`
	InputData map[string]any `json:"inputData"`
}

type nmtResponse struct {
	PipelineResponse []struct {
		Output []struct {
			Target string `json:"target"`
		} `json:"output"`
	} `json:"pipelineResponse"`
}

// Translate translates text using the Bhashini NMT pipeline.
func (t *BhashiniNMT) Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	if sourceLang == targetLang || text == "" {
		return text, nil
	}
	req := nmtRequest{
		InputData: map[string]any{"input": []map[string]any{{"source": text}}},
	}
	req.PipelineTasks = append(req.PipelineTasks, struct {
		TaskType string         `json:"taskType"`
		Config   map[string]any `json:"config"`
	}{
		TaskType: "translation",
		Config: map[string]any{
			"language":  map[string]any{"sourceLanguage": sourceLang, "targetLanguage": targetLang},
			"serviceId": t.ServiceID,
		},
	})

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.BaseURL+"/services/inference/pipeline", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", t.APIKey)
	if t.UserID != "" {
		httpReq.Header.Set("userID", t.UserID)
	}

	resp, err := t.Client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("bhashini nmt status %d: %s", resp.StatusCode, string(respBody))
	}
	var out nmtResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", err
	}
	for _, pr := range out.PipelineResponse {
		for _, o := range pr.Output {
			if o.Target != "" {
				return o.Target, nil
			}
		}
	}
	return text, nil
}
