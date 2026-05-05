package whisperx

import (
	"encoding/json"
	"markflow-studio/internal/storage"
	"markflow-studio/internal/types"
	"markflow-studio/log"
	"markflow-studio/pkg/util"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

func (c *WhisperXProcessor) Transcription(audioFile, language, workDir string) (*types.TranscriptionData, error) {
	var (
		cmdArgs []string
		envPath string
		cmd     *exec.Cmd
	)
	if runtime.GOOS == "windows" {
		envPath = ".\\bin\\whisperx\\.venv\\Scripts\\activate"
	} else {
		envPath = storage.WhisperXPath
	}
	if runtime.GOOS == "windows" {
		cmdArgs = []string{
			"&&",
			storage.WhisperXPath,
			audioFile,
			"--model_dir", "./models/whisperx",
			"--model", c.Model,
			"--language", language,
			"--output_dir", workDir,
			"--compute_type", "float16",
			"--batch_size", "8",
			"--model_cache_only", "True",
		}
		cmd = exec.Command(envPath, cmdArgs...)
	} else {
		cmdArgs = []string{
			audioFile,
			"--model_dir", "./models/whisperx",
			"--model", c.Model,
			"--language", language,
			"--output_dir", workDir,
			"--compute_type", "float16",
			"--batch_size", "16",
			"--model_cache_only", "True",
		}
		cmd = exec.Command(envPath, cmdArgs...)
		cudaLibPath := "LD_LIBRARY_PATH=./bin/whisperx/.venv/lib/python3.12/site-packages/nvidia/cudnn/lib"
		currentEnv := os.Environ()
		newEnv := append(currentEnv, cudaLibPath)
		cmd.Env = newEnv
	}
	log.GetLogger().Info("WhisperXProcessor transcription started", zap.String("cmd", cmd.String()))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.GetLogger().Error("WhisperXProcessor cmd execution failed", zap.String("output", string(output)), zap.Error(err))
		return nil, err
	}
	log.GetLogger().Info("WhisperXProcessor transcription JSON generation complete", zap.String("audio file", audioFile))

	var result types.WhisperXOutput
	fileData, err := os.Open(util.ChangeFileExtension(audioFile, ".json"))
	if err != nil {
		log.GetLogger().Error("WhisperXProcessor failed to open JSON file", zap.Error(err))
		return nil, err
	}
	defer fileData.Close()
	decoder := json.NewDecoder(fileData)
	if err = decoder.Decode(&result); err != nil {
		log.GetLogger().Error("WhisperXProcessor failed to parse JSON file", zap.Error(err))
		return nil, err
	}

	var (
		transcriptionData types.TranscriptionData
		num               int
	)
	for _, segment := range result.Segments {
		transcriptionData.Text += strings.ReplaceAll(segment.Text, "—", " ") // Hyphen handling, as the model often incorrectly adds hyphens
		for _, word := range segment.Words {
			if strings.Contains(word.Word, "—") {
				// Symmetric splitting
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
	log.GetLogger().Info("WhisperXProcessor transcription success")
	return &transcriptionData, nil
}
