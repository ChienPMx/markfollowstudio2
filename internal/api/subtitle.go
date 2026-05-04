package api

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WordReplacement represents a word replacement pair
type WordReplacement struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// SubtitleTask represents a subtitle generation task
type SubtitleTask struct {
	URL                     string   `json:"url"`                                    // Video URL
	Language                string   `json:"language"`                               // Interface Language
	OriginLang              string   `json:"origin_lang"`                            // Source Language
	TargetLang              string   `json:"target_lang"`                            // Target Language
	Bilingual               int      `json:"bilingual"`                              // Is Bilingual 1:Yes 2:No
	TranslationSubtitlePos  int      `json:"translation_subtitle_pos"`               // Translation Subtitle Position 1:Top 2:Bottom
	TTS                     int      `json:"tts"`                                    // Is Dubbing 1:Yes 2:No
	TTSVoiceCode            string   `json:"tts_voice_code,omitempty"`               // Dubbing Voice Code
	TTSVoiceCloneSrcFileURL string   `json:"tts_voice_clone_src_file_url,omitempty"` // Voice Clone Source File URL
	ModalFilter             int      `json:"modal_filter"`                           // Filter Filler Words 1:Yes 2:No
	Replace                 []string `json:"replace,omitempty"`                      // Word Replacement List
	EmbedSubtitleVideoType  string   `json:"embed_subtitle_video_type"`              // Subtitle Embedding Type: none, horizontal, vertical, all
	VerticalMajorTitle      string   `json:"vertical_major_title,omitempty"`         // Vertical Video Major Title
	VerticalMinorTitle      string   `json:"vertical_minor_title,omitempty"`         // Vertical Video Minor Title
	EnableReview            bool     `json:"enable_review"`                          // Pause pipeline for subtitle review before TTS
}

// SubtitleResult represents the result of a subtitle generation
type SubtitleResult struct {
	Name        string `json:"name"`         // Filename
	DownloadURL string `json:"download_url"` // Download URL
}

// TaskStatus represents the current status of a task
type TaskStatus struct {
	TaskId            string           `json:"task_id"`             // Task ID
	ProcessPercent    int              `json:"process_percent"`     // Progress Percentage
	Status            string           `json:"status"`              // Task Status
	Message           string           `json:"message"`             // Status Message
	SubtitleInfo      []SubtitleResult `json:"subtitle_info"`       // Subtitle Information
	SpeechDownloadURL string           `json:"speech_download_url"` // Dubbing Download URL
	ReviewSrtContent  string           `json:"review_srt_content"`  // SRT content for user to review/edit
}

// CreateSubtitleTask creates a new subtitle task
func CreateSubtitleTask(task *SubtitleTask) (*TaskStatus, error) {
	// Generate Task ID
	taskId := generateTaskId()

	// Create Task Directory
	taskDir := filepath.Join("tasks", taskId)
	if err := createTaskDirectory(taskDir); err != nil {
		return nil, fmt.Errorf("Failed to create task directory: %v", err)
	}

	// Start asynchronous task processing
	go processTask(taskId, task)

	return &TaskStatus{
		TaskId:         taskId,
		ProcessPercent: 0,
		Status:         "created",
		Message:        "Task created",
	}, nil
}

// GetSubtitleTaskStatus retrieves the status of a subtitle task
func GetSubtitleTaskStatus(taskId string) (*TaskStatus, error) {
	// Get task status
	status, err := getTaskStatus(taskId)
	if err != nil {
		return nil, fmt.Errorf("Failed to get task status: %v", err)
	}

	// If task complete, add download links
	if status.ProcessPercent >= 100 {
		status.SubtitleInfo = []SubtitleResult{
			{
				Name:        "subtitle.srt",
				DownloadURL: fmt.Sprintf("/tasks/%s/output/subtitle.srt", taskId),
			},
			{
				Name:        "subtitle.ass",
				DownloadURL: fmt.Sprintf("/tasks/%s/output/subtitle.ass", taskId),
			},
		}

		// If dubbing enabled, add dubbing download link
		if status.SpeechDownloadURL == "" {
			status.SpeechDownloadURL = fmt.Sprintf("/tasks/%s/output/speech.mp3", taskId)
		}
	}

	return status, nil
}

// Helper functions to be implemented
func generateTaskId() string {
	// TODO: Implement task ID generation logic
	return "task-" + time.Now().Format("20060102150405")
}

func createTaskDirectory(taskDir string) error {
	// TODO: Implement task directory creation logic
	return os.MkdirAll(taskDir, 0755)
}

func processTask(taskId string, task *SubtitleTask) {
	// TODO: Implement task processing logic
	// 1. Download video
	// 2. Extract audio
	// 3. Speech recognition
	// 4. Translate subtitles
	// 5. Generate subtitle files
	// 6. Generate dubbing if needed
	// 7. Embed subtitles into video if needed
	// 8. Update task status
}

func getTaskStatus(taskId string) (*TaskStatus, error) {
	// TODO: Implement task status retrieval logic
	return &TaskStatus{
		TaskId:         taskId,
		ProcessPercent: 50,
		Status:         "processing",
		Message:        "Processing",
	}, nil
}
