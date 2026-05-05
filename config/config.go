package config

import (
	"errors"
	"fmt"
	"markflow-studio/log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

var ConfigBackup Config // Used to detect if config is updated before starting a task; restart server if updated

type App struct {
	SegmentDuration       int      `toml:"segment_duration"`
	TranscribeParallelNum int      `toml:"transcribe_parallel_num"`
	TranslateParallelNum  int      `toml:"translate_parallel_num"`
	TranscribeMaxAttempts int      `toml:"transcribe_max_attempts"`
	TranslateMaxAttempts  int      `toml:"translate_max_attempts"`
	MaxSentenceLength     int      `toml:"max_sentence_length"`
	Proxy                 string   `toml:"proxy"`
	ParsedProxy           *url.URL `toml:"-"`
}

type Server struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type OpenaiCompatibleConfig struct {
	BaseUrl string `toml:"base_url"`
	ApiKey  string `toml:"api_key"`
	Model   string `toml:"model"`
	Voice   string `toml:"voice"`
}

type LocalModelConfig struct {
	Model string `toml:"model"`
}

type Transcribe struct {
	Provider              string                 `toml:"provider"`
	EnableGpuAcceleration bool                   `toml:"enable_gpu_acceleration"`
	Openai                OpenaiCompatibleConfig `toml:"openai"`
	Fasterwhisper         LocalModelConfig       `toml:"fasterwhisper"`
	Whisperkit            LocalModelConfig       `toml:"whisperkit"`
	Whispercpp            LocalModelConfig       `toml:"whispercpp"`
}

type VoiceEntry struct {
	Name string `toml:"name" json:"name"`
	ID   string `toml:"id" json:"id"`
}

type VClipConfig struct {
	ApiKey  string  `toml:"api_key"`
	VoiceID string  `toml:"voice_id"`
	Speed   float64 `toml:"speed"`
}

type Tts struct {
	Provider string                 `toml:"provider"`
	Openai   OpenaiCompatibleConfig `toml:"openai"`
	VClip    VClipConfig            `toml:"vclip"`
	Voices   []VoiceEntry           `toml:"voices" json:"voices"`
}

type Config struct {
	App        App                    `toml:"app"`
	Server     Server                 `toml:"server"`
	Llm        OpenaiCompatibleConfig `toml:"llm"`
	Transcribe Transcribe             `toml:"transcribe"`
	Tts        Tts                    `toml:"tts"`
}

var Conf = Config{
	App: App{
		SegmentDuration:       5,
		TranslateParallelNum:  3,
		TranscribeParallelNum: 1,
		TranscribeMaxAttempts: 3,
		TranslateMaxAttempts:  3,
		MaxSentenceLength:     70,
	},
	Server: Server{
		Host: "127.0.0.1",
		Port: 8888,
	},
	Llm: OpenaiCompatibleConfig{
		Model: "gpt-4o-mini",
	},
	Transcribe: Transcribe{
		Provider:              "openai",
		EnableGpuAcceleration: false,
		Openai: OpenaiCompatibleConfig{
			Model: "whisper-1",
		},
		Fasterwhisper: LocalModelConfig{
			Model: "large-v2",
		},
		Whisperkit: LocalModelConfig{
			Model: "large-v2",
		},
		Whispercpp: LocalModelConfig{
			Model: "large-v2",
		},
	},
	Tts: Tts{
		Provider: "vclip",
		Openai: OpenaiCompatibleConfig{
			Model: "tts-1",
			Voice: "alloy",
		},
		VClip: VClipConfig{
			Speed: 1.0,
		},
		Voices: []VoiceEntry{
			{Name: "VClip Nam - Quang Anh", ID: "j8Lc9oB8wuT719K57mQKyG"},
			{Name: "VClip Nữ - Hoài Thu", ID: "tALuifcrB7kpJ74hVXsSWx"},
			{Name: "OpenAI Alloy", ID: "alloy"},
			{Name: "OpenAI Shimmer", ID: "shimmer"},
		},
	},
}

// validateConfig checks if required configurations are complete
func validateConfig() error {
	// Check transcription provider configuration
	switch Conf.Transcribe.Provider {
	case "openai":
		if Conf.Transcribe.Openai.ApiKey == "" {
			return errors.New("Transcription service (OpenAI) requires an API Key")
		}
	case "fasterwhisper":
		if Conf.Transcribe.Fasterwhisper.Model != "tiny" && Conf.Transcribe.Fasterwhisper.Model != "medium" && Conf.Transcribe.Fasterwhisper.Model != "large-v2" && Conf.Transcribe.Fasterwhisper.Model != "large-v3" {
			return errors.New("Fasterwhisper model configuration is incorrect")
		}
	case "whisperkit":
		if runtime.GOOS != "darwin" {
			return fmt.Errorf("WhisperKit only supports macOS")
		}
	case "whispercpp":
		if runtime.GOOS != "windows" {
			return fmt.Errorf("Whisper.cpp only supports Windows")
		}
	}

	// Check TTS provider configuration
	switch Conf.Tts.Provider {
	case "openai":
		if Conf.Tts.Openai.ApiKey == "" {
			return errors.New("Dubbing service (OpenAI TTS) requires an API Key")
		}
	case "vclip":
		if Conf.Tts.VClip.ApiKey == "" {
			return errors.New("Dubbing service (VClip) requires an API Key")
		}
	}

	return nil
}

func LoadConfig() bool {
	var err error
	configPath := "./config/config.toml"
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		log.GetLogger().Info("Config file not found")
		return false
	} else {
		log.GetLogger().Info("Config file found, loading configuration...")
		if _, err = toml.DecodeFile(configPath, &Conf); err != nil {
			log.GetLogger().Error("Failed to load config file", zap.Error(err))
			return false
		}
		return true
	}
}

// CheckConfig validates the configuration
func CheckConfig() error {
	var err error
	// Parse proxy address
	Conf.App.ParsedProxy, err = url.Parse(Conf.App.Proxy)
	if err != nil {
		return err
	}
	return validateConfig()
}

// SaveConfig saves the configuration to file
func SaveConfig() error {
	configPath := filepath.Join("config", "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
		if err != nil {
			return err
		}
	}

	data, err := toml.Marshal(Conf)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
