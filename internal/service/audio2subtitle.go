package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"krillin-ai/config"
	"krillin-ai/internal/types"
	"krillin-ai/log"
	"krillin-ai/pkg/util"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Translation result data structure
type TranslatedItem struct {
	OriginText     string
	TranslatedText string
}

func (s Service) audioToSubtitle(ctx context.Context, stepParam *types.SubtitleTaskStepParam) error {
	var err error
	err = s.audioToSrt(ctx, stepParam) // Progress updated to 90% here
	if err != nil {
		return fmt.Errorf("audioToSubtitle audioToSrt error: %w", err)
	}
	err = splitSrt(stepParam)
	if err != nil {
		return fmt.Errorf("audioToSubtitle splitSrt error: %w", err)
	}
		// Update subtitle task info
	stepParam.TaskPtr.ProcessPct = 95
	return nil
}

//func splitAudio(stepParam *types.SubtitleTaskStepParam) error {
//	log.GetLogger().Info("audioToSubtitle.splitAudio start", zap.String("task id", stepParam.TaskId))
//	var err error
//	// Split audio using ffmpeg
//	outputPattern := filepath.Join(stepParam.TaskBasePath, types.SubtitleTaskSplitAudioFileNamePattern) // Output file pattern
//	segmentDuration := config.Conf.App.SegmentDuration * 60
//
//	cmd := exec.Command(
//		storage.FfmpegPath,
//		"-i", stepParam.AudioFilePath, // Input
//		"-f", "segment", // Output file format is segments
//		"-segment_time", fmt.Sprintf("%d", segmentDuration), // Duration of each segment (in seconds)
//		"-reset_timestamps", "1", // Reset timestamps for each segment
//		"-y", // Overwrite output file
//		outputPattern,
//	)
//	err = cmd.Run()
//	if err != nil {
//		log.GetLogger().Error("audioToSubtitle splitAudio ffmpeg err", zap.Any("stepParam", stepParam), zap.Error(err))
//		return fmt.Errorf("audioToSubtitle splitAudio ffmpeg err: %w", err)
//	}
//
//	// Get the list of split files
//	audioFiles, err := filepath.Glob(filepath.Join(stepParam.TaskBasePath, fmt.Sprintf("%s_*.mp3", types.SubtitleTaskSplitAudioFileNamePrefix)))
//	if err != nil {
//		log.GetLogger().Error("audioToSubtitle splitAudio filepath.Glob err", zap.Any("stepParam", stepParam), zap.Error(err))
//		return fmt.Errorf("audioToSubtitle splitAudio filepath.Glob err: %w", err)
//	}
//	if len(audioFiles) == 0 {
//		log.GetLogger().Error("audioToSubtitle splitAudio no audio files found", zap.Any("stepParam", stepParam))
//		return errors.New("audioToSubtitle splitAudio no audio files found")
//	}
//
//	for _, audioFile := range audioFiles {
//		stepParam.SmallAudios = append(stepParam.SmallAudios, &types.SmallAudio{
//			AudioFile: audioFile,
//		})
//	}
//
//	// Update subtitle task info
//	stepParam.TaskPtr.ProcessPct = 20
//
//	log.GetLogger().Info("audioToSubtitle.splitAudio end", zap.String("task id", stepParam.TaskId))
//	return nil
//}

func (s Service) transcribeAudio(id int, audioFilePath string, language string, taskBasePath string) (transcriptionData *types.TranscriptionData, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("audioToSubtitle transcribeAudio panic recovered: %v", r)
		}
	}()

	if language == "zh_cn" {
		language = "zh" // Switch language code
	}
	transcriptionData, err = s.Transcriber.Transcription(audioFilePath, language, taskBasePath)

	if err != nil {
		return nil, fmt.Errorf("audioToSubtitle transcribeAudio Transcription err: %w", err)
	}

	_ = util.SaveToDisk(transcriptionData, filepath.Join(taskBasePath, fmt.Sprintf(types.SubtitleTaskAudioTranscriptionDataPersistenceFileNamePattern, id)))

	if transcriptionData.Text == "" {
		log.GetLogger().Info("audioToSubtitle transcribeAudio TranscriptionData.Text is empty", zap.Any("audioFilePath", audioFilePath), zap.Any("taskBasePath", taskBasePath))
	}
	return transcriptionData, nil
}

func (s Service) IsSplitUseSpace(language types.StandardLanguageCode) bool {
	if language == types.LanguageNameSimplifiedChinese || language == types.LanguageNameTraditionalChinese ||
		language == types.LanguageNameJapanese || language == types.LanguageNameKorean || language == types.LanguageNameThai {
		return true
	}

	return false
}

func (s Service) splitTextAndTranslateV2(basePath, inputText string, originLang, targetLang types.StandardLanguageCode, enableModalFilter bool, id int) ([]*TranslatedItem, error) {
	sentences := util.SplitTextSentences(inputText, config.Conf.App.MaxSentenceLength)
	if len(sentences) == 0 {
		return []*TranslatedItem{}, nil
	}
	// Patch: whisper often skips punctuation at the end of Chinese sentences, which breaks the split logic above.
	if s.IsSplitUseSpace(originLang) {
		newSentences := make([]string, 0)
		for _, sentence := range sentences {
			newSentences = append(newSentences, strings.Split(sentence, " ")...)
		}
		sentences = newSentences
	}

	shortSentences := make([]string, 0)
		// If sentence is still too long, use LLM to split further
	for _, sentence := range sentences {
		if sentence == "" {
			continue
		}
		if util.CountEffectiveChars(sentence) <= config.Conf.App.MaxSentenceLength {
			shortSentences = append(shortSentences, sentence)
			continue
		}

				// Recursively split long sentences until they meet length requirements while maintaining order
		splitSentences, err := s.splitSentenceRecursively(sentence, 0, 5) // Max 5 levels of recursion
		if err != nil {
			log.GetLogger().Error("splitSentenceRecursively error", zap.Error(err), zap.Any("sentence", sentence))
			// If split fails, add original sentence directly
			shortSentences = append(shortSentences, sentence)
		} else {
			shortSentences = append(shortSentences, splitSentences...)
		}
	}

	sentences = shortSentences

	var (
		signal  = make(chan struct{}, config.Conf.App.TranslateParallelNum) // Control max concurrency
		wg      sync.WaitGroup
		results = make([]*TranslatedItem, len(sentences))
		// errChan = make(chan error, 1)
		// mutex   sync.Mutex
	)

	for i, sentence := range sentences {
		wg.Add(1)
		signal <- struct{}{}

		go func(index int, originText string) {
			defer wg.Done()
			defer func() { <-signal }()

			contextSentenceNum := 3

			// Generate string for previous 3 sentences
			var previousSentences string
			if index > 0 {
				start := 0
				if index-contextSentenceNum > 0 {
					start = index - contextSentenceNum
				}
				for i := start; i < index; i++ {
					previousSentences += sentences[i] + "\n"
				}
			}

			// Generate string for next 3 sentences
			var nextSentences string
			if index < len(sentences)-1 {
				end := len(sentences) - 1
				if index+contextSentenceNum < end {
					end = index + contextSentenceNum
				}
				for i := index + 1; i <= end; i++ {
					if i > index+1 {
						nextSentences += "\n"
					}
					nextSentences += sentences[i]
				}
			}

			prompt := fmt.Sprintf(types.SplitTextWithContextPrompt, types.GetStandardLanguageName(targetLang), previousSentences, originText, nextSentences)

			translatedText, err := s.ChatCompleter.ChatCompletion(prompt)
			if err != nil {
				log.GetLogger().Error("splitTextAndTranslateV2 llm translate error", zap.Error(err), zap.Any("original text", originText))
				results[index] = &TranslatedItem{
					OriginText:     originText,
					TranslatedText: originText,
				}
			} else {
				translatedText = strings.TrimSpace(translatedText)
				results[index] = &TranslatedItem{
					OriginText:     originText,
					TranslatedText: translatedText,
				}
			}
		}(i, sentence)
	}

	wg.Wait()
	// close(errChan)

	return results, nil
}

func (s Service) audioToSrt(ctx context.Context, stepParam *types.SubtitleTaskStepParam) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.GetLogger().Error("audioToSubtitle audioToSrt panic recovered", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))
			err = fmt.Errorf("audioToSubtitle audioToSrt panic recovered: %v", r)
		}
	}()

	log.GetLogger().Info("audioToSubtitle.audioToSrt start", zap.Any("taskId", stepParam.TaskId))
	timePoints, err := GetSplitPoints(stepParam.AudioFilePath, float64(config.Conf.App.SegmentDuration)*60)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle audioToSrt GetSplitPoints err", zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle audioToSrt GetSplitPoints err: %w", err)
	}
	log.GetLogger().Info("audioToSubtitle audioToSrt GetSplitPoints completed", zap.Any("taskId", stepParam.TaskId), zap.Any("timePoints", timePoints))

	// Update subtitle task info
	stepParam.TaskPtr.ProcessPct = 15
	segmentNum := len(timePoints) - 1

	type DataWithId[T any] struct {
		Data T
		Id   int
	}

	var (
		// Queue for audio segments to be clipped
		pendingSplitQueue = make(chan DataWithId[[2]float64], segmentNum)
		// Queue for clipping results
		splitResultQueue = make(chan DataWithId[string], segmentNum)
		// Queue for audio files to be transcribed
		pendingTranscriptionQueue = make(chan DataWithId[string], segmentNum)
		// Queue for transcription results
		transcribedQueue = make(chan DataWithId[*types.TranscriptionData], segmentNum)
		// Queue for text to be translated
		pendingTranslationQueue = make(chan DataWithId[string], segmentNum)
		// Queue for translation results
		translatedQueue = make(chan DataWithId[[]*TranslatedItem], segmentNum)
	)
	eg, ctx := errgroup.WithContext(ctx)

	log.GetLogger().Info("audioToSubtitle.audioToSrt start", zap.Any("taskId", stepParam.TaskId))

	// Construct audio segments slice of length segmentNum
	type AudioSegment struct {
		AudioFile         string
		TranscriptionData *types.TranscriptionData
		SrtNoTsFile       string
	}
	audioSegments := make([]AudioSegment, segmentNum)

	// Input audio files to split queue
	for i := range segmentNum {
		pendingSplitQueue <- DataWithId[[2]float64]{
			Data: [2]float64{timePoints[i], timePoints[i+1]},
			Id:   i,
		}
	}

	// Split audio
	for range runtime.NumCPU() {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case splitItem, ok := <-pendingSplitQueue:
					if !ok {
						return nil
					}
					log.GetLogger().Info("Begin split audio", zap.Any("taskId", stepParam.TaskId), zap.Any("splitId", splitItem.Id))
					// Split audio
					outputFileName := filepath.Join(stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitAudioFileNamePattern, splitItem.Id))
					err := ClipAudio(stepParam.AudioFilePath, outputFileName, splitItem.Data[0], splitItem.Data[1])
					if err != nil {
						return fmt.Errorf("audioToSubtitle audioToSrt ClipAudio err: %w", err)
					}
					log.GetLogger().Info("Split audio completed", zap.Any("taskId", stepParam.TaskId), zap.Any("splitId", splitItem.Id))

					// Send split results
					splitResultQueue <- DataWithId[string]{
						Data: outputFileName,
						Id:   splitItem.Id,
					}
				}
			}
		})
	}

	// Audio transcription
	for range config.Conf.App.TranscribeParallelNum {
		eg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case audioFileItem, ok := <-pendingTranscriptionQueue:
					if !ok {
						return nil
					}
					var (
						err               error
						transcriptionData *types.TranscriptionData
					)
					log.GetLogger().Info("Begin transcribe", zap.Any("taskId", stepParam.TaskId), zap.Any("splitId", audioFileItem.Id))
					// Speech to text
					for range config.Conf.App.TranscribeMaxAttempts {
						transcriptionData, err = s.transcribeAudio(audioFileItem.Id, audioFileItem.Data, string(stepParam.OriginLanguage), stepParam.TaskBasePath)
						if err == nil {
							break
						}
					}
					if err != nil {
						return fmt.Errorf("audioToSubtitle audioToSrt Transcription err: %w", err)
					}
					log.GetLogger().Info("Transcribe completed", zap.Any("taskId", stepParam.TaskId), zap.Any("splitId", audioFileItem.Id))

					// Send transcription results
					transcribedQueue <- DataWithId[*types.TranscriptionData]{
						Data: transcriptionData,
						Id:   audioFileItem.Id,
					}
				}
			}
		})
	}

	// Sentence split + translation
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case translateItem, ok := <-pendingTranslationQueue:
				if !ok {
					return nil
				}
				var translatedResults []*TranslatedItem
				var err error
				// Translate text
				log.GetLogger().Info("Begin to translate", zap.Any("taskId", stepParam.TaskId), zap.Any("splitId", translateItem.Id))
				for range config.Conf.App.TranslateMaxAttempts {
					translatedResults, err = s.splitTextAndTranslateV2(stepParam.TaskBasePath, translateItem.Data, stepParam.OriginLanguage, stepParam.TargetLanguage, stepParam.EnableModalFilter, translateItem.Id)
					if err == nil {
						break
					}
				}
				if err != nil {
					return fmt.Errorf("audioToSubtitle audioToSrt splitTextAndTranslate err: %w", err)
				}
				_ = util.SaveToDisk(translatedResults, filepath.Join(stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskTranslationDataPersistenceFileNamePattern, translateItem.Id)))
				log.GetLogger().Info("Translate completed", zap.Any("taskId", stepParam.TaskId), zap.Any("splitId", translateItem.Id))
				// Secondary split for long sentences
				splitResults, err := s.splitTranslateItem(translatedResults)
				if err != nil {
					// Do not interrupt
					log.GetLogger().Error("audioToSubtitle audioToSrt splitTranslateItem err", zap.Any("taskId", stepParam.TaskId), zap.Any("splitId", translateItem.Id), zap.Error(err))
					translatedQueue <- DataWithId[[]*TranslatedItem]{
						Data: translatedResults,
						Id:   translateItem.Id,
					}
				} else {
					translatedQueue <- DataWithId[[]*TranslatedItem]{
						Data: splitResults,
						Id:   translateItem.Id,
					}
				}
			}
		}
	})

	// Process results, update subtitle task info
	eg.Go(func() error {
		// SPLIT_WEIGHT + TRANSCRIBE_WEIGHT + TRANSLATE_WEIGHT == 1
		const (
			SPLIT_WEIGHT      = 0.1
			TRANSCRIBE_WEIGHT = 0.4
			TRANSLATE_WEIGHT  = 0.5
		)
		// Weight of overall task in progress bar
		taskWeight := (90 - 15) / float64(segmentNum)
		processPct := 15.0
		// Number of completed tasks
		completedTasks := 0
		for {
			select {
			case <-ctx.Done():
				return nil
			case splitResultItem := <-splitResultQueue:
				// Update subtitle task info
				processPct += taskWeight * SPLIT_WEIGHT
				stepParam.TaskPtr.ProcessPct = uint8(processPct)
				// Process split results
				audioSegments[splitResultItem.Id].AudioFile = splitResultItem.Data
				// Send transcription task
				pendingTranscriptionQueue <- DataWithId[string]{
					Data: splitResultItem.Data,
					Id:   splitResultItem.Id,
				}
			case transcribedItem := <-transcribedQueue:
				// Update subtitle task info
				processPct += taskWeight * TRANSCRIBE_WEIGHT
				stepParam.TaskPtr.ProcessPct = uint8(processPct)
				// Process transcription results
				audioSegments[transcribedItem.Id].TranscriptionData = transcribedItem.Data
				// Send translation task
				pendingTranslationQueue <- DataWithId[string]{
					Data: transcribedItem.Data.Text,
					Id:   transcribedItem.Id,
				}
			case translatedItems := <-translatedQueue:
				// Update subtitle task info
				processPct += taskWeight * TRANSLATE_WEIGHT
				stepParam.TaskPtr.ProcessPct = uint8(processPct)
				// Process translation results, save raw subtitles without timestamps
				originNoTsSrtFileName := filepath.Join(stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitSrtNoTimestampFileNamePattern, translatedItems.Id))
				originNoTsSrtFile, err := os.Create(originNoTsSrtFileName)
				if err != nil {
					return fmt.Errorf("audioToSubtitle audioToSrt create srt file err: %w", err)
				}
				// Save raw subtitles without timestamps
				for i, translatedItem := range translatedItems.Data {
					// if util.IsAsianLanguage(stepParam.TargetLanguage) {
					// 	translatedItem.TranslatedText = util.BeautifyAsianLanguageSentence(translatedItem.TranslatedText)
					// }
					// if util.IsAsianLanguage(stepParam.OriginLanguage) {
					// 	translatedItem.OriginText = util.BeautifyAsianLanguageSentence(translatedItem.OriginText)
					// }
					_, _ = originNoTsSrtFile.WriteString(fmt.Sprintf("%d\n", i+1))
					_, _ = originNoTsSrtFile.WriteString(fmt.Sprintf("%s\n", translatedItem.TranslatedText))
					_, _ = originNoTsSrtFile.WriteString(fmt.Sprintf("%s\n\n", translatedItem.OriginText))
				}

				// Fix for unknown issue where file is not created
				originNoTsSrtFile.Sync()
				originNoTsSrtFile.Close()
				audioSegments[translatedItems.Id].SrtNoTsFile = originNoTsSrtFileName
				// Generate timestamps
				var srtBlocks []*util.SrtBlock
				for i, translatedItem := range translatedItems.Data {
					srtBlocks = append(srtBlocks, &util.SrtBlock{
						Index:                  i + 1,
						Timestamp:              "",
						OriginLanguageSentence: translatedItem.OriginText,
						TargetLanguageSentence: translatedItem.TranslatedText,
					})
				}

				segmentIdx := translatedItems.Id

				err = generateSrtWithTimestamps(srtBlocks, timePoints[segmentIdx], audioSegments[segmentIdx].TranscriptionData.Words, segmentIdx, stepParam)
				if err != nil {
					return fmt.Errorf("audioToSubtitle audioToSrt generateTimestamps err: %w", err)
				}
				completedTasks++
				// Split, transcribe, and translate tasks completed
				if completedTasks >= segmentNum {
					close(pendingSplitQueue)
					close(splitResultQueue)
					close(pendingTranscriptionQueue)
					close(transcribedQueue)
					close(pendingTranslationQueue)
					close(translatedQueue)
					return nil
				}
			}
		}
	})

	if err := eg.Wait(); err != nil {
		log.GetLogger().Error("audioToSubtitle audioToSrt errgroup wait err", zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle audioToSrt errgroup wait err: %w", err)
	}

	// Merge files
	originNoTsFiles := make([]string, 0)
	bilingualFiles := make([]string, 0)
	shortOriginMixedFiles := make([]string, 0)
	shortOriginFiles := make([]string, 0)
	for i := range segmentNum {
		splitOriginNoTsFile := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitSrtNoTimestampFileNamePattern, i))
		originNoTsFiles = append(originNoTsFiles, splitOriginNoTsFile)
		splitBilingualFile := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitBilingualSrtFileNamePattern, i))
		bilingualFiles = append(bilingualFiles, splitBilingualFile)
		shortOriginMixedFile := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitShortOriginMixedSrtFileNamePattern, i))
		shortOriginMixedFiles = append(shortOriginMixedFiles, shortOriginMixedFile)
		shortOriginFile := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitShortOriginSrtFileNamePattern, i))
		shortOriginFiles = append(shortOriginFiles, shortOriginFile)
	}

	// Merge raw subtitles without timestamps
	originNoTsFile := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, types.SubtitleTaskSrtNoTimestampFileName)
	err = util.MergeFile(originNoTsFile, originNoTsFiles...)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle audioToSrt merge originNoTsFile err",
			zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle audioToSrt merge originNoTsFile err: %w", err)
	}

	// Merge final bilingual subtitles
	bilingualFile := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, types.SubtitleTaskBilingualSrtFileName)
	err = util.MergeSrtFiles(bilingualFile, bilingualFiles...)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle audioToSrt merge BilingualFile err",
			zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle audioToSrt merge BilingualFile err: %w", err)
	}

	// Merge final bilingual subtitles (Long Target + Short Origin)
	shortOriginMixedFile := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, types.SubtitleTaskShortOriginMixedSrtFileName)
	err = util.MergeSrtFiles(shortOriginMixedFile, shortOriginMixedFiles...)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle audioToSrt merge shortOriginMixedFile err",
			zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSrt merge shortOriginMixedFile err: %w", err)
	}
	stepParam.ShortOriginMixedSrtFilePath = shortOriginMixedFile

	// Merge final original subtitles (Short origin)
	shortOriginFile := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, types.SubtitleTaskShortOriginSrtFileName)
	err = util.MergeSrtFiles(shortOriginFile, shortOriginFiles...)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle audioToSrt mergeShortOriginFile err",
			zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSrt mergeShortOriginFile err: %w", err)
	}

	// For subsequent monolingual splitting
	stepParam.BilingualSrtFilePath = bilingualFile

	// Update subtitle task info
	stepParam.TaskPtr.ProcessPct = 90

	log.GetLogger().Info("audioToSubtitle.audioToSrt end", zap.Any("taskId", stepParam.TaskId))

	return nil
}

func splitSrt(stepParam *types.SubtitleTaskStepParam) error {
	log.GetLogger().Info("audioToSubtitle.splitSrt start", zap.Any("task id", stepParam.TaskId))

	originLanguageSrtFilePath := filepath.Join(stepParam.TaskBasePath, types.SubtitleTaskOriginLanguageSrtFileName)
	originLanguageTextFilePath := filepath.Join(stepParam.TaskBasePath, "output", types.SubtitleTaskOriginLanguageTextFileName)
	targetLanguageSrtFilePath := filepath.Join(stepParam.TaskBasePath, types.SubtitleTaskTargetLanguageSrtFileName)
	targetLanguageTextFilePath := filepath.Join(stepParam.TaskBasePath, "output", types.SubtitleTaskTargetLanguageTextFileName)
	// Open bilingual subtitle file
	file, err := os.Open(stepParam.BilingualSrtFilePath)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle splitSrt open bilingual srt file error", zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle splitSrt open bilingual srt file error: %w", err)
	}
	defer file.Close()

	// Open output subtitle and script files
	originLanguageSrtFile, err := os.Create(originLanguageSrtFilePath)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle splitSrt create originLanguageSrtFile error", zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle splitSrt create originLanguageSrtFile error: %w", err)
	}
	defer originLanguageSrtFile.Close()

	originLanguageTextFile, err := os.Create(originLanguageTextFilePath)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle splitSrt create originLanguageTextFile error", zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle splitSrt create originLanguageTextFile error: %w", err)
	}
	defer originLanguageTextFile.Close()

	targetLanguageSrtFile, err := os.Create(targetLanguageSrtFilePath)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle.splitSrt create targetLanguageSrtFile error", zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle.splitSrt create targetLanguageSrtFile error: %w", err)
	}
	defer targetLanguageSrtFile.Close()

	targetLanguageTextFile, err := os.Create(targetLanguageTextFilePath)
	if err != nil {
		log.GetLogger().Error("audioToSubtitle.splitSrt create targetLanguageTextFile error", zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle.splitSrt create targetLanguageTextFile error: %w", err)
	}
	defer targetLanguageTextFile.Close()

	isTargetOnTop := stepParam.SubtitleResultType == types.SubtitleResultTypeBilingualTranslationOnTop

	scanner := bufio.NewScanner(file)
	var block []string

	for scanner.Scan() {
		line := scanner.Text()
		// Empty line represents the end of a subtitle block
		if line == "" {
			if len(block) > 0 {
				util.ProcessBlock(block, targetLanguageSrtFile, targetLanguageTextFile, originLanguageSrtFile, originLanguageTextFile, isTargetOnTop)
				block = nil
			}
		} else {
			block = append(block, line)
		}
	}
	// Process subtitle block at end of file
	if len(block) > 0 {
		util.ProcessBlock(block, targetLanguageSrtFile, targetLanguageTextFile, originLanguageSrtFile, originLanguageTextFile, isTargetOnTop)
	}

	if err = scanner.Err(); err != nil {
		log.GetLogger().Error("audioToSubtitle splitSrt scan bilingual srt file error", zap.Any("taskId", stepParam.TaskId), zap.Error(err))
		return fmt.Errorf("audioToSubtitle splitSrt scan bilingual srt file error: %w", err)
	}
	// Add original language monolingual subtitle
	subtitleInfo := types.SubtitleFileInfo{
		Path:               originLanguageSrtFilePath,
		LanguageIdentifier: string(stepParam.OriginLanguage),
	}
	} else if stepParam.UserUILanguage == types.LanguageNameSimplifiedChinese {
		subtitleInfo.Name = types.GetStandardLanguageName(stepParam.OriginLanguage) + " Monolingual Subtitle"
	}
	stepParam.SubtitleInfos = append(stepParam.SubtitleInfos, subtitleInfo)
	// Add target language monolingual subtitle
	if stepParam.SubtitleResultType == types.SubtitleResultTypeTargetOnly || stepParam.SubtitleResultType == types.SubtitleResultTypeBilingualTranslationOnBottom || stepParam.SubtitleResultType == types.SubtitleResultTypeBilingualTranslationOnTop {
		subtitleInfo = types.SubtitleFileInfo{
			Path:               targetLanguageSrtFilePath,
			LanguageIdentifier: string(stepParam.TargetLanguage),
		}
		if stepParam.UserUILanguage == types.LanguageNameEnglish {
			subtitleInfo.Name = types.GetStandardLanguageName(stepParam.TargetLanguage) + " Subtitle"
		} else if stepParam.UserUILanguage == types.LanguageNameSimplifiedChinese {
			subtitleInfo.Name = types.GetStandardLanguageName(stepParam.TargetLanguage) + " Monolingual Subtitle"
		}
		stepParam.SubtitleInfos = append(stepParam.SubtitleInfos, subtitleInfo)
	}
	// Add bilingual subtitle
	if stepParam.SubtitleResultType == types.SubtitleResultTypeBilingualTranslationOnTop || stepParam.SubtitleResultType == types.SubtitleResultTypeBilingualTranslationOnBottom {
		subtitleInfo = types.SubtitleFileInfo{
			Path:               stepParam.BilingualSrtFilePath,
			LanguageIdentifier: "bilingual",
		}
		if stepParam.UserUILanguage == types.LanguageNameEnglish {
			subtitleInfo.Name = "Bilingual Subtitle"
		} else if stepParam.UserUILanguage == types.LanguageNameSimplifiedChinese {
			subtitleInfo.Name = "Bilingual Subtitle"
		}
		stepParam.SubtitleInfos = append(stepParam.SubtitleInfos, subtitleInfo)
	}

	// For generating voiceover
	stepParam.TtsSourceFilePath = stepParam.BilingualSrtFilePath

	log.GetLogger().Info("audioToSubtitle.splitSrt end", zap.Any("task id", stepParam.TaskId))
	return nil
}

func getSentenceTimestamps(words []types.Word, sentence string, lastTs float64, language types.StandardLanguageCode) (types.SrtSentence, []types.Word, float64, error) {
	var srtSt types.SrtSentence
	var sentenceWordList []string
	sentenceWords := make([]types.Word, 0)
	if language == types.LanguageNameEnglish || language == types.LanguageNameGerman || language == types.LanguageNameTurkish || language == types.LanguageNameRussian { // Different processing method
		sentenceWordList = util.SplitSentence(sentence)
		if len(sentenceWordList) == 0 {
			return srtSt, sentenceWords, 0, fmt.Errorf("getSentenceTimestamps sentence is empty")
		}

		thisLastTs := lastTs
		sentenceWordIndex := 0
		wordNow := words[sentenceWordIndex]
		for _, sentenceWord := range sentenceWordList {
			for sentenceWordIndex < len(words) {
				for sentenceWordIndex < len(words) && !strings.EqualFold(words[sentenceWordIndex].Text, sentenceWord) {
					sentenceWordIndex++
				}

				if sentenceWordIndex >= len(words) {
					break
				}

				wordNow = words[sentenceWordIndex]
				if wordNow.Start < thisLastTs {
					sentenceWordIndex++
					continue
				} else {
					break
				}
			}

			if sentenceWordIndex >= len(words) {
				sentenceWords = append(sentenceWords, types.Word{
					Text: sentenceWord,
				})
				sentenceWordIndex = 0
				continue
			}

			sentenceWords = append(sentenceWords, wordNow)
			sentenceWordIndex = 0
		}
		beginWordIndex, endWordIndex := findMaxIncreasingSubArray(sentenceWords)
		if (endWordIndex - beginWordIndex) == 0 {
			return srtSt, sentenceWords, 0, errors.New("getSentenceTimestamps no valid sentence")
		}

		// Find entire sentence start/end timestamps after finding max contiguous subarray
		beginWord := sentenceWords[beginWordIndex]
		endWord := sentenceWords[endWordIndex-1]
		if endWordIndex-beginWordIndex == len(sentenceWords) {
			srtSt.Start = beginWord.Start
			srtSt.End = endWord.End
			thisLastTs = endWord.End
			return srtSt, sentenceWords, thisLastTs, nil
		}

		if beginWordIndex > 0 {
			for i, j := beginWordIndex-1, beginWord.Num-1; i >= 0 && j >= 0; {
				if words[j].Text == "" {
					j--
					continue
				}
				if strings.EqualFold(words[j].Text, sentenceWords[i].Text) {
					beginWord = words[j]
					sentenceWords[i] = beginWord
				} else {
					break
				}

				i--
				j--
			}
		}

		if endWordIndex < len(sentenceWords) {
			for i, j := endWordIndex, endWord.Num+1; i < len(sentenceWords) && j < len(words); {
				if words[j].Text == "" {
					j++
					continue
				}
				if strings.EqualFold(words[j].Text, sentenceWords[i].Text) {
					endWord = words[j]
					sentenceWords[i] = endWord
				} else {
					break
				}

				i++
				j++
			}
		}

		if beginWord.Num > sentenceWords[0].Num && beginWord.Num-sentenceWords[0].Num < 10 {
			beginWord = sentenceWords[0]
		}

		if sentenceWords[len(sentenceWords)-1].Num > endWord.Num && sentenceWords[len(sentenceWords)-1].Num-endWord.Num < 10 {
			endWord = sentenceWords[len(sentenceWords)-1]
		}

		srtSt.Start = beginWord.Start
		if srtSt.Start < thisLastTs {
			srtSt.Start = thisLastTs
		}
		srtSt.End = endWord.End
		if beginWord.Num != endWord.Num && endWord.End > thisLastTs {
			thisLastTs = endWord.End
		}

		return srtSt, sentenceWords, thisLastTs, nil
	} else {
		sentenceWordList = strings.Split(util.GetRecognizableString(sentence), "")
		if len(sentenceWordList) == 0 {
			return srtSt, sentenceWords, 0, errors.New("getSentenceTimestamps sentence is empty")
		}

		// Note: sentenceWords may not be strictly contiguous and could contain duplicates. Use readableSentenceWords for readable contiguous text.
		var readableSentenceWords []types.Word
		thisLastTs := lastTs
		sentenceWordIndex := 0
		wordNow := words[sentenceWordIndex]
		for _, sentenceWord := range sentenceWordList {
			for sentenceWordIndex < len(words) {
				if !strings.EqualFold(words[sentenceWordIndex].Text, sentenceWord) && !strings.HasPrefix(words[sentenceWordIndex].Text, sentenceWord) {
					sentenceWordIndex++
				} else {
					wordNow = words[sentenceWordIndex]
					if wordNow.Start >= thisLastTs {
						// Record it, but keep searching further.
						sentenceWords = append(sentenceWords, wordNow)
					}
					sentenceWordIndex++
				}
			}
			// Finished searching for current sentenceWord.
			sentenceWordIndex = 0

		}
		// Attempted to find []Word for each word in the sentence.
		var beginWordIndex, endWordIndex int
		beginWordIndex, endWordIndex, readableSentenceWords = jumpFindMaxIncreasingSubArray(sentenceWords)
		if (endWordIndex - beginWordIndex) == 0 {
			return srtSt, readableSentenceWords, 0, errors.New("getSentenceTimestamps no valid sentence")
		}

		beginWord := sentenceWords[beginWordIndex]
		endWord := sentenceWords[endWordIndex]
		//sequence := util.FindClosestConsecutiveWords(words, sentence)
		//beginWord := sequence[0]
		//endWord := sequence[len(sequence)-1]

		srtSt.Start = beginWord.Start
		if srtSt.Start < thisLastTs {
			srtSt.Start = thisLastTs
		}
		srtSt.End = endWord.End
		if beginWord.Num != endWord.Num && endWord.End > thisLastTs {
			thisLastTs = endWord.End
		}

		return srtSt, readableSentenceWords, thisLastTs, nil
	}
}

// Find the maximum contiguous subarray with increasing Num values.
func findMaxIncreasingSubArray(words []types.Word) (int, int) {
	if len(words) == 0 {
		return 0, 0
	}

	// Record the start index and length of the current maximum increasing subarray.
	maxStart, maxLen := 0, 1
	// Record the start index and length of the current increasing subarray.
	currStart, currLen := 0, 1

	for i := 1; i < len(words); i++ {
		if words[i].Num == words[i-1].Num+1 {
			// Current element is greater than the previous one, continue increasing sequence.
			currLen++
		} else {
			// Increasing sequence ends, check if it's the longest.
			if currLen > maxLen {
				maxStart = currStart
				maxLen = currLen
			}
			// Start a new increasing sequence.
			currStart = i
			currLen = 1
		}
	}

	// Perform one final check as the max subarray may be at the end.
	if currLen > maxLen {
		maxStart = currStart
		maxLen = currLen
	}

	// Return the maximum increasing subarray.
	return maxStart, maxStart + maxLen
}

// Find the maximum increasing subarray (non-contiguous) using Num values.
func jumpFindMaxIncreasingSubArray(words []types.Word) (int, int, []types.Word) {
	if len(words) == 0 {
		return -1, -1, nil
	}

	if len(words) == 1 {
		return 0, 0, words
	}

	// dp[i] represents the length of the increasing subarray ending at words[i].
	dp := make([]int, len(words))
	// prev[i] records the index of the previous element in the increasing subarray.
	prev := make([]int, len(words))

	// Initialization: all dp[i] are 1, as each element is a subarray of length 1.
	for i := range len(words) {
		dp[i] = 1
		prev[i] = -1
	}

	maxLen := 0
	startIdx := -1
	endIdx := -1

	// Iterate through each element.
	for i := 1; i < len(words); i++ {
		// Compare each element with previous ones to check for increasing subarray.
		for j := 0; j < i; j++ {
			if words[i].Num == words[j].Num+1 {
				if dp[i] < dp[j]+1 {
					dp[i] = dp[j] + 1
					prev[i] = j
				}
			}
		}

		// Update maximum subarray length and index.
		if dp[i] > maxLen {
			maxLen = dp[i]
			endIdx = i
		}
	}

	// Return if no increasing subarray is found.
	if endIdx == -1 {
		return -1, -1, nil
	}

	// Backtrack to find the starting index of the subarray.
	startIdx = endIdx
	for prev[startIdx] != -1 {
		startIdx = prev[startIdx]
	}

	// Construct the resulting subarray.
	result := make([]types.Word, 0, maxLen)
	current := endIdx
	for current != -1 {
		result = append(result, words[current])
		current = prev[current]
	}

	// Reverse the subarray as it was constructed backwards.
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return startIdx, endIdx, result
}

func generateSrtWithTimestamps(srtBlocks []*util.SrtBlock, tsOffset float64, words []types.Word, segmentIdx int, stepParam *types.SubtitleTaskStepParam) error {
	if len(srtBlocks) == 0 || len(words) == 0 {
		return nil
	}

	// Retrieve timestamps for each subtitle block.
	var lastTs float64
	shortOriginSrtMap := make(map[int][]util.SrtBlock, 0)
	timeMatcher := NewTimestampGenerator()
	newSrtBlocks, err := timeMatcher.GenerateTimestamps(srtBlocks, words, stepParam.OriginLanguage, tsOffset)
	if err != nil {
		return fmt.Errorf("audioToSubtitle generateTimestamps GenerateTimestamps error: %w", err)
	}

	for _, srtBlock := range srtBlocks {
		if srtBlock.OriginLanguageSentence == "" {
			continue
		}
		sentenceTs, sentenceWords, ts, err := getSentenceTimestamps(words, srtBlock.OriginLanguageSentence, lastTs, stepParam.OriginLanguage)
		if err != nil || ts < lastTs {
			continue
		}
		srtBlock.Timestamp = fmt.Sprintf("%s --> %s", util.FormatTime(float32(sentenceTs.Start+tsOffset)), util.FormatTime(float32(sentenceTs.End+tsOffset)))

		// Generate English subtitles for short sentences.
		var (
			originSentence string
			startWord      types.Word
			endWord        types.Word
		)

		if len(sentenceWords) <= stepParam.MaxWordOneLine {
			shortOriginSrtMap[srtBlock.Index] = append(shortOriginSrtMap[srtBlock.Index], util.SrtBlock{
				Index:                  srtBlock.Index,
				Timestamp:              fmt.Sprintf("%s --> %s", util.FormatTime(float32(sentenceTs.Start+tsOffset)), util.FormatTime(float32(sentenceTs.End+tsOffset))),
				OriginLanguageSentence: srtBlock.OriginLanguageSentence,
			})
			lastTs = ts
			continue
		}

		thisLineWord := stepParam.MaxWordOneLine
		if len(sentenceWords) > stepParam.MaxWordOneLine && len(sentenceWords) <= 2*stepParam.MaxWordOneLine {
			thisLineWord = len(sentenceWords)/2 + 1
		} else if len(sentenceWords) > 2*stepParam.MaxWordOneLine && len(sentenceWords) <= 3*stepParam.MaxWordOneLine {
			thisLineWord = len(sentenceWords)/3 + 1
		} else if len(sentenceWords) > 3*stepParam.MaxWordOneLine && len(sentenceWords) <= 4*stepParam.MaxWordOneLine {
			thisLineWord = len(sentenceWords)/4 + 1
		} else if len(sentenceWords) > 4*stepParam.MaxWordOneLine && len(sentenceWords) <= 5*stepParam.MaxWordOneLine {
			thisLineWord = len(sentenceWords)/5 + 1
		}

		i := 1
		nextStart := true
		for _, word := range sentenceWords {
			if nextStart {
				startWord = word
				if startWord.Start < lastTs {
					startWord.Start = lastTs
				}
				if startWord.Start < endWord.End {
					startWord.Start = endWord.End
				}

				if startWord.Start < sentenceTs.Start {
					startWord.Start = sentenceTs.Start
				}
				// First word's start timestamp is later than the sentence end; discarding incorrect word.
				if startWord.End > sentenceTs.End {
					originSentence += word.Text + " "
					continue
				}
				originSentence += word.Text + " "
				endWord = startWord
				i++
				nextStart = false
				continue
			}

			originSentence += word.Text + " "
			if endWord.End < word.End {
				endWord = word
			}

			if endWord.End > sentenceTs.End {
				endWord.End = sentenceTs.End
			}

			if i%thisLineWord == 0 && i > 1 {
				shortOriginSrtMap[srtBlock.Index] = append(shortOriginSrtMap[srtBlock.Index], util.SrtBlock{
					Index:                  srtBlock.Index,
					Timestamp:              fmt.Sprintf("%s --> %s", util.FormatTime(float32(startWord.Start+tsOffset)), util.FormatTime(float32(endWord.End+tsOffset))),
					OriginLanguageSentence: originSentence,
				})
				originSentence = ""
				nextStart = true
			}
			i++
		}

		if originSentence != "" {
			shortOriginSrtMap[srtBlock.Index] = append(shortOriginSrtMap[srtBlock.Index], util.SrtBlock{
				Index:                  srtBlock.Index,
				Timestamp:              fmt.Sprintf("%s --> %s", util.FormatTime(float32(startWord.Start+tsOffset)), util.FormatTime(float32(endWord.End+tsOffset))),
				OriginLanguageSentence: originSentence,
			})
		}
		lastTs = ts
	}

	// Save original subtitles with timestamps.
	finalBilingualSrtFileName := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitBilingualSrtFileNamePattern, segmentIdx))
	finalBilingualSrtFile, err := os.Create(finalBilingualSrtFileName)
	if err != nil {
		return fmt.Errorf("audioToSubtitle generateTimestamps create bilingual srt file error: %w", err)
	}
	defer finalBilingualSrtFile.Close()

	// Write to subtitle file.
	for _, srtBlock := range newSrtBlocks {
		_, _ = finalBilingualSrtFile.WriteString(fmt.Sprintf("%d\n", srtBlock.Index))
		_, _ = finalBilingualSrtFile.WriteString(srtBlock.Timestamp + "\n")
		if stepParam.SubtitleResultType == types.SubtitleResultTypeBilingualTranslationOnTop {
			_, _ = finalBilingualSrtFile.WriteString(srtBlock.TargetLanguageSentence + "\n")
			_, _ = finalBilingualSrtFile.WriteString(srtBlock.OriginLanguageSentence + "\n\n")
		} else {
			// on bottom or monolingual type, both use on bottom
			_, _ = finalBilingualSrtFile.WriteString(srtBlock.OriginLanguageSentence + "\n")
			_, _ = finalBilingualSrtFile.WriteString(srtBlock.TargetLanguageSentence + "\n\n")
		}
	}

	// Save subtitles with timestamps (Long Target + Short Origin).
	srtShortOriginMixedFileName := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitShortOriginMixedSrtFileNamePattern, segmentIdx))
	srtShortOriginMixedFile, err := os.Create(srtShortOriginMixedFileName)
	if err != nil {
		return fmt.Errorf("audioToSubtitle generateTimestamps create srtShortOriginMixedFile err: %w", err)
	}
	defer srtShortOriginMixedFile.Close()

	// Save short original subtitles with timestamps.
	srtShortOriginFileName := fmt.Sprintf("%s/%s", stepParam.TaskBasePath, fmt.Sprintf(types.SubtitleTaskSplitShortOriginSrtFileNamePattern, segmentIdx))
	srtShortOriginFile, err := os.Create(srtShortOriginFileName)
	if err != nil {
		return fmt.Errorf("audioToSubtitle generateTimestamps create srtShortOriginFile err: %w", err)
	}
	defer srtShortOriginMixedFile.Close()

	mixedSrtNum := 1
	shortSrtNum := 1
	// Write mixed short origin subtitle file.
	for _, srtBlock := range srtBlocks {
		srtShortOriginMixedFile.WriteString(fmt.Sprintf("%d\n", mixedSrtNum))
		srtShortOriginMixedFile.WriteString(srtBlock.Timestamp + "\n")
		srtShortOriginMixedFile.WriteString(srtBlock.TargetLanguageSentence + "\n\n")
		mixedSrtNum++
		shortOriginSentence := shortOriginSrtMap[srtBlock.Index]
		for _, shortOriginBlock := range shortOriginSentence {
			srtShortOriginMixedFile.WriteString(fmt.Sprintf("%d\n", mixedSrtNum))
			srtShortOriginMixedFile.WriteString(shortOriginBlock.Timestamp + "\n")
			srtShortOriginMixedFile.WriteString(shortOriginBlock.OriginLanguageSentence + "\n\n")
			mixedSrtNum++

			srtShortOriginFile.WriteString(fmt.Sprintf("%d\n", shortSrtNum))
			srtShortOriginFile.WriteString(shortOriginBlock.Timestamp + "\n")
			srtShortOriginFile.WriteString(shortOriginBlock.OriginLanguageSentence + "\n\n")
			shortSrtNum++
		}
	}

	return nil
}

func parseAndCheckContent(splitContent, originalText string) ([]*TranslatedItem, error) {
	var result []*TranslatedItem

	// Handle empty content scenarios.
	if splitContent == "" || originalText == "" {
		if splitContent == originalText {
			return result, nil
		} else if splitContent == "" {
			return nil, fmt.Errorf("splitContent is empty but originalText is not, originalText: " + originalText)
		} else {
			return nil, errors.New("originalText is empty but splitContent is not, splitContent: " + splitContent)
		}
	}

	// Handle [no_text] markers
	if strings.Contains(splitContent, "[no_text]") {
		// Check if source text is a music marker or similar
		lowerOriginal := strings.ToLower(strings.TrimSpace(originalText))
		if len(lowerOriginal) < 30 && (strings.Contains(lowerOriginal, "music") ||
			strings.Contains(lowerOriginal, "playing") ||
			strings.Contains(lowerOriginal, "♪") ||
			strings.Contains(lowerOriginal, "♫") ||
			len(lowerOriginal) < 10) {
			// If source text is a music marker or very short, return empty result
			return result, nil
		} else {
			// Log warning but allow processing to continue
			log.GetLogger().Warn("originalText might contain actual content but splitContent contains [no_text]",
				zap.String("originalText", originalText),
				zap.String("splitContent", splitContent))
			return result, nil
		}
	}

	lines := strings.Split(splitContent, "\n")
	if len(lines) < 3 { // Requires at least one complete block
		log.GetLogger().Error("audioToSubtitle invaild Format, not enough lines", zap.Any("splitContent", splitContent))
		return nil, fmt.Errorf("audioToSubtitle invaild Format, not enough lines")
	}

	// Validate format and extract original text.
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Check if it's a sequence number line.
		if _, err := strconv.Atoi(line); err != nil {
			continue
		}

		if i+2 >= len(lines) {
			log.GetLogger().Error("audioToSubtitle invaild Format, block is not complete", zap.Any("splitContent", splitContent), zap.Any("line", line))
			return nil, fmt.Errorf("audioToSubtitle invaild Format, block is not complete")
		}
		// Get translation and original lines.
		translatedLine := strings.TrimSpace(lines[i+1])
		translatedLine = strings.TrimPrefix(translatedLine, "[")
		translatedLine = strings.TrimSuffix(translatedLine, "]")
		originalLine := strings.TrimSpace(lines[i+2])
		originalLine = strings.TrimPrefix(originalLine, "[")
		originalLine = strings.TrimSuffix(originalLine, "]")
		result = append(result, &TranslatedItem{
			OriginText:     originalLine,
			TranslatedText: translatedLine,
		})
		i += 2 // Skip translation and original lines.
	}

	// Combine original text and compare character count.
	combinedLength := 0
	for _, translatedItem := range result {
		combinedLength += len(strings.TrimSpace(translatedItem.OriginText))
	}
	originalTextLength := len(strings.TrimSpace(originalText))

	lenError := originalTextLength - combinedLength
	if lenError < 0 {
		lenError = -lenError
	}

	if lenError > len(originalText)/10 {
		// return nil, fmt.Errorf("audioToSubtitle invaild Format, originalText and splitContent length not match", zap.Any("splitContent", splitContent), zap.Any("originalText", originalText))
		log.GetLogger().Warn("audioToSubtitle invaild Format, originalText and splitContent length not match", zap.Any("splitContent", splitContent), zap.Any("originalText", originalText))
	}
	return result, nil
}

// calcLength calculates visual length of text.
func calcLength(text string) float64 {
	var length float64
	for _, r := range text {
		code := r
		switch {
		case (code >= 0x4E00 && code <= 0x9FFF) || (code >= 0x3040 && code <= 0x30FF): // CJK
			length += 1.75
		case (code >= 0xAC00 && code <= 0xD7A3) || (code >= 0x1100 && code <= 0x11FF): // Korean
			length += 1.5
		case code >= 0x0E00 && code <= 0x0E7F: // Thai
			length += 1
		case code >= 0xFF01 && code <= 0xFF5E: // Full-width symbols
			length += 1.75
		default: // Other characters (Latin, etc.)
			length += 1
		}
	}
	return length
}

// splitTranslateItem splits long sentences based on character weights and max length.
func (s Service) splitTranslateItem(items []*TranslatedItem) ([]*TranslatedItem, error) {
	var result []*TranslatedItem
	maxLength := config.Conf.App.MaxSentenceLength + 30

	for _, item := range items {
		// Calculate weighted length of translated text.
		if calcLength(item.OriginText) <= float64(maxLength) && calcLength(item.TranslatedText) <= float64(maxLength) {
			result = append(result, item)
			continue
		}

		// Use LLM to split.
		log.GetLogger().Info("splitTranslateItem long sentence detected, need split", zap.Any("item", item))
		splitItems, err := s.splitLongSentence(item)
		if err != nil {
			log.GetLogger().Error("splitTranslateItem splitLongSentence error", zap.Error(err), zap.Any("item", item))
			return nil, fmt.Errorf("split long sentence error: %w", err)
		}
		result = append(result, splitItems...)
	}

	return result, nil
}

// splitLongSentence uses LLM to split long sentences while maintaining alignment.
func (s Service) splitLongSentence(item *TranslatedItem) ([]*TranslatedItem, error) {
	prompt := fmt.Sprintf(types.SplitLongSentencePrompt, item.OriginText, item.TranslatedText)

	response, err := s.ChatCompleter.ChatCompletion(prompt)
	if err != nil {
		return nil, fmt.Errorf("chat completion error: %w", err)
	}

	var splitResult struct {
		Align []struct {
			OriginPart     string `json:"origin_part"`
			TranslatedPart string `json:"translated_part"`
		} `json:"align"`
	}
	if err := json.Unmarshal([]byte(util.CleanMarkdownCodeBlock(response)), &splitResult); err != nil {
		log.GetLogger().Error("splitLongSentence parse split result error", zap.Error(err), zap.Any("response", response))
		return nil, fmt.Errorf("parse split result error: %w", err)
	}

	// Convert to TranslatedItem slice.
	var splitItems []*TranslatedItem
	for _, part := range splitResult.Align {
		splitItems = append(splitItems, &TranslatedItem{
			OriginText:     part.OriginPart,
			TranslatedText: part.TranslatedPart,
		})
	}

	return splitItems, nil
}

func (s Service) splitOriginLongSentence(sentence string) ([]string, error) {
	prompt := fmt.Sprintf(types.SplitOriginLongSentencePrompt, sentence)
	if len(sentence) > 200 {
		prompt = fmt.Sprintf(types.SplitLongTextByMeaningPrompt, sentence)
	}

	var response string
	var err error
	shortSentences := make([]string, 0)
	// Try calling 3 times.
	for i := range 3 {
		response, err = s.ChatCompleter.ChatCompletion(prompt)
		if err != nil {
			log.GetLogger().Error("splitOriginLongSentence chat completion error", zap.Error(err), zap.String("sentence", sentence), zap.Any("time", i))
			continue
		}
		var splitResult struct {
			ShortSentences []struct {
				Text string `json:"text"`
			} `json:"short_sentences"`
		}

		cleanResponse := util.CleanMarkdownCodeBlock(response)
		if err = json.Unmarshal([]byte(cleanResponse), &splitResult); err != nil {
			log.GetLogger().Error("splitOriginLongSentence parse split result error", zap.Error(err), zap.Any("response", response))
			continue
		}

		for _, shortSentence := range splitResult.ShortSentences {
			shortSentences = append(shortSentences, shortSentence.Text)
		}
		break
	}

	if err != nil {
		return nil, fmt.Errorf("parse split result error: %w", err)
	}

	return shortSentences, nil
}

// splitSentenceRecursively splits sentences recursively while maintaining order.
func (s Service) splitSentenceRecursively(sentence string, depth int, maxDepth int) ([]string, error) {
	// Prevent infinite recursion.
	if depth >= maxDepth {
		log.GetLogger().Warn("reached max split depth", zap.Any("sentence", sentence), zap.Int("depth", depth))
		return []string{sentence}, nil
	}

	// Return if sentence already meets length requirements.
	if util.CountEffectiveChars(sentence) <= config.Conf.App.MaxSentenceLength {
		return []string{sentence}, nil
	}

	// Use LLM to split.
	log.GetLogger().Info("use llm split origin long sentence", zap.Any("sentence", sentence), zap.Int("depth", depth))
	splitItems, err := s.splitOriginLongSentence(sentence)
	if err != nil {
		log.GetLogger().Error("splitSentenceRecursively splitLongSentence error", zap.Error(err), zap.Any("sentence", sentence), zap.Int("depth", depth))
		return []string{sentence}, nil // Return original sentence instead of error.
	}

	// Return original sentence if no splits occurred.
	if len(splitItems) <= 1 {
		return []string{sentence}, nil
	}

	// Recursively process each split result while maintaining order.
	var result []string
	for _, item := range splitItems {
		subResults, err := s.splitSentenceRecursively(item, depth+1, maxDepth)
		if err != nil {
			log.GetLogger().Error("splitSentenceRecursively recursive error", zap.Error(err), zap.Any("item", item), zap.Int("depth", depth))
			result = append(result, item) // If recursion fails, add the original item
		} else {
			result = append(result, subResults...)
		}
	}

	return result, nil
}

//func beautifyTranslateItems(language types.StandardLanguageCode, items []*TranslatedItem) {
//	if language != types.LanguageNameSimplifiedChinese && language != types.LanguageNameTraditionalChinese {
//		return
//	}
//	for _, item := range items {
//
//	}
//}
