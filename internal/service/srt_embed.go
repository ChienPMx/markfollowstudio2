package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"krillin-ai/internal/storage"
	"krillin-ai/internal/types"
	"krillin-ai/log"
	"krillin-ai/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

func (s Service) embedSubtitles(ctx context.Context, stepParam *types.SubtitleTaskStepParam) error {
	var err error
	if stepParam.EmbedSubtitleVideoType == "horizontal" || stepParam.EmbedSubtitleVideoType == "vertical" || stepParam.EmbedSubtitleVideoType == "all" {
		var width, height int
		width, height, err = getResolution(stepParam.InputVideoPath)
		if err != nil {
			log.GetLogger().Error("embedSubtitles getResolution error", zap.Any("step param", stepParam), zap.Error(err))
			return fmt.Errorf("embedSubtitles getResolution error: %w", err)
		}

		// Landscape can be converted to portrait, but portrait to landscape is not supported yet.
		if stepParam.EmbedSubtitleVideoType == "horizontal" || stepParam.EmbedSubtitleVideoType == "all" {
			if width < height {
				log.GetLogger().Info("Input video is vertical, cannot create landscape video, skipping...")
				return nil
			}
			log.GetLogger().Info("Generating video: Landscape")
			err = embedSubtitles(stepParam, true, stepParam.EnableTts)
			if err != nil {
				log.GetLogger().Error("embedSubtitles embedSubtitles error", zap.Any("step param", stepParam), zap.Error(err))
				return fmt.Errorf("embedSubtitles embedSubtitles error: %w", err)
			}
		}
		if stepParam.EmbedSubtitleVideoType == "vertical" || stepParam.EmbedSubtitleVideoType == "all" {
			if width > height {
				// Generate portrait video
				transferredVerticalVideoPath := filepath.Join(stepParam.TaskBasePath, types.SubtitleTaskTransferredVerticalVideoFileName)
				err = convertToVertical(stepParam.InputVideoPath, transferredVerticalVideoPath, stepParam.VerticalVideoMajorTitle, stepParam.VerticalVideoMinorTitle)
				if err != nil {
					log.GetLogger().Error("embedSubtitles convertToVertical error", zap.Any("step param", stepParam), zap.Error(err))
					return fmt.Errorf("embedSubtitles convertToVertical error: %w", err)
				}
				stepParam.InputVideoPath = transferredVerticalVideoPath
			}
			log.GetLogger().Info("Generating video: Portrait")
			err = embedSubtitles(stepParam, false, stepParam.EnableTts)
			if err != nil {
				log.GetLogger().Error("embedSubtitles embedSubtitles error", zap.Any("step param", stepParam), zap.Error(err))
				return fmt.Errorf("embedSubtitles embedSubtitles error: %w", err)
			}
		}
		log.GetLogger().Info("Successfully embedded subtitles into video")
		return nil
	}
	log.GetLogger().Info("Generating video: No embedding")
	return nil
}

func splitMajorTextInHorizontal(text string, language types.StandardLanguageCode, maxWordOneLine int) []string {
	// Split based on language
	var (
		segments []string
		sep      string
	)
	if language == types.LanguageNameSimplifiedChinese || language == types.LanguageNameTraditionalChinese ||
		language == types.LanguageNameJapanese || language == types.LanguageNameKorean || language == types.LanguageNameThai {
		segments = regexp.MustCompile(`.`).FindAllString(text, -1)
		sep = ""
	} else {
		segments = strings.Split(text, " ")
		sep = " "
	}

	totalWidth := len(segments)

	// Return original sentence directly
	if totalWidth <= maxWordOneLine {
		return []string{text}
	}

	// Determine split point, splitting by 2/5 and 3/5 ratio.
	line1MaxWidth := int(float64(totalWidth) * 2 / 5)
	currentWidth := 0
	splitIndex := 0

	for i := range segments {
		currentWidth++

		// Set split point when reaching 2/5 width.
		if currentWidth >= line1MaxWidth {
			splitIndex = i + 1
			break
		}
	}

	// Split text while preserving original sentence format.

	line1 := util.CleanPunction(strings.Join(segments[:splitIndex], sep))
	line2 := util.CleanPunction(strings.Join(segments[splitIndex:], sep))

	return []string{line1, line2}
}

func splitChineseText(text string, maxWordLine int) []string {
	var lines []string
	words := []rune(text)
	for i := 0; i < len(words); i += maxWordLine {
		end := i + maxWordLine
		if end > len(words) {
			end = len(words)
		}
		lines = append(lines, string(words[i:end]))
	}
	return lines
}

func parseSrtTime(timeStr string) (time.Duration, error) {
	timeStr = strings.Replace(timeStr, ",", ".", 1)
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("parseSrtTime invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	secondsAndMilliseconds := strings.Split(parts[2], ".")
	if len(secondsAndMilliseconds) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}
	seconds, err := strconv.Atoi(secondsAndMilliseconds[0])
	if err != nil {
		return 0, err
	}
	milliseconds, err := strconv.Atoi(secondsAndMilliseconds[1])
	if err != nil {
		return 0, err
	}

	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second +
		time.Duration(milliseconds)*time.Millisecond

	return duration, nil
}

func formatTimestamp(t time.Duration) string {
	hours := int(t.Hours())
	minutes := int(t.Minutes()) % 60
	seconds := int(t.Seconds()) % 60
	milliseconds := int(t.Milliseconds()) % 1000 / 10
	return fmt.Sprintf("%02d:%02d:%02d.%02d", hours, minutes, seconds, milliseconds)
}

func srtToAss(inputSRT, outputASS string, isHorizontal bool, stepParam *types.SubtitleTaskStepParam) error {
	file, err := os.Open(inputSRT)
	if err != nil {
		log.GetLogger().Error("srtToAss Open input srt error", zap.Error(err))
		return fmt.Errorf("srtToAss Open input srt error: %w", err)
	}
	defer file.Close()

	assFile, err := os.Create(outputASS)
	if err != nil {
		log.GetLogger().Error("srtToAss Create output ass error", zap.Error(err))
		return fmt.Errorf("srtToAss Create output ass error: %w", err)
	}
	defer assFile.Close()
	scanner := bufio.NewScanner(file)

	if isHorizontal {
		_, _ = assFile.WriteString(types.AssHeaderHorizontal)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			// Read timestamp line
			if !scanner.Scan() {
				break
			}
			timestampLine := scanner.Text()
			parts := strings.Split(timestampLine, " --> ")
			if len(parts) != 2 {
				continue // Invalid timestamp format
			}

			startTimeStr := strings.TrimSpace(parts[0])
			endTimeStr := strings.TrimSpace(parts[1])
			startTime, err := parseSrtTime(startTimeStr)
			if err != nil {
				log.GetLogger().Error("srtToAss parseSrtTime error", zap.Error(err))
				return fmt.Errorf("srtToAss parseSrtTime error: %w", err)
			}
			endTime, err := parseSrtTime(endTimeStr)
			if err != nil {
				log.GetLogger().Error("srtToAss parseSrtTime error", zap.Error(err))
				return fmt.Errorf("srtToAss parseSrtTime error: %w", err)
			}

			var subtitleLines []string
			for scanner.Scan() {
				textLine := scanner.Text()
				if textLine == "" {
					break // End of subtitle block
				}
				subtitleLines = append(subtitleLines, textLine)
			}

			if len(subtitleLines) < 2 {
				continue
			}
			//var majorTextLanguage types.StandardLanguageCode
			//if stepParam.SubtitleResultType == types.SubtitleResultTypeBilingualTranslationOnTop { // Must be bilingual
			//	majorTextLanguage = stepParam.TargetLanguage
			//} else {
			//	majorTextLanguage = stepParam.OriginLanguage
			//}

			//majorLine := strings.Join(splitMajorTextInHorizontal(subtitleLines[0], majorTextLanguage, stepParam.MaxWordOneLine), "      \\N")

			// ASS entry
			startFormatted := formatTimestamp(startTime)
			endFormatted := formatTimestamp(endTime)
			combinedText := fmt.Sprintf("{\\an2}{\\rMajor}%s\\N{\\rMinor}%s", subtitleLines[0], util.CleanPunction(subtitleLines[1]))
			_, _ = assFile.WriteString(fmt.Sprintf("Dialogue: 0,%s,%s,Major,,0,0,0,,%s\n", startFormatted, endFormatted, combinedText))
		}
	} else {
		// TODO portrait split tuning
		_, _ = assFile.WriteString(types.AssHeaderVertical)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !scanner.Scan() {
				break
			}
			timestampLine := scanner.Text()
			parts := strings.Split(timestampLine, " --> ")
			if len(parts) != 2 {
				continue // Invalid timestamp format
			}

			startTimeStr := strings.TrimSpace(parts[0])
			endTimeStr := strings.TrimSpace(parts[1])
			startTime, err := parseSrtTime(startTimeStr)
			if err != nil {
				return err
			}
			endTime, err := parseSrtTime(endTimeStr)
			if err != nil {
				return err
			}

			var content string
			scanner.Scan()
			content = scanner.Text()
			if content == "" {
				continue
			}
			totalTime := endTime - startTime

			if !util.ContainsAlphabetic(content) {
				// Handle CJK subtitles
				chineseLines := splitChineseText(content, 10)
				for i, line := range chineseLines {
					iStart := startTime + time.Duration(float64(i)*float64(totalTime)/float64(len(chineseLines)))
					iEnd := startTime + time.Duration(float64(i+1)*float64(totalTime)/float64(len(chineseLines)))
					if iEnd > endTime {
						iEnd = endTime
					}

					startFormatted := formatTimestamp(iStart)
					endFormatted := formatTimestamp(iEnd)
					cleanedText := util.CleanPunction(line)
					combinedText := fmt.Sprintf("{\\an2}{\\rMajor}%s", cleanedText)
					_, _ = assFile.WriteString(fmt.Sprintf("Dialogue: 0,%s,%s,Major,,0,0,0,,%s\n", startFormatted, endFormatted, combinedText))
				}
			} else {
				// Handle Latin subtitles
				startFormatted := formatTimestamp(startTime)
				endFormatted := formatTimestamp(endTime)
				cleanedText := util.CleanPunction(content)
				combinedText := fmt.Sprintf("{\\an2}{\\rMinor}%s", cleanedText)
				_, _ = assFile.WriteString(fmt.Sprintf("Dialogue: 0,%s,%s,Minor,,0,0,0,,%s\n", startFormatted, endFormatted, combinedText))
			}
		}
	}
	return nil
}

func embedSubtitles(stepParam *types.SubtitleTaskStepParam, isHorizontal bool, withTts bool) error {
	outputFileName := types.SubtitleTaskVerticalEmbedVideoFileName
	if isHorizontal {
		outputFileName = types.SubtitleTaskHorizontalEmbedVideoFileName
	}
	assPath := filepath.Join(stepParam.TaskBasePath, "formatted_subtitles.ass")

	if err := srtToAss(stepParam.BilingualSrtFilePath, assPath, isHorizontal, stepParam); err != nil {
		log.GetLogger().Error("embedSubtitles srtToAss error", zap.Any("step param", stepParam), zap.Error(err))
		return fmt.Errorf("embedSubtitles srtToAss error: %w", err)
	}
	input := stepParam.InputVideoPath
	if withTts {
		input = stepParam.VideoWithTtsFilePath
	}

	cmd := exec.Command(storage.FfmpegPath, "-y", "-i", input, "-vf", fmt.Sprintf("ass=%s", strings.ReplaceAll(assPath, "\\", "/")), "-c:a", "aac", "-b:a", "192k", filepath.Join(stepParam.TaskBasePath, fmt.Sprintf("/output/%s", outputFileName)))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.GetLogger().Error("embedSubtitles embed subtitle into video ffmpeg error", zap.String("video path", stepParam.InputVideoPath), zap.String("output", string(output)), zap.Error(err))
		return fmt.Errorf("embedSubtitles embed subtitle into video ffmpeg error: %w", err)
	}
	return nil
}

func getFontPaths() (string, string, error) {
	switch runtime.GOOS {
	case "windows":
		return "C\\:/Windows/Fonts/msyhbd.ttc", "C\\:/Windows/Fonts/msyh.ttc", nil // Must be written like this in ffmpeg parameters
	case "darwin":
		return "/System/Library/Fonts/Supplemental/Arial Bold.ttf", "/System/Library/Fonts/Supplemental/Arial.ttf", nil
	case "linux":
		return "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", nil
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func getResolution(inputVideo string) (int, int, error) {
	// Get video information
	cmdArgs := []string{
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		inputVideo,
	}
	cmd := exec.Command(storage.FfprobePath, cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		log.GetLogger().Error("Failed to get video resolution", zap.String("output", out.String()), zap.Error(err))
		return 0, 0, err
	}

	output := strings.TrimSpace(out.String())
	output = strings.TrimSuffix(output, "x") // Remove trailing 'x', e.g., 1920x1080x

	re := regexp.MustCompile(`^(\d+)x(\d+)$`)
	dimensions := re.FindStringSubmatch(output)
	if len(dimensions) != 3 {
		log.GetLogger().Error("Failed to get video resolution", zap.String("output", output))
		return 0, 0, fmt.Errorf("invalid resolution format: %s", output)
	}

	width, _ := strconv.Atoi(dimensions[1])
	height, _ := strconv.Atoi(dimensions[2])
	return width, height, nil
}

func convertToVertical(inputVideo, outputVideo, majorTitle, minorTitle string) error {
	if _, err := os.Stat(outputVideo); err == nil {
		log.GetLogger().Info("Portrait video already exists", zap.String("outputVideo", outputVideo))
		return nil
	}

	fontBold, fontRegular, err := getFontPaths()
	if err != nil {
		log.GetLogger().Error("Failed to get font paths", zap.Error(err))
		return err
	}

	cmdArgs := []string{
		"-i", inputVideo,
		"-vf", fmt.Sprintf("scale=720:1280:force_original_aspect_ratio=decrease,pad=720:1280:(ow-iw)/2:(oh-ih)*2/5,drawbox=y=0:h=100:c=black@1:t=fill,drawtext=text='%s':x=(w-text_w)/2:y=210:fontsize=55:fontcolor=yellow:box=1:boxcolor=black@0.5:fontfile='%s',drawtext=text='%s':x=(w-text_w)/2:y=280:fontsize=40:fontcolor=yellow:box=1:boxcolor=black@0.5:fontfile='%s'",
			majorTitle, fontBold, minorTitle, fontRegular),
		"-r", "30",
		"-b:v", "7587k",
		"-c:a", "aac",
		"-b:a", "192k",
		"-c:v", "libx264",
		"-preset", "fast",
		"-y",
		outputVideo,
	}
	cmd := exec.Command(storage.FfmpegPath, cmdArgs...)
	var output []byte
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.GetLogger().Error("Failed to convert video to portrait", zap.String("output", string(output)), zap.Error(err))
		return err
	}

	fmt.Printf("Portrait video saved to: %s\n", outputVideo)
	return nil
}
