package vclip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"krillin-ai/log"
)

const (
	baseURL         = "https://api-tts.vclip.io/json-rpc"
	pollInterval    = 3 * time.Second
	maxPollAttempts = 120 // 6 minutes max
)

// Client implements the Ttser interface for VClip TTS
type Client struct {
	apiKey  string
	voiceID string
	speed   float64
}

// NewClient creates a new VClip TTS client
func NewClient(apiKey, voiceID string, speed float64) *Client {
	if speed <= 0 {
		speed = 1.0
	}
	return &Client{
		apiKey:  apiKey,
		voiceID: voiceID,
		speed:   speed,
	}
}

// jsonRPCRequest is the generic request body for VClip API
type jsonRPCRequest struct {
	Method string      `json:"method"`
	Input  interface{} `json:"input"`
}

// ttsLongTextInput is the input body for ttsLongText
type ttsLongTextInput struct {
	Text        string  `json:"text"`
	UserVoiceID string  `json:"userVoiceId"`
	Speed       float64 `json:"speed"`
}

// getExportStatusInput is the input body for getExportStatus
type getExportStatusInput struct {
	ProjectExportID string `json:"projectExportId"`
}

// ttsLongTextResponse is the response from ttsLongText
type ttsLongTextResponse struct {
	Result struct {
		ProjectExportID string `json:"projectExportId"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// exportStatusResponse is the response from getExportStatus
type exportStatusResponse struct {
	Result struct {
		State string `json:"state"`
		URL   string `json:"url"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// doPost sends a POST request to the VClip API and decodes the response
func (c *Client) doPost(payload interface{}, out interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("vclip: failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("vclip: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("vclip: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vclip: non-200 status %d: %s", resp.StatusCode, string(b))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("vclip: failed to decode response: %w", err)
	}
	return nil
}

// Text2Speech implements Ttser. It submits a TTS job, polls until complete,
// downloads the audio, and saves it to outputFile.
func (c *Client) Text2Speech(text, voice, outputFile string) error {
	// Override voice if caller provides one; otherwise use the client default
	voiceID := c.voiceID
	if voice != "" {
		voiceID = voice
	}
	if voiceID == "" {
		return fmt.Errorf("vclip: voice ID is required")
	}

	log.GetLogger().Info("VClip TTS: submitting job",
		zap.String("voiceId", voiceID),
		zap.Float64("speed", c.speed),
		zap.Int("textLen", len(text)))

	// Step 1: Submit the TTS job
	req := jsonRPCRequest{
		Method: "ttsLongText",
		Input: ttsLongTextInput{
			Text:        text,
			UserVoiceID: voiceID,
			Speed:       c.speed,
		},
	}
	var ttsResp ttsLongTextResponse
	if err := c.doPost(req, &ttsResp); err != nil {
		return err
	}
	if ttsResp.Error != nil {
		return fmt.Errorf("vclip: API error %d: %s", ttsResp.Error.Code, ttsResp.Error.Message)
	}

	exportID := ttsResp.Result.ProjectExportID
	if exportID == "" {
		return fmt.Errorf("vclip: empty projectExportId in response")
	}
	log.GetLogger().Info("VClip TTS: job submitted", zap.String("exportId", exportID))

	// Step 2: Poll until completed
	statusReq := jsonRPCRequest{
		Method: "getExportStatus",
		Input:  getExportStatusInput{ProjectExportID: exportID},
	}

	var audioURL string
	for attempt := 0; attempt < maxPollAttempts; attempt++ {
		time.Sleep(pollInterval)

		var statusResp exportStatusResponse
		if err := c.doPost(statusReq, &statusResp); err != nil {
			log.GetLogger().Warn("VClip TTS: polling error", zap.Error(err))
			continue
		}
		if statusResp.Error != nil {
			return fmt.Errorf("vclip: status API error %d: %s", statusResp.Error.Code, statusResp.Error.Message)
		}

		state := statusResp.Result.State
		log.GetLogger().Info("VClip TTS: polling status",
			zap.String("exportId", exportID),
			zap.String("state", state),
			zap.Int("attempt", attempt+1))

		if state == "completed" {
			audioURL = statusResp.Result.URL
			break
		}
		if state == "failed" || state == "error" {
			return fmt.Errorf("vclip: job failed with state %q", state)
		}
	}

	if audioURL == "" {
		return fmt.Errorf("vclip: timed out waiting for job %s to complete", exportID)
	}

	// Step 3: Download the audio file
	log.GetLogger().Info("VClip TTS: downloading audio", zap.String("url", audioURL))
	dlResp, err := http.Get(audioURL)
	if err != nil {
		return fmt.Errorf("vclip: failed to download audio: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		return fmt.Errorf("vclip: download returned status %d", dlResp.StatusCode)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("vclip: failed to create output file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, dlResp.Body); err != nil {
		return fmt.Errorf("vclip: failed to write audio to file: %w", err)
	}

	log.GetLogger().Info("VClip TTS: audio saved", zap.String("outputFile", outputFile))
	return nil
}
