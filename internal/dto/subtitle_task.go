package dto

type StartVideoSubtitleTaskReq struct {
	AppId                     uint32   `json:"app_id"`
	Url                       string   `json:"url"`
	OriginLanguage            string   `json:"origin_lang"`
	TargetLang                string   `json:"target_lang"`
	Bilingual                 uint8    `json:"bilingual"`
	TranslationSubtitlePos    uint8    `json:"translation_subtitle_pos"`
	ModalFilter               uint8    `json:"modal_filter"`
	Tts                       uint8    `json:"tts"`
	TtsVoiceCode              string   `json:"tts_voice_code"`
	TtsVoiceCloneSrcFileUrl   string   `json:"tts_voice_clone_src_file_url"`
	Replace                   []string `json:"replace"`
	Language                  string   `json:"language"`
	EmbedSubtitleVideoType    string   `json:"embed_subtitle_video_type"`
	VerticalMajorTitle        string   `json:"vertical_major_title"`
	VerticalMinorTitle        string   `json:"vertical_minor_title"`
	OriginLanguageWordOneLine int      `json:"origin_language_word_one_line"`
	EnableReview              bool     `json:"enable_review"` // If true, pipeline pauses for subtitle review before TTS
}

type StartVideoSubtitleTaskResData struct {
	TaskId string `json:"task_id"`
}

type StartVideoSubtitleTaskRes struct {
	Error int32                          `json:"error"`
	Msg   string                         `json:"msg"`
	Data  *StartVideoSubtitleTaskResData `json:"data"`
}

type GetVideoSubtitleTaskReq struct {
	TaskId string `form:"taskId"`
}

type VideoInfo struct {
	Title                 string `json:"title"`
	Description           string `json:"description"`
	TranslatedTitle       string `json:"translated_title"`
	TranslatedDescription string `json:"translated_description"`
	Language              string `json:"language"`
}

type SubtitleInfo struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"download_url"`
}

type GetVideoSubtitleTaskResData struct {
	TaskId            string          `json:"task_id"`
	ProcessPercent    uint8           `json:"process_percent"`
	Status            string          `json:"status"` // "processing", "waiting_review", "success", "failed"
	VideoInfo         *VideoInfo      `json:"video_info"`
	VideoUrl          string          `json:"video_url"`
	SubtitleInfo      []*SubtitleInfo `json:"subtitle_info"`
	TargetLanguage    string          `json:"target_language"`
	TtsVoiceCode      string          `json:"tts_voice_code"`
	SpeechDownloadUrl string          `json:"speech_download_url"`
	ReviewSrtContent  string          `json:"review_srt_content,omitempty"` // SRT content for user to review/edit
}

type GetVideoSubtitleTaskRes struct {
	Error int32                        `json:"error"`
	Msg   string                       `json:"msg"`
	Data  *GetVideoSubtitleTaskResData `json:"data"`
}

type RenderSettings struct {
	OriginalVolume    int      `json:"original_volume"`
	SubtitleStyle     string   `json:"subtitle_style"`
	FontFamily        string   `json:"font_family"`
	FontSize          float64  `json:"font_size"`
	FontColor         string   `json:"font_color"`
	BorderColor       string   `json:"border_color"`
	BorderWidth       int      `json:"border_width"`
	BgPadding         int      `json:"bg_padding"`
	BottomDistance    int      `json:"bottom_distance"`
	LineSpacing       float64  `json:"line_spacing"`
	BgColor           string   `json:"bg_color"`
	IsBold            bool     `json:"is_bold"`
	DisplayMode       string   `json:"display_mode"`
	HighlightColor    string   `json:"highlight_color"`
	MaxWordsPerLine   int      `json:"max_words_per_line"`
	VideoRatio        string   `json:"video_ratio"`
	FitMode           string   `json:"fit_mode"`
}

type VoiceSettings struct {
	VoiceId string  `json:"voice_id"`
	Speed   float64 `json:"speed"`
	Emotion float64 `json:"emotion"`
}

// ApproveReviewReq is submitted by the user when they are done reviewing/editing subtitles
type ApproveReviewReq struct {
	TaskId           string          `json:"task_id"`
	EditedSrtContent string          `json:"edited_srt_content"` // The (potentially edited) SRT content
	RenderSettings   *RenderSettings `json:"render_settings,omitempty"`
	VoiceSettings    *VoiceSettings  `json:"voice_settings,omitempty"`
}
