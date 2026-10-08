package voice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BhashiniProvider implements STT and TTS using the Bhashini (Dhruva) pipeline API.
type BhashiniProvider struct {
	APIKey  string
	UserID  string
	BaseURL string
	// Optional explicit service IDs; when empty, language is used to select.
	ASRServiceID string
	TTSServiceID string
	Client       *http.Client
}

// NewBhashiniProvider constructs a Bhashini provider.
func NewBhashiniProvider(apiKey, userID, baseURL string) *BhashiniProvider {
	if baseURL == "" {
		baseURL = "https://dhruva-api.bhashini.gov.in"
	}
	return &BhashiniProvider{
		APIKey:  apiKey,
		UserID:  userID,
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 45 * time.Second},
	}
}

type pipelineRequest struct {
	PipelineTasks []pipelineTask `json:"pipelineTasks"`
	InputData     map[string]any `json:"inputData"`
}

type pipelineTask struct {
	TaskType string         `json:"taskType"`
	Config   map[string]any `json:"config"`
}

type pipelineResponse struct {
	PipelineResponse []struct {
		TaskType string `json:"taskType"`
		Output   []struct {
			Source       string `json:"source"`
			Target       string `json:"target"`
			AudioContent string `json:"audioContent"`
		} `json:"output"`
	} `json:"pipelineResponse"`
}

func (p *BhashiniProvider) call(ctx context.Context, path string, req pipelineRequest) (*pipelineResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", p.APIKey)
	if p.UserID != "" {
		httpReq.Header.Set("userID", p.UserID)
	}

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<24))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bhashini status %d: %s", resp.StatusCode, string(respBody))
	}
	var out pipelineResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("decode bhashini response: %w", err)
	}
	return &out, nil
}

// Transcribe runs ASR through the Bhashini pipeline.
func (p *BhashiniProvider) Transcribe(ctx context.Context, audio []byte, lang, format string) (*TranscribeResult, error) {
	if lang == "" {
		lang = "hi"
	}
	req := pipelineRequest{
		PipelineTasks: []pipelineTask{{
			TaskType: "asr",
			Config: map[string]any{
				"language":     map[string]any{"sourceLanguage": lang},
				"serviceId":    p.ASRServiceID,
				"audioFormat":  normalizeFormat(format),
				"samplingRate": 16000,
			},
		}},
		InputData: map[string]any{
			"audio": []map[string]any{{"audioContent": base64.StdEncoding.EncodeToString(audio)}},
		},
	}
	out, err := p.call(ctx, "/services/inference/pipeline", req)
	if err != nil {
		return nil, err
	}
	for _, pr := range out.PipelineResponse {
		for _, o := range pr.Output {
			if o.Source != "" {
				return &TranscribeResult{Text: o.Source, Language: lang, Confidence: 0.9}, nil
			}
		}
	}
	return nil, fmt.Errorf("bhashini asr returned no transcript")
}

// Synthesize runs TTS through the Bhashini pipeline.
func (p *BhashiniProvider) Synthesize(ctx context.Context, text, lang string) (*SynthesizeResult, error) {
	if lang == "" {
		lang = "hi"
	}
	req := pipelineRequest{
		PipelineTasks: []pipelineTask{{
			TaskType: "tts",
			Config: map[string]any{
				"language":  map[string]any{"sourceLanguage": lang},
				"serviceId": p.TTSServiceID,
				"gender":    "female",
			},
		}},
		InputData: map[string]any{
			"input": []map[string]any{{"source": text}},
		},
	}
	out, err := p.call(ctx, "/services/inference/pipeline", req)
	if err != nil {
		return nil, err
	}
	for _, pr := range out.PipelineResponse {
		for _, o := range pr.Output {
			if o.AudioContent != "" {
				audio, err := base64.StdEncoding.DecodeString(o.AudioContent)
				if err != nil {
					return nil, err
				}
				return &SynthesizeResult{Audio: audio, Format: "wav", SampleRate: 22050}, nil
			}
		}
	}
	return nil, fmt.Errorf("bhashini tts returned no audio")
}
