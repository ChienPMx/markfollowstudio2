package service

import (
	"context"
	"fmt"
	"markflow-studio/config"
	"markflow-studio/internal/storage"
	"markflow-studio/internal/types"
	"markflow-studio/log"
	"markflow-studio/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Input subtitles, generate voiceover
func (s Service) srtFileToSpeech(ctx context.Context, stepParam *types.SubtitleTaskStepParam) error {
	if !stepParam.EnableTts {
		return nil
	}
	// Step 1: Parse subtitle file
	subtitles, err := parseSRT(stepParam.TtsSourceFilePath, stepParam.SubtitleResultType)
	if err != nil {
		log.GetLogger().Error("srtFileToSpeech parseSRT error", zap.Any("stepParam", stepParam), zap.Error(err))
		return fmt.Errorf("srtFileToSpeech parseSRT error: %w", err)
	}

	var audioFiles []string
	var currentTime time.Time

	// Create file to record audio start and end times
	durationDetailFile, err := os.Create(filepath.Join(stepParam.TaskBasePath, types.TtsAudioDurationDetailsFileName))
	if err != nil {
		log.GetLogger().Error("srtFileToSpeech create durationDetailFile error", zap.Any("stepParam", stepParam), zap.Error(err))
		return fmt.Errorf("srtFileToSpeech create durationDetailFile error: %w", err)
	}
	defer durationDetailFile.Close()

	// Step 2: TTS conversion
	voiceCode := stepParam.TtsVoiceCode

	// Handle TTS conversion concurrently
	err = s.processSubtitlesConcurrently(subtitles, voiceCode, stepParam)
	if err != nil {
		log.GetLogger().Error("srtFileToSpeech processSubtitlesConcurrently error", zap.Any("stepParam", stepParam), zap.Error(err))
		return fmt.Errorf("srtFileToSpeech processSubtitlesConcurrently error: %w", err)
	}

	for i, sub := range subtitles {
		outputFile := filepath.Join(stepParam.TaskBasePath, fmt.Sprintf("subtitle_%d.wav", i+1))

		// Step 3: Adjust audio duration
		startTime, err := time.Parse("15:04:05,000", sub.Start)
		if err != nil {
			log.GetLogger().Error("srtFileToSpeech parse time error", zap.Any("stepParam", stepParam), zap.Any("num", i+1), zap.String("time str", sub.Start), zap.Error(err))
			return fmt.Errorf("srtFileToSpeech parse time error: %w", err)
		}
		endTime, err := time.Parse("15:04:05,000", sub.End)
		if err != nil {
			log.GetLogger().Error("audioToSubtitle.time.Parse error", zap.Any("stepParam", stepParam), zap.Any("num", i+1), zap.String("time str", sub.Start), zap.Error(err))
			return fmt.Errorf("srtFileToSpeech audioToSubtitle.time.Parse error: %w", err)
		}
		if i == 0 {
			// If the first subtitle doesn't start at 00:00, add silence frames
			if startTime.Second() > 0 {
				silenceDurationMs := startTime.Sub(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)).Milliseconds()
				silenceFilePath := filepath.Join(stepParam.TaskBasePath, "silence_0.wav")
				err := newGenerateSilence(silenceFilePath, float64(silenceDurationMs)/1000)
				if err != nil {
					log.GetLogger().Error("srtFileToSpeech newGenerateSilence error", zap.Any("stepParam", stepParam), zap.Error(err))
					return fmt.Errorf("srtFileToSpeech newGenerateSilence error: %w", err)
				}
				audioFiles = append(audioFiles, silenceFilePath)

				// Calculate the end time of silence frames
				silenceEndTime := currentTime.Add(time.Duration(silenceDurationMs) * time.Millisecond)
				durationDetailFile.WriteString(fmt.Sprintf("Silence: start=%s, end=%s\n", currentTime.Format("15:04:05,000"), silenceEndTime.Format("15:04:05,000")))
				currentTime = silenceEndTime
			}
		}

		duration := endTime.Sub(startTime).Seconds()
		if i < len(subtitles)-1 {
			// If it's not the last subtitle, increase silence frame duration
			nextStartTime, err := time.Parse("15:04:05,000", subtitles[i+1].Start)
			if err != nil {
				log.GetLogger().Error("srtFileToSpeech parse time error", zap.Any("stepParam", stepParam), zap.Any("num", i+2), zap.String("time str", subtitles[i+1].Start), zap.Error(err))
				return fmt.Errorf("srtFileToSpeech parse time error: %w", err)
			}
			if endTime.Before(nextStartTime) {
				duration = nextStartTime.Sub(startTime).Seconds()
			}
		}

		adjustedFile := filepath.Join(stepParam.TaskBasePath, fmt.Sprintf("adjusted_%d.wav", i+1))
		err = adjustAudioDuration(outputFile, adjustedFile, stepParam.TaskBasePath, duration)
		if err != nil {
			log.GetLogger().Error("srtFileToSpeech adjustAudioDuration error", zap.Any("stepParam", stepParam), zap.Any("num", i+1), zap.Error(err))
			return fmt.Errorf("srtFileToSpeech adjustAudioDuration error: %w", err)
		}

		audioFiles = append(audioFiles, adjustedFile)

		// Calculate actual audio duration
		audioDuration, err := util.GetAudioDuration(adjustedFile)
		if err != nil {
			log.GetLogger().Error("srtFileToSpeech GetAudioDuration error", zap.Any("stepParam", stepParam), zap.Any("num", i+1), zap.Error(err))
			return fmt.Errorf("srtFileToSpeech GetAudioDuration error: %w", err)
		}

		// Calculate audio end time
		audioEndTime := currentTime.Add(time.Duration(audioDuration*1000) * time.Millisecond)
		// Write to file
		durationDetailFile.WriteString(fmt.Sprintf("Audio %d: start=%s, end=%s\n", i+1, currentTime.Format("15:04:05,000"), audioEndTime.Format("15:04:05,000")))
		currentTime = audioEndTime
	}

	// Step 6: Concatenate all audio files
	finalOutput := filepath.Join(stepParam.TaskBasePath, types.TtsResultAudioFileName)
	err = concatenateAudioFiles(audioFiles, finalOutput, stepParam.TaskBasePath)
	if err != nil {
		log.GetLogger().Error("srtFileToSpeech concatenateAudioFiles error", zap.Any("stepParam", stepParam), zap.Error(err))
		return fmt.Errorf("srtFileToSpeech concatenateAudioFiles error: %w", err)
	}
	stepParam.TtsResultFilePath = finalOutput

	videoWithTtsPath := filepath.Join(stepParam.TaskBasePath, types.SubtitleTaskVideoWithTtsFileName)
	// Compose new video with replaced audio
	err = util.ReplaceAudioInVideo(stepParam.InputVideoPath, finalOutput, videoWithTtsPath)
	if err != nil {
		log.GetLogger().Error("srtFileToSpeech ReplaceAudioInVideo error", zap.Any("stepParam", stepParam), zap.Error(err))
	}
	stepParam.VideoWithTtsFilePath = videoWithTtsPath
	// Update subtitle task info
	stepParam.TaskPtr.ProcessPct = 98
	log.GetLogger().Info("srtFileToSpeech success", zap.String("task id", stepParam.TaskId))
	return nil
}

func (s Service) processSubtitlesConcurrently(subtitles []types.SrtSentenceWithStrTime, voiceCode string, stepParam *types.SubtitleTaskStepParam) error {
	// Create a results array to store processing result for each subtitle
	type processingResult struct {
		index int
		err   error
	}

	maxConcurrency := 3 // Reduce concurrency to lower network load
	if config.Conf.Tts.Provider == "vclip" {
		maxConcurrency = 1 // VClip only allows 1 concurrent export
	}
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	resultCh := make(chan processingResult, len(subtitles))

	// Generate all audio files concurrently
	for i, sub := range subtitles {
		wg.Add(1)
		go func(index int, subtitle types.SrtSentenceWithStrTime) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			outputFile := filepath.Join(stepParam.TaskBasePath, fmt.Sprintf("subtitle_%d.wav", index+1))
			err := s.TtsClient.Text2Speech(subtitle.Text, voiceCode, outputFile)
			if err != nil {
				log.GetLogger().Error("processSubtitlesConcurrently Text2Speech error",
					zap.Any("index", index+1),
					zap.String("text", subtitle.Text),
					zap.Error(err))
				resultCh <- processingResult{index: index, err: fmt.Errorf("subtitle %d TTS error: %w", index+1, err)}
				return
			}

			// Successfully processed
			resultCh <- processingResult{index: index, err: nil}
		}(i, sub)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultCh)

	// Collect results and count errors
	results := make([]processingResult, len(subtitles))
	errorCount := 0
	var firstError error

	for result := range resultCh {
		results[result.index] = result
		if result.err != nil {
			errorCount++
			if firstError == nil {
				firstError = result.err
			}
		}
	}

	// If more than half the subtitles fail, return an error
	failureThreshold := len(subtitles) / 2
	if errorCount > failureThreshold {
		log.GetLogger().Error("processSubtitlesConcurrently: too many failures",
			zap.Int("total", len(subtitles)),
			zap.Int("errors", errorCount),
			zap.Int("threshold", failureThreshold))
		return fmt.Errorf("too many TTS failures: %d/%d failed, first error: %w", errorCount, len(subtitles), firstError)
	}

	// Verify existence of successful files; generate silence for failed ones
	for i, result := range results {
		outputFile := filepath.Join(stepParam.TaskBasePath, fmt.Sprintf("subtitle_%d.wav", i+1))

		if result.err != nil {
			// Generate silence for failed subtitles
			log.GetLogger().Warn("Creating silence file as replacement for failed TTS",
				zap.Int("index", i+1),
				zap.String("file", outputFile))

			// Generate 0.5s silence as a replacement
			err := newGenerateSilence(outputFile, 0.5)
			if err != nil {
				log.GetLogger().Error("Failed to create replacement silence file",
					zap.Int("index", i+1),
					zap.Error(err))
				return fmt.Errorf("failed to generate silence for subtitle %d: %w", i+1, err)
			}
		} else {
			// Verify if the successfully generated file exists
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				log.GetLogger().Error("processSubtitlesConcurrently output file not exist",
					zap.Any("index", i+1),
					zap.String("file", outputFile))
				return fmt.Errorf("subtitle %d output file not exist: %s", i+1, outputFile)
			}
		}
	}

	if errorCount > 0 {
		log.GetLogger().Warn("processSubtitlesConcurrently completed with some failures",
			zap.Int("total", len(subtitles)),
			zap.Int("errors", errorCount),
			zap.Int("success", len(subtitles)-errorCount))
	} else {
		log.GetLogger().Info("processSubtitlesConcurrently completed successfully", zap.Int("total", len(subtitles)))
	}

	return nil
}

func parseSRT(filePath string, resultType types.SubtitleResultType) ([]types.SrtSentenceWithStrTime, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parseSRT read file error: %w", err)
	}

	var subtitles []types.SrtSentenceWithStrTime
	
	// Normalize newlines
	dataStr := strings.ReplaceAll(string(data), "\r\n", "\n")
	blocks := strings.Split(dataStr, "\n\n")

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		lines := strings.Split(block, "\n")
		if len(lines) < 3 {
			continue
		}

		// Line 1 is Time
		timeLine := lines[1]
		timeParts := strings.Split(timeLine, " --> ")
		if len(timeParts) != 2 {
			continue
		}

		// Text lines
		textLines := lines[2:]
		var targetText string

		if len(textLines) == 1 {
			targetText = textLines[0]
		} else {
			// Bilingual block
			if resultType == types.SubtitleResultTypeBilingualTranslationOnTop {
				targetText = textLines[0]
			} else if resultType == types.SubtitleResultTypeBilingualTranslationOnBottom {
				targetText = textLines[1]
			} else {
				targetText = textLines[0] // fallback
			}
		}

		subtitles = append(subtitles, types.SrtSentenceWithStrTime{
			Start: strings.TrimSpace(timeParts[0]),
			End:   strings.TrimSpace(timeParts[1]),
			Text:  targetText,
		})
	}

	return subtitles, nil
}

func newGenerateSilence(outputAudio string, duration float64) error {
	// Generate silence in PCM format
	cmd := exec.Command(storage.FfmpegPath, "-y", "-f", "lavfi", "-i", "anullsrc=channel_layout=mono:sample_rate=44100", "-t",
		fmt.Sprintf("%.3f", duration), "-ar", "44100", "-ac", "1", "-c:a", "pcm_s16le", outputAudio)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("newGenerateSilence failed to generate PCM silence: %w", err)
	}

	return nil
}

// Adjust audio duration to match subtitle duration
func adjustAudioDuration(inputFile, outputFile, taskBasePath string, subtitleDuration float64) error {
	// Get audio duration
	audioDuration, err := util.GetAudioDuration(inputFile)
	if err != nil {
		return err
	}

	// If audio is shorter than subtitle, insert silence to extend it
	if audioDuration < subtitleDuration {
		// Calculate duration of silence to insert
		silenceDuration := subtitleDuration - audioDuration

		// Generate silence audio
		silenceFile := filepath.Join(taskBasePath, "silence.wav")
		err := newGenerateSilence(silenceFile, silenceDuration)
		if err != nil {
			return fmt.Errorf("error generating silence: %v", err)
		}

		silenceAudioDuration, _ := util.GetAudioDuration(silenceFile)
		log.GetLogger().Info("adjustAudioDuration", zap.Any("silenceDuration", silenceAudioDuration))

		// Concatenate audio and silence
		concatFile := filepath.Join(taskBasePath, "concat.txt")
		f, err := os.Create(concatFile)
		if err != nil {
			return fmt.Errorf("adjustAudioDuration create concat file error: %w", err)
		}
		defer os.Remove(concatFile)

		_, err = f.WriteString(fmt.Sprintf("file '%s'\nfile '%s'\n", filepath.Base(inputFile), filepath.Base(silenceFile)))
		if err != nil {
			return fmt.Errorf("adjustAudioDuration write to concat file error: %v", err)
		}
		f.Close()

		cmd := exec.Command(storage.FfmpegPath, "-y", "-f", "concat", "-safe", "0", "-i", concatFile, "-c", "copy", outputFile)
		log.GetLogger().Info("adjustAudioDuration", zap.Any("inputFile", inputFile), zap.Any("outputFile", outputFile), zap.String("run command", cmd.String()))
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("adjustAudioDuration concat audio and silence  error: %v", err)
		}

		concatFileDuration, _ := util.GetAudioDuration(outputFile)
		log.GetLogger().Info("adjustAudioDuration", zap.Any("concatFileDuration", concatFileDuration))
		return nil
	}

	// If audio is longer than subtitle, scale the audio playback rate
	if audioDuration > subtitleDuration {
		// Calculate playback rate
		speed := audioDuration / subtitleDuration
		//if speed < 0.5 || speed > 2.0 {
		//	// The rate is generally within the range supported by FFmpeg [0.5, 2.0]
		//	return fmt.Errorf("speed factor %.2f is out of range (0.5 to 2.0)", speed)
		//}

		// Adjust playback rate using atempo filter
		cmd := exec.Command(storage.FfmpegPath, "-y", "-i", inputFile, "-filter:a", fmt.Sprintf("atempo=%.2f", speed), outputFile)
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// If durations match, copy the file directly
	return util.CopyFile(inputFile, outputFile)
}

// Concatenate audio files
func concatenateAudioFiles(audioFiles []string, outputFile, taskBasePath string) error {
	// Create a temporary file to save the audio file list
	listFile := filepath.Join(taskBasePath, "audio_list.txt")
	f, err := os.Create(listFile)
	if err != nil {
		return err
	}
	defer os.Remove(listFile)

	for _, file := range audioFiles {
		_, err := f.WriteString(fmt.Sprintf("file '%s'\n", filepath.Base(file)))
		if err != nil {
			return err
		}
	}
	f.Close()

	cmd := exec.Command(storage.FfmpegPath, "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", outputFile)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}