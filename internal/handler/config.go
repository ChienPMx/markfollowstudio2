package handler

import (
	"markflow-studio/config"
	"markflow-studio/internal/response"
	"markflow-studio/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Global variable to mark if configuration needs re-initialization
var configUpdated bool

// ConfigRequest defines the configuration data structure from frontend
type ConfigRequest struct {
	App struct {
		SegmentDuration       int    `json:"segment_duration"`
		TranscribeParallelNum int    `json:"transcribe_parallel_num"`
		TranslateParallelNum  int    `json:"translate_parallel_num"`
		TranscribeMaxAttempts int    `json:"transcribe_max_attempts"`
		TranslateMaxAttempts  int    `json:"translate_max_attempts"`
		MaxSentenceLength     int    `json:"max_sentence_length"`
		Proxy                 string `json:"proxy"`
	} `json:"app"`
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	Llm struct {
		BaseUrl string `json:"base_url"`
		ApiKey  string `json:"api_key"`
		Model   string `json:"model"`
	} `json:"llm"`
	Transcribe struct {
		Provider              string `json:"provider"`
		EnableGpuAcceleration bool   `json:"enable_gpu_acceleration"`
		Openai                struct {
			BaseUrl string `json:"base_url"`
			ApiKey  string `json:"api_key"`
			Model   string `json:"model"`
		} `json:"openai"`
		Fasterwhisper struct {
			Model string `json:"model"`
		} `json:"fasterwhisper"`
		Whisperkit struct {
			Model string `json:"model"`
		} `json:"whisperkit"`
		Whispercpp struct {
			Model string `json:"model"`
		} `json:"whispercpp"`
	} `json:"transcribe"`
	Tts struct {
		Provider string `json:"provider"`
		Openai   struct {
			BaseUrl string `json:"base_url"`
			ApiKey  string `json:"api_key"`
			Model   string `json:"model"`
			Voice   string `json:"voice"`
		} `json:"openai"`

		VClip struct {
			ApiKey  string  `json:"api_key"`
			VoiceID string  `json:"voice_id"`
			Speed   float64 `json:"speed"`
		} `json:"vclip"`
		Voices []config.VoiceEntry `json:"voices"`
	} `json:"tts"`
}

// GetConfig retrieves the current configuration
func (h Handler) GetConfig(c *gin.Context) {
	log.GetLogger().Info("Getting configuration info")

	if configUpdated {
		log.GetLogger().Info("Detected config update, re-initializing service...")
	}

	// Convert configuration to the format needed by frontend
	configResponse := ConfigRequest{
		App: struct {
			SegmentDuration       int    `json:"segment_duration"`
			TranscribeParallelNum int    `json:"transcribe_parallel_num"`
			TranslateParallelNum  int    `json:"translate_parallel_num"`
			TranscribeMaxAttempts int    `json:"transcribe_max_attempts"`
			TranslateMaxAttempts  int    `json:"translate_max_attempts"`
			MaxSentenceLength     int    `json:"max_sentence_length"`
			Proxy                 string `json:"proxy"`
		}{
			SegmentDuration:       config.Conf.App.SegmentDuration,
			TranscribeParallelNum: config.Conf.App.TranscribeParallelNum,
			TranslateParallelNum:  config.Conf.App.TranslateParallelNum,
			TranscribeMaxAttempts: config.Conf.App.TranscribeMaxAttempts,
			TranslateMaxAttempts:  config.Conf.App.TranslateMaxAttempts,
			MaxSentenceLength:     config.Conf.App.MaxSentenceLength,
			Proxy:                 config.Conf.App.Proxy,
		},
		Server: struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		}{
			Host: config.Conf.Server.Host,
			Port: config.Conf.Server.Port,
		},
		Llm: struct {
			BaseUrl string `json:"base_url"`
			ApiKey  string `json:"api_key"`
			Model   string `json:"model"`
		}{
			BaseUrl: config.Conf.Llm.BaseUrl,
			ApiKey:  config.Conf.Llm.ApiKey,
			Model:   config.Conf.Llm.Model,
		},
	}

	// Transcribe configuration
	configResponse.Transcribe.Provider = config.Conf.Transcribe.Provider
	configResponse.Transcribe.EnableGpuAcceleration = config.Conf.Transcribe.EnableGpuAcceleration
	configResponse.Transcribe.Openai.BaseUrl = config.Conf.Transcribe.Openai.BaseUrl
	configResponse.Transcribe.Openai.ApiKey = config.Conf.Transcribe.Openai.ApiKey
	configResponse.Transcribe.Openai.Model = config.Conf.Transcribe.Openai.Model
	configResponse.Transcribe.Fasterwhisper.Model = config.Conf.Transcribe.Fasterwhisper.Model
	configResponse.Transcribe.Whisperkit.Model = config.Conf.Transcribe.Whisperkit.Model
	configResponse.Transcribe.Whispercpp.Model = config.Conf.Transcribe.Whispercpp.Model

	// TTS configuration
	configResponse.Tts.Provider = config.Conf.Tts.Provider
	configResponse.Tts.Openai.BaseUrl = config.Conf.Tts.Openai.BaseUrl
	configResponse.Tts.Openai.ApiKey = config.Conf.Tts.Openai.ApiKey
	configResponse.Tts.Openai.Model = config.Conf.Tts.Openai.Model
	configResponse.Tts.Openai.Voice = config.Conf.Tts.Openai.Voice
	configResponse.Tts.VClip.ApiKey = config.Conf.Tts.VClip.ApiKey
	configResponse.Tts.VClip.VoiceID = config.Conf.Tts.VClip.VoiceID
	configResponse.Tts.VClip.Speed = config.Conf.Tts.VClip.Speed
	configResponse.Tts.Voices = config.Conf.Tts.Voices

	response.R(c, response.Response{
		Error: 0,
		Msg:   "Get configuration successful",
		Data:  configResponse,
	})
}

// UpdateConfig updates the configuration
func (h Handler) UpdateConfig(c *gin.Context) {
	var req ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.GetLogger().Error("UpdateConfig ShouldBindJSON err", zap.Error(err))
		response.R(c, response.Response{
			Error: -1,
			Msg:   "Parameter error",
			Data:  nil,
		})
		return
	}

	log.GetLogger().Info("Updating configuration info")

	// Update config backup to ensure desktop app detects changes
	config.ConfigBackup = config.Conf

	// Mark config as updated, service needs re-initialization
	configUpdated = true

	// Update app configuration
	config.Conf.App.SegmentDuration = req.App.SegmentDuration
	config.Conf.App.TranscribeParallelNum = req.App.TranscribeParallelNum
	config.Conf.App.TranslateParallelNum = req.App.TranslateParallelNum
	config.Conf.App.TranscribeMaxAttempts = req.App.TranscribeMaxAttempts
	config.Conf.App.TranslateMaxAttempts = req.App.TranslateMaxAttempts
	config.Conf.App.MaxSentenceLength = req.App.MaxSentenceLength
	config.Conf.App.Proxy = req.App.Proxy

	// Update server configuration
	config.Conf.Server.Host = req.Server.Host
	config.Conf.Server.Port = req.Server.Port

	// Update LLM configuration
	config.Conf.Llm.BaseUrl = req.Llm.BaseUrl
	config.Conf.Llm.ApiKey = req.Llm.ApiKey
	config.Conf.Llm.Model = req.Llm.Model

	// Update transcribe configuration
	config.Conf.Transcribe.Provider = req.Transcribe.Provider
	config.Conf.Transcribe.EnableGpuAcceleration = req.Transcribe.EnableGpuAcceleration
	config.Conf.Transcribe.Openai.BaseUrl = req.Transcribe.Openai.BaseUrl
	config.Conf.Transcribe.Openai.ApiKey = req.Transcribe.Openai.ApiKey
	config.Conf.Transcribe.Openai.Model = req.Transcribe.Openai.Model
	config.Conf.Transcribe.Fasterwhisper.Model = req.Transcribe.Fasterwhisper.Model
	config.Conf.Transcribe.Whisperkit.Model = req.Transcribe.Whisperkit.Model
	config.Conf.Transcribe.Whispercpp.Model = req.Transcribe.Whispercpp.Model

	// Update TTS configuration
	config.Conf.Tts.Provider = req.Tts.Provider
	config.Conf.Tts.Openai.BaseUrl = req.Tts.Openai.BaseUrl
	config.Conf.Tts.Openai.ApiKey = req.Tts.Openai.ApiKey
	config.Conf.Tts.Openai.Model = req.Tts.Openai.Model
	config.Conf.Tts.Openai.Voice = req.Tts.Openai.Voice
	config.Conf.Tts.VClip.ApiKey = req.Tts.VClip.ApiKey
	config.Conf.Tts.VClip.VoiceID = req.Tts.VClip.VoiceID
	config.Conf.Tts.VClip.Speed = req.Tts.VClip.Speed
	config.Conf.Tts.Voices = req.Tts.Voices

	// Validate configuration
	if err := config.CheckConfig(); err != nil {
		log.GetLogger().Error("Configuration validation failed", zap.Error(err))
		// Restore original configuration
		config.Conf = config.ConfigBackup
		response.R(c, response.Response{
			Error: -1,
			Msg:   "Configuration validation failed: " + err.Error(),
			Data:  nil,
		})
		return
	}

	// Save configuration to file
	if err := config.SaveConfig(); err != nil {
		log.GetLogger().Error("Save configuration failed", zap.Error(err))
		response.R(c, response.Response{
			Error: -1,
			Msg:   "Save configuration failed: " + err.Error(),
			Data:  nil,
		})
		return
	}

	log.GetLogger().Info("Update configuration successful")
	response.R(c, response.Response{
		Error: 0,
		Msg:   "Update configuration successful",
		Data:  nil,
	})
}
