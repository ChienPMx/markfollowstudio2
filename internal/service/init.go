package service

import (
	"krillin-ai/config"
	"krillin-ai/internal/types"
	"krillin-ai/log"
	"krillin-ai/pkg/fasterwhisper"
	"krillin-ai/pkg/localtts"
	"krillin-ai/pkg/openai"
	"krillin-ai/pkg/vclip"
	"krillin-ai/pkg/whisper"
	"krillin-ai/pkg/whispercpp"
	"krillin-ai/pkg/whisperkit"

	"go.uber.org/zap"
)

type Service struct {
	Transcriber   types.Transcriber
	ChatCompleter types.ChatCompleter
	TtsClient     types.Ttser
}

func NewService() *Service {
	var transcriber types.Transcriber
	var chatCompleter types.ChatCompleter
	var ttsClient types.Ttser

	switch config.Conf.Transcribe.Provider {
	case "openai":
		transcriber = whisper.NewClient(config.Conf.Transcribe.Openai.BaseUrl, config.Conf.Transcribe.Openai.ApiKey, config.Conf.App.Proxy)
	case "fasterwhisper":
		transcriber = fasterwhisper.NewFastwhisperProcessor(config.Conf.Transcribe.Fasterwhisper.Model)
	case "whispercpp":
		transcriber = whispercpp.NewWhispercppProcessor(config.Conf.Transcribe.Whispercpp.Model)
	case "whisperkit":
		transcriber = whisperkit.NewWhisperKitProcessor(config.Conf.Transcribe.Whisperkit.Model)
	default:
		log.GetLogger().Error("Unsupported transcription provider", zap.String("provider", config.Conf.Transcribe.Provider))
	}
	log.GetLogger().Info("Currently selected transcription source: ", zap.String("transcriber", config.Conf.Transcribe.Provider))

	chatCompleter = openai.NewClient(config.Conf.Llm.BaseUrl, config.Conf.Llm.ApiKey, config.Conf.App.Proxy)

	switch config.Conf.Tts.Provider {
	case "openai":
		ttsClient = openai.NewClient(config.Conf.Tts.Openai.BaseUrl, config.Conf.Tts.Openai.ApiKey, config.Conf.App.Proxy)
	case "edge-tts":
		ttsClient = localtts.NewEdgeTtsClient()
	case "vclip":
		ttsClient = vclip.NewClient(config.Conf.Tts.VClip.ApiKey, config.Conf.Tts.VClip.VoiceID, config.Conf.Tts.VClip.Speed)
	}

	return &Service{
		Transcriber:   transcriber,
		ChatCompleter: chatCompleter,
		TtsClient:     ttsClient,
	}
}
