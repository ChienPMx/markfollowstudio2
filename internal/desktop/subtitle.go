package desktop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"krillin-ai/config"
	"krillin-ai/internal/api"
	"krillin-ai/internal/handler"
	"krillin-ai/log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"go.uber.org/zap"
)

// SubtitleManager subtitle manager
type SubtitleManager struct {
	window             fyne.Window
	handler            *handler.Handler
	videoUrl           string   // Unified field to store video URL (uploaded or input)
	videoPaths         []string // Stores multiple video paths
	audioPath          string
	uploadedAudioURL   string
	sourceLang         string
	targetLang         string
	bilingualEnabled   bool
	bilingualPosition  int
	voiceoverEnabled   bool
	ttsVoiceCode       string // Voice code
	fillerFilter       bool
	enableReview       bool   // Pause pipeline for subtitle review before TTS
	wordReplacements   []api.WordReplacement
	embedSubtitle      string // none, horizontal, vertical, all
	verticalTitles     [2]string
	progressBar        *widget.ProgressBar
	progressLabel      *widget.Label // Progress percentage label
	downloadContainer  *fyne.Container
	tipsLabel          *widget.Label
	onVideoSelected    func(string)
	onVideosSelected   func([]string) // Multi-file selection callback
	onAudioSelected    func(string)
	voiceoverAudioPath  string
	multiTaskResults    []taskResult // Stores multi-task results
	reviewDialogShowing bool         // Whether the review dialog is currently displayed
}

// Used to store result info for each task
type taskResult struct {
	fileName          string // Original file name
	subtitleInfo      []api.SubtitleResult
	speechDownloadURL string
	taskId            string
}

// NewSubtitleManager creates the subtitle manager
func NewSubtitleManager(window fyne.Window) *SubtitleManager {
	return &SubtitleManager{
		window:            window,
		sourceLang:        "en",
		targetLang:        "zh_cn",
		bilingualEnabled:  true,
		bilingualPosition: 1,
		fillerFilter:      true,
		enableReview:      true, // Enable by default
		voiceoverEnabled:  false,
		ttsVoiceCode:      "",
		embedSubtitle:     "none",
		downloadContainer: container.NewVBox(),
		tipsLabel:         widget.NewLabel(""),
		videoPaths:        make([]string, 0),
	}
}

func (sm *SubtitleManager) showReviewDialog(taskId, srtContent string) {
	sm.reviewDialogShowing = true

	// Create multi-line entry for SRT editing
	entry := widget.NewMultiLineEntry()
	entry.SetText(srtContent)
	// Make it reasonably large
	entryContainer := container.NewGridWrap(fyne.NewSize(600, 400), entry)

	var dlg dialog.Dialog
	
	approveBtn := widget.NewButton("Approve & Continue", func() {
		// Send the approved/edited content to the API
		reqBody := map[string]string{
			"task_id":            taskId,
			"edited_srt_content": entry.Text,
		}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to encode review request: %v", err), sm.window)
			return
		}

		resp, err := http.Post(fmt.Sprintf("http://%s:%d/api/capability/subtitleTask/review", config.Conf.Server.Host, config.Conf.Server.Port), "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to submit review: %v", err), sm.window)
			return
		}
		defer resp.Body.Close()

		sm.reviewDialogShowing = false
		dlg.Hide()
	})
	approveBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabel("Review and edit the translated subtitles below before dubbing:"),
		entryContainer,
		container.NewHBox(layout.NewSpacer(), approveBtn, layout.NewSpacer()),
	)

	dlg = dialog.NewCustom("Review Subtitles", "Cancel", content, sm.window)
	dlg.SetOnClosed(func() {
		sm.reviewDialogShowing = false
	})
	dlg.Resize(fyne.NewSize(650, 500))
	dlg.Show()
}

func (sm *SubtitleManager) SetVideoSelectedCallback(callback func(string)) {
	sm.onVideoSelected = callback
}

// Multi-file selection callback
func (sm *SubtitleManager) SetVideosSelectedCallback(callback func([]string)) {
	sm.onVideosSelected = callback
}

func (sm *SubtitleManager) ShowFileDialog() {
	sm.videoPaths = make([]string, 0)

	sm.addVideoFile(false)
}

// addVideoFile adds video files
// continueAdding true means adding more files
func (sm *SubtitleManager) addVideoFile(continueAdding bool) {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, sm.window)
			return
		}
		if reader == nil {
			// User cancelled selection
			if len(sm.videoPaths) > 0 {
				// Ask whether to upload
				confirmDialog := dialog.NewConfirm(
					"Upload Files",
					fmt.Sprintf("Selected %d files. Start uploading?", len(sm.videoPaths)),
					func(confirm bool) {
						if confirm {
							sm.uploadMultipleFiles()
						}
					},
					sm.window)
				confirmDialog.Show()
			}
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()

		sm.videoPaths = append(sm.videoPaths, filePath)

		// Ask whether to continue adding
		// Build selected file message
		filesMessage := fmt.Sprintf("Selected %d files:\n", len(sm.videoPaths))
		for i, path := range sm.videoPaths {
			filesMessage += fmt.Sprintf("%d. %s\n", i+1, filepath.Base(path))
		}
		filesMessage += "\nDo you want to add more files?"

		confirmDialog := dialog.NewConfirm(
			"Continue Selecting",
			filesMessage,
			func(cont bool) {
				if cont {
					// Continue adding files
					sm.addVideoFile(true)
				} else {
					// Start uploading
					sm.uploadMultipleFiles()
				}
			},
			sm.window,
		)
		confirmDialog.Show()
	}, sm.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".mp4", ".mov", ".avi", ".mkv", ".wmv"}))
	fd.Show()
}

// uploadMultipleFiles uploads multiple files
func (sm *SubtitleManager) uploadMultipleFiles() {
	if len(sm.videoPaths) == 0 {
		return
	}

	// Create progress dialog
	filesList := fmt.Sprintf("Uploading %d files:\n", len(sm.videoPaths))
	for i, path := range sm.videoPaths {
		filesList += fmt.Sprintf("%d. %s\n", i+1, filepath.Base(path))
	}

	progressDialog := dialog.NewProgress("Uploading", filesList, sm.window)
	progressDialog.Show()

	go func() {
		defer progressDialog.Hide()

		// Create multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add multiple files to form
		for i, filePath := range sm.videoPaths {
			file, err := os.Open(filePath)
			if err != nil {
				dialog.ShowError(err, sm.window)
				return
			}

			part, err := writer.CreateFormFile("file", filepath.Base(filePath))
			if err != nil {
				file.Close()
				dialog.ShowError(err, sm.window)
				return
			}

			_, err = io.Copy(part, file)
			file.Close()
			if err != nil {
				dialog.ShowError(err, sm.window)
				return
			}

			// Update progress
			progressDialog.SetValue(float64(i+1) / float64(len(sm.videoPaths)))
		}

		err := writer.Close()
		if err != nil {
			dialog.ShowError(err, sm.window)
			return
		}

		// Send request
		resp, err := http.Post(fmt.Sprintf("http://%s:%d/api/file", config.Conf.Server.Host, config.Conf.Server.Port), writer.FormDataContentType(), body)
		if err != nil {
			dialog.ShowError(err, sm.window)
			return
		}
		defer resp.Body.Close()

		var result struct {
			Error int    `json:"error"`
			Msg   string `json:"msg"`
			Data  struct {
				FilePath []string `json:"file_path"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			dialog.ShowError(err, sm.window)
			return
		}

		if result.Error != 0 && result.Error != 200 {
			dialog.ShowError(fmt.Errorf(result.Msg), sm.window)
			return
		}

		tempPaths := make([]string, len(result.Data.FilePath))
		copy(tempPaths, result.Data.FilePath)
		sm.videoPaths = tempPaths

		// If there's only one file, also set videoUrl
		if len(result.Data.FilePath) > 0 {
			sm.videoUrl = result.Data.FilePath[0]
		}

		if sm.onVideosSelected != nil {
			sm.onVideosSelected(result.Data.FilePath)
		} else if sm.onVideoSelected != nil && len(result.Data.FilePath) > 0 {
			sm.onVideoSelected(result.Data.FilePath[0])
		}

		// Build message
		successMessage := fmt.Sprintf("Successfully uploaded %d files:\n", len(result.Data.FilePath))
		for i, url := range result.Data.FilePath {
			successMessage += fmt.Sprintf("%d. %s\n", i+1, filepath.Base(url))
		}

		dialog.ShowInformation("Upload Successful", successMessage, sm.window)
	}()
}

func (sm *SubtitleManager) ShowAudioFileDialog() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, sm.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		tempFile, err := os.CreateTemp("", "audio-*.wav")
		if err != nil {
			dialog.ShowError(err, sm.window)
			return
		}
		defer tempFile.Close()

		_, err = io.Copy(tempFile, reader)
		if err != nil {
			dialog.ShowError(err, sm.window)
			return
		}

		// Set audio path
		sm.voiceoverAudioPath = tempFile.Name()
		if sm.onAudioSelected != nil {
			sm.onAudioSelected(tempFile.Name())
		}
	}, sm.window)
}

func (sm *SubtitleManager) uploadVideo(localPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("Failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return fmt.Errorf("Failed to create form: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("Failed to copy file content: %w", err)
	}
	writer.Close()

	resp, err := http.Post(fmt.Sprintf("http://%s:%d/api/file", config.Conf.Server.Host, config.Conf.Server.Port), writer.FormDataContentType(), body)
	if err != nil {
		return fmt.Errorf("Failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  struct {
			FilePath string `json:"file_path"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("Failed to parse response: %w", err)
	}

	if result.Error != 0 && result.Error != 200 {
		return fmt.Errorf(result.Msg)
	}

	sm.videoUrl = result.Data.FilePath
	return nil
}

func (sm *SubtitleManager) uploadAudio() error {
	file, err := os.Open(sm.audioPath)
	if err != nil {
		return fmt.Errorf("Failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(sm.audioPath))
	if err != nil {
		return fmt.Errorf("Failed to create form: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("Failed to copy file content: %w", err)
	}
	writer.Close()

	resp, err := http.Post(fmt.Sprintf("http://%s:%d/api/file", config.Conf.Server.Host, config.Conf.Server.Port), writer.FormDataContentType(), body)
	if err != nil {
		return fmt.Errorf("Failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  struct {
			FilePath string `json:"file_path"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("Failed to parse response: %w", err)
	}

	if result.Error != 0 && result.Error != 200 {
		return fmt.Errorf(result.Msg)
	}

	sm.uploadedAudioURL = result.Data.FilePath
	return nil
}

func (sm *SubtitleManager) SetSourceLang(lang string) {
	sm.sourceLang = lang
}

func (sm *SubtitleManager) SetTargetLang(lang string) {
	sm.targetLang = lang
}

// SetBilingualEnabled sets whether to enable bilingual subtitles
func (sm *SubtitleManager) SetBilingualEnabled(enabled bool) {
	sm.bilingualEnabled = enabled
}

// SetBilingualPosition sets bilingual subtitle position
func (sm *SubtitleManager) SetBilingualPosition(position int) {
	sm.bilingualPosition = position
}

// SetFillerFilter sets whether to enable filler word filtering
func (sm *SubtitleManager) SetFillerFilter(enabled bool) {
	sm.fillerFilter = enabled
}

// SetVoiceoverEnabled sets whether to enable dubbing
func (sm *SubtitleManager) SetVoiceoverEnabled(enabled bool) {
	sm.voiceoverEnabled = enabled
}

// SetTtsVoiceCode sets voice code
func (sm *SubtitleManager) SetTtsVoiceCode(code string) {
	sm.ttsVoiceCode = code
}

// SetEmbedSubtitle sets subtitle embedding mode
func (sm *SubtitleManager) SetEmbedSubtitle(mode string) {
	sm.embedSubtitle = mode
}

// SetVerticalTitles sets vertical titles
func (sm *SubtitleManager) SetVerticalTitles(mainTitle, subTitle string) {
	sm.verticalTitles = [2]string{mainTitle, subTitle}
}

// SetProgressBar sets the progress bar
func (sm *SubtitleManager) SetProgressBar(progress *widget.ProgressBar) {
	sm.progressBar = progress
}

// SetDownloadContainer sets download container
func (sm *SubtitleManager) SetDownloadContainer(container *fyne.Container) {
	sm.downloadContainer = container
}

// SetTipsLabel sets tips label
func (sm *SubtitleManager) SetTipsLabel(label *widget.Label) {
	sm.tipsLabel = label
}

// SetAudioSelectedCallback sets audio selection callback
func (sm *SubtitleManager) SetAudioSelectedCallback(callback func(string)) {
	sm.onAudioSelected = callback
}

// SetVideoUrl sets video URL
func (sm *SubtitleManager) SetVideoUrl(url string) {
	sm.videoUrl = url
}

// SetEnableReview sets enableReview
func (sm *SubtitleManager) SetEnableReview(enable bool) {
	sm.enableReview = enable
}

// GetVideoUrl gets video URL
func (sm *SubtitleManager) GetVideoUrl() string {
	return sm.videoUrl
}

// SetProgressLabel sets progress percentage label
func (sm *SubtitleManager) SetProgressLabel(label *widget.Label) {
	sm.progressLabel = label
}

// StartTask starts the subtitle task
func (sm *SubtitleManager) StartTask() error {
	// Check if there are multiple video paths to process
	if len(sm.videoPaths) > 1 {
		// Start tasks for multiple videos sequentially
		go sm.processMultipleVideos()
		return nil
	} else if len(sm.videoPaths) == 1 {
		// Ensure the first URL in videoPaths is used
		sm.videoUrl = sm.videoPaths[0]
	}

	// Single video processing
	task := &api.SubtitleTask{
		URL:                     sm.videoUrl,
		Language:                "zh_cn",
		OriginLang:              sm.sourceLang,
		TargetLang:              sm.targetLang,
		Bilingual:               boolToInt(sm.bilingualEnabled),
		TranslationSubtitlePos:  sm.bilingualPosition,
		TTS:                     boolToInt(sm.voiceoverEnabled),
		TTSVoiceCode:            sm.ttsVoiceCode,
		TTSVoiceCloneSrcFileURL: sm.voiceoverAudioPath,
		ModalFilter:             boolToInt(sm.fillerFilter),
		EnableReview:            sm.enableReview,
		EmbedSubtitleVideoType:  sm.embedSubtitle,
		VerticalMajorTitle:      sm.verticalTitles[0],
		VerticalMinorTitle:      sm.verticalTitles[1],
	}

	jsonData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("Failed to serialize task data: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("http://%s:%d/api/capability/subtitleTask", config.Conf.Server.Host, config.Conf.Server.Port), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Failed to send task request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  struct {
			TaskId string `json:"task_id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("Failed to parse response: %w", err)
	}

	if result.Error != 0 && result.Error != 200 {
		return fmt.Errorf(result.Msg)
	}

	// Start polling task status
	go sm.pollTaskStatus(result.Data.TaskId)
	return nil
}

// Process multiple videos
func (sm *SubtitleManager) processMultipleVideos() {
	// Original video URL
	originalURL := sm.videoUrl

	// Clear previous task results
	sm.multiTaskResults = make([]taskResult, 0, len(sm.videoPaths))

	// Reset progress bar
	sm.progressBar.SetValue(0)
	sm.progressBar.Show()

	// Update progress label
	if sm.progressLabel != nil {
		sm.progressLabel.SetText("0%")
		sm.progressLabel.Show()
	}

	// Clear download container
	sm.downloadContainer.Objects = []fyne.CanvasObject{}
	sm.downloadContainer.Hide()

	// Hide tips label
	sm.tipsLabel.Hide()

	go func() {
		for i, url := range sm.videoPaths {
			fileName := filepath.Base(url)

			percentage := float64(i) / float64(len(sm.videoPaths))
			sm.progressBar.SetValue(percentage)

			if sm.progressLabel != nil {
				displayName := fileName
				if len(displayName) > 20 {
					displayName = displayName[:17] + "..."
				}
				sm.progressLabel.SetText(fmt.Sprintf("Processing: %d/%d\n%s", i+1, len(sm.videoPaths), displayName))
				sm.progressLabel.Show()
			}

			sm.videoUrl = url

			task := &api.SubtitleTask{
				URL:                     url,
				Language:                "zh_cn",
				OriginLang:              sm.sourceLang,
				TargetLang:              sm.targetLang,
				Bilingual:               boolToInt(sm.bilingualEnabled),
				TranslationSubtitlePos:  sm.bilingualPosition,
				TTS:                     boolToInt(sm.voiceoverEnabled),
				TTSVoiceCode:            sm.ttsVoiceCode,
				TTSVoiceCloneSrcFileURL: sm.voiceoverAudioPath,
				ModalFilter:             boolToInt(sm.fillerFilter),
				EmbedSubtitleVideoType:  sm.embedSubtitle,
				VerticalMajorTitle:      sm.verticalTitles[0],
				VerticalMinorTitle:      sm.verticalTitles[1],
			}

			jsonData, err := json.Marshal(task)
			if err != nil {
				log.GetLogger().Error("Failed to serialize task data", zap.Error(err))
				continue
			}

			resp, err := http.Post(fmt.Sprintf("http://%s:%d/api/capability/subtitleTask", config.Conf.Server.Host, config.Conf.Server.Port), "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.GetLogger().Error("Failed to send task request", zap.Error(err))
				continue
			}

			var result struct {
				Error int    `json:"error"`
				Msg   string `json:"msg"`
				Data  struct {
					TaskId string `json:"task_id"`
				} `json:"data"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				log.GetLogger().Error("Failed to parse response", zap.Error(err))
				continue
			}
			resp.Body.Close()

			if result.Error != 0 && result.Error != 200 {
				log.GetLogger().Error("Failed to create task", zap.String("msg", result.Msg))
				continue
			}

			taskRes := sm.waitTaskCompleted(result.Data.TaskId, fileName)

			sm.multiTaskResults = append(sm.multiTaskResults, taskRes)
		}

		sm.videoUrl = originalURL

		// Display all file download links
		sm.displayMultiTaskDownloadLinks()

		// Complete all video processing
		dialog.ShowInformation("Done", "All videos processed successfully", sm.window)
	}()
}

// Wait for task completion and return result
func (sm *SubtitleManager) waitTaskCompleted(taskId string, originalFileName string) taskResult {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Last progress percentage
	lastPercent := 0

	// Prepare task result
	res := taskResult{
		fileName: originalFileName,
		taskId:   taskId,
	}

	// Poll task status
	for {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/api/capability/subtitleTask?taskId=%s", config.Conf.Server.Host, config.Conf.Server.Port, taskId))
		if err != nil {
			log.GetLogger().Error("Failed to get task status", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		var result struct {
			Error int    `json:"error"`
			Msg   string `json:"msg"`
			Data  struct {
				ProcessPercent    int                  `json:"process_percent"`
				Status            string               `json:"status"`
				SubtitleInfo      []api.SubtitleResult `json:"subtitle_info"`
				SpeechDownloadURL string               `json:"speech_download_url"`
				TaskId            string               `json:"task_id"`
				ReviewSrtContent  string               `json:"review_srt_content"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.GetLogger().Error("Failed to parse response", zap.Error(err))
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()

		if result.Data.ProcessPercent != lastPercent {
			progress := float64(result.Data.ProcessPercent) / 100.0
			sm.progressBar.SetValue(progress)

			// Update progress label
			if sm.progressLabel != nil {
				sm.progressLabel.SetText(fmt.Sprintf("%d%%", result.Data.ProcessPercent))
				sm.progressLabel.Show()
			}

			// Update last progress percentage
			lastPercent = result.Data.ProcessPercent
		}

		if result.Data.ProcessPercent >= 100 {
			res.subtitleInfo = result.Data.SubtitleInfo
			res.speechDownloadURL = result.Data.SpeechDownloadURL
			break
		}

		if result.Data.Status == "waiting_review" && !sm.reviewDialogShowing {
			sm.showReviewDialog(taskId, result.Data.ReviewSrtContent)
		}

		time.Sleep(2 * time.Second)
	}

	return res
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 2
}

// pollTaskStatus polls task status
func (sm *SubtitleManager) pollTaskStatus(taskId string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Record last progress percentage to avoid frequent updates
	lastPercent := 0

	for range ticker.C {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/api/capability/subtitleTask?taskId=%s", config.Conf.Server.Host, config.Conf.Server.Port, taskId))
		if err != nil {
			log.GetLogger().Error("Failed to get task status", zap.Error(err))
			dialog.ShowError(fmt.Errorf("Failed to get task status: %v", err), sm.window)
			return
		}

		var result struct {
			Error int    `json:"error"`
			Msg   string `json:"msg"`
			Data  struct {
				ProcessPercent    int                  `json:"process_percent"`
				Status            string               `json:"status"`
				SubtitleInfo      []api.SubtitleResult `json:"subtitle_info"`
				SpeechDownloadURL string               `json:"speech_download_url"`
				TaskId            string               `json:"task_id"`
				ReviewSrtContent  string               `json:"review_srt_content"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.GetLogger().Error("Failed to parse response", zap.Error(err))
			resp.Body.Close()
			dialog.ShowError(fmt.Errorf("Failed to get task status: %v", err), sm.window)
			return
		}
		resp.Body.Close()

		if result.Error != 0 {
			log.GetLogger().Error("Failed to get task status", zap.String("msg", result.Msg))
			dialog.ShowError(fmt.Errorf("Failed to get task status: %v", result.Msg), sm.window)
			return
		}

		if result.Data.ProcessPercent != lastPercent {
			progress := float64(result.Data.ProcessPercent) / 100.0
			sm.progressBar.SetValue(progress)

			if sm.progressLabel != nil {
				sm.progressLabel.SetText(fmt.Sprintf("%d%%", result.Data.ProcessPercent))
				sm.progressLabel.Show()
			}

			// Update last progress percentage
			lastPercent = result.Data.ProcessPercent
		}

		if result.Data.ProcessPercent >= 100 {
			// For single-file tasks, create and display a task result
			taskRes := taskResult{
				fileName:          filepath.Base(sm.videoUrl),
				subtitleInfo:      result.Data.SubtitleInfo,
				speechDownloadURL: result.Data.SpeechDownloadURL,
				taskId:            result.Data.TaskId,
			}

			sm.multiTaskResults = []taskResult{taskRes}

			sm.displayMultiTaskDownloadLinks()

			sm.tipsLabel.SetText(fmt.Sprintf("Check composite videos or transcripts in /tasks/%s/output in the app directory.", result.Data.TaskId))
			sm.tipsLabel.Show()

			return
		}

		if result.Data.Status == "waiting_review" && !sm.reviewDialogShowing {
			sm.showReviewDialog(taskId, result.Data.ReviewSrtContent)
		}
	}
}

// Display download links for multiple tasks
func (sm *SubtitleManager) displayMultiTaskDownloadLinks() {
	// Clear existing links
	sm.downloadContainer.Objects = []fyne.CanvasObject{}

	if len(sm.multiTaskResults) == 0 {
		return
	}

	allTasksContainer := container.NewVBox()

	for _, taskRes := range sm.multiTaskResults {
		taskLabel := widget.NewLabelWithStyle(
			fmt.Sprintf("File: %s", taskRes.fileName),
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		)

		taskContainer := container.NewVBox(taskLabel)

		for _, result := range taskRes.subtitleInfo {
			downloadURL := result.DownloadURL
			fileName := result.Name

			btn := widget.NewButton("Download "+fileName, func(url string) func() {
				return func() {
					go sm.downloadFile(url, filepath.Base(url))
				}
			}(downloadURL))
			btn.Importance = widget.MediumImportance
			taskContainer.Add(btn)
		}

		if taskRes.speechDownloadURL != "" {
			url := taskRes.speechDownloadURL
			ttsFileName := fmt.Sprintf("tts_%s.wav", filepath.Base(taskRes.speechDownloadURL))

			speechBtn := widget.NewButton("Download Dubbing File", func(u, f string) func() {
				return func() {
					go sm.downloadFile(u, f)
				}
			}(url, ttsFileName))
			speechBtn.Importance = widget.MediumImportance
			taskContainer.Add(speechBtn)
		}

		taskTip := widget.NewLabel(fmt.Sprintf("View video or transcript: /tasks/%s/output", taskRes.taskId))
		taskTip.Alignment = fyne.TextAlignCenter
		taskContainer.Add(taskTip)

		if &taskRes != &sm.multiTaskResults[len(sm.multiTaskResults)-1] {
			divider := canvas.NewLine(color.NRGBA{R: 200, G: 200, B: 200, A: 128})
			divider.StrokeWidth = 1
			taskContainer.Add(divider)
		}

		allTasksContainer.Add(taskContainer)
	}

	// Add all task containers to the download container
	sm.downloadContainer.Add(allTasksContainer)
	sm.downloadContainer.Show()
}

// General method to download files
func (sm *SubtitleManager) downloadFile(downloadURL, suggestedFileName string) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d", config.Conf.Server.Host, config.Conf.Server.Port) + downloadURL)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Download failed: %v", err), sm.window)
		return
	}

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, sm.window)
			return
		}
		if writer == nil {
			return // User cancelled
		}
		defer writer.Close()
		defer resp.Body.Close()

		_, err = io.Copy(writer, resp.Body)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save file: %v", err), sm.window)
			return
		}

		dialog.ShowInformation("Download Complete", "File saved successfully", sm.window)
	}, sm.window)

	saveDialog.SetFileName(suggestedFileName)
	saveDialog.Show()
}
