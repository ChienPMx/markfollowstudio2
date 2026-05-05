package fasterwhisper

import (
	"encoding/json"
	"markflow-studio/config"
	"markflow-studio/internal/storage"
	"markflow-studio/internal/types"
	"markflow-studio/log"
	"markflow-studio/pkg/util"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

func (c *FastwhisperProcessor) Transcription(audioFile, language, workDir string) (*types.TranscriptionData, error) {
	cmdArgs := []string{
		"--model_dir", "./models/",
		"--model", c.Model,
		"--one_word", "2",
		"--output_format", "json",
		"--language", language,
		"--output_dir", workDir,
		audioFile,
	}

	if config.Conf.Transcribe.EnableGpuAcceleration {
		cmdArgs = append(cmdArgs[:len(cmdArgs)-1], "--compute_type", "float16", cmdArgs[len(cmdArgs)-1])
		log.GetLogger().Info("FastwhisperProcessor GPU acceleration enabled", zap.String("model", c.Model))
	}

	cmd := exec.Command(storage.FasterwhisperPath, cmdArgs...)
	log.GetLogger().Info("FastwhisperProcessor transcription started", zap.String("cmd", cmd.String()))
	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "Subtitles are written to") {
		log.GetLogger().Error("FastwhisperProcessor command execution failed", zap.String("output", string(output)), zap.Error(err))
		return nil, err
	}
	log.GetLogger().Info("FastwhisperProcessor transcription JSON generated", zap.String("audio file", audioFile))

	var result types.FasterWhisperOutput
	fileData, err := os.Open(util.ChangeFileExtension(audioFile, ".json"))
	if err != nil {
		log.GetLogger().Error("FastwhisperProcessor failed to open JSON file", zap.Error(err))
		return nil, err
	}
	defer fileData.Close()
	decoder := json.NewDecoder(fileData)
	if err = decoder.Decode(&result); err != nil {
		log.GetLogger().Error("FastwhisperProcessor failed to parse JSON file", zap.Error(err))
		return nil, err
	}

	var (
		transcriptionData types.TranscriptionData
		num               int
	)
	for _, segment := range result.Segments {
		transcriptionData.Text += strings.ReplaceAll(segment.Text, "—", " ") // Hyphen handling, as the model often adds erroneous hyphens
		for _, word := range segment.Words {
			if strings.Contains(word.Word, "—") {
				// Symmetrical splitting
				mid := (word.Start + word.End) / 2
				seperatedWords := strings.Split(word.Word, "—")
				transcriptionData.Words = append(transcriptionData.Words, []types.Word{
					{
						Num:   num,
						Text:  util.CleanPunction(strings.TrimSpace(seperatedWords[0])),
						Start: word.Start,
						End:   mid,
					},
					{
						Num:   num + 1,
						Text:  util.CleanPunction(strings.TrimSpace(seperatedWords[1])),
						Start: mid,
						End:   word.End,
					},
				}...)
				num += 2
			} else {
				transcriptionData.Words = append(transcriptionData.Words, types.Word{
					Num:   num,
					Text:  util.CleanPunction(strings.TrimSpace(word.Word)),
					Start: word.Start,
					End:   word.End,
				})
				num++
			}
		}
	}
	log.GetLogger().Info("FastwhisperProcessor transcription successful")
	return &transcriptionData, nil
}
