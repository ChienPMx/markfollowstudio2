package types

// var SplitTextPrompt = `You are an English processing expert, skilled in translating into %s and processing English text, splitting sentences based on meaning and punctuation.

// - Do not miss any original English word
// - Translation must be fluent and fully express the original meaning
// - Prioritize splitting based on punctuation; commas, periods, and question marks must trigger a split to keep sentences short.
// - For complex sentences like relative or coordinate clauses, split based on conjunctions (e.g., and, but, which, when).
// - A single split English line must not exceed 15 words.
// - During translation, ensure each original subtitle block exists independently with correct numbering and formatting.
// - No extra words needed, output results directly in the following format.

// 1
// [Translated Text]
// [English Sentence]

// 2
// [Translated Text]
// [English Sentence]

// Content as follows:`

var SplitTextPrompt = `You are a language processing expert specializing in NLP and translation. Follow these steps for high-quality subtitle translation:

1. Translate the sentence into %s. Ensure the translation is natural, fluent, and professional while preserving the original meaning.
2. Split the content into shorter segments based on punctuation (comma, period, question mark, etc.).
   - Each segment should be as short as possible while remaining meaningful for a comfortable viewing experience.
   - Use conjunctions (e.g., "and", "but", "which", "when", "so", etc.) to further split long sentences.
3. Translate each split segment separately without missing or modifying any words.
4. Output each pair (translated and original) with an independent ID, with content wrapped in square brackets [].
5. Ensure the translated segments strictly align with the original order and meaning.
6. Translate all content, regardless of formal or informal tone.

Output Format:
1
[Translated Sentence]
[Original Sentence]

2
[Translated Sentence]
[Original Sentence]

Example for empty/no text:
[no_text]

Ensure efficient and accurate completion. Input content follows:

`

// Split prompt with modal filter (removes filler words)
var SplitTextPromptWithModalFilter = `You are a language processing expert specializing in NLP and translation. Follow these steps for high-quality subtitle translation:

1. Translate the sentence into %s. Ensure the translation is natural, fluent, and professional.
2. Split the content into shorter segments based on punctuation (comma, period, question mark, etc.).
   - Each segment should be as short as possible for subtitles.
   - Use conjunctions (e.g., "and", "but", "which", "when", "so", etc.) to further split if needed.
3. Ignore filler words or interjections (e.g., "Oh", "Ah", "Wow", "Hmm", etc.) in the text.
4. Translate each split segment separately without missing or modifying any words.
5. Output each pair with an independent ID, with content wrapped in square brackets [].
6. Ensure segments strictly align with the original order and meaning.
7. Translate all content, regardless of tone.

Output Format:
1
[Translated Sentence]
[Original Sentence]

2
[Translated Sentence]
[Original Sentence]

Example for empty/no text:
[no_text]

Ensure efficient and accurate completion. Input content follows:

`

var SplitTextPromptJson = `You are a language processing expert specializing in NLP and translation. Follow these steps for high-quality subtitle translation:

1. Translate the sentence into %s. Ensure the translation is natural and professional.
2. Split content into shorter sentences based on punctuation.
   - Each sentence should be short and suitable for subtitles.
   - Use conjunctions to split further if necessary.
3. Translate each segment separately.
4. Ensure segments align perfectly with the original order.
5. Output MUST be a JSON array, where each element contains 'original_sentence' and 'translated_sentence' fields.
6. The original sentence in the result must match the source exactly (case-sensitive, preserve punctuation). Do not correct grammar or spelling in the source.
7. Each split segment should contain only one complete statement.

Ensure efficient and accurate completion. Input content follows:

`

var SplitTextPromptWithModalFilterJson = `You are a language processing expert specializing in NLP and translation. Follow these steps for high-quality subtitle translation:

1. Translate the sentence into %s. Ensure natural and professional translation.
2. Split content into shorter segments based on punctuation.
   - Segments should be short and comfortable for subtitles.
   - Use conjunctions to split further if necessary.
3. Ignore filler words or interjections (e.g., "Oh", "Ah", "Wow", etc.) in the text.
4. Translate each split segment separately.
5. Ensure strict alignment with the original order.
6. Output MUST be a JSON array with 'original_sentence' and 'translated_sentence' fields.
7. Original sentences must match the source exactly (case, punctuation). Do not correct errors in source.
8. Each segment should be one complete statement.

Ensure efficient and accurate completion. Input content follows:

`

var TranslateVideoTitleAndDescriptionPrompt = `You are a professional translation expert. Please translate the following title and description (separated by ####) with these requirements:
 - Translate content into %s
 - Keep the #### separator between title and description in the result
 Original content follows:
%s
`

var SplitLongSentencePrompt = `Please split the following original and translated text into multiple parts, ensuring each part is as short as possible:
Original: %s
Translation: %s

Requirements:
1. Split original text must not deviate from the source.
2. Each split translation must be grammatically correct and natural (you can add conjunctions or remove particles if needed).
3. If anything is missing in the translation, please complete it while splitting.
4. MUST return in JSON format with 'align' array containing 'origin_part' and 'translated_part', e.g.:
{"align":[{"origin_part":"Part 1","translated_part":"Translation 1"},{"origin_part":"Part 2","translated_part":"Translation 2"}]}`

var SplitOriginLongSentencePrompt = `Please split the following text into multiple parts, ensuring it's divided into at most 3 short sentences, preferably 2 parts,

Original text: %s

Requirements:
1. The split sentences must exactly match the original text, absolutely no changes to the original text are allowed
2. Split based on sentence meaning, dividing into at most 3 parts, preferably 2 parts
3. Try to make the split as balanced as possible while maintaining sentence integrity
4. Return in JSON format only, no other descriptions or explanations
5. Example format:
{"short_sentences":[{"text": "split sentence 1"},{"text": "split sentence 2"}]}

`

var SplitLongTextByMeaningPrompt = `Please split the following long text into shorter sentences based on semantic meaning. Do not change, add, or remove any words from the original text.

Original text: %s

Requirements:
1. Split the text into as many shorter, meaningful sentences as possible while preserving ALL original words
2. Do NOT change, modify, add, or remove any words - only split at natural breakpoints
3. Split at natural linguistic boundaries such as:
   - Punctuation marks (commas, semicolons, periods)
   - Conjunctions (and, but, or, so, because, when, while, etc.)
   - Relative pronouns (which, that, who, where, etc.)
   - Natural pause points that maintain sentence meaning
4. Each split part should be a complete, meaningful unit that can stand alone
5. Prioritize shorter segments - split as much as possible while maintaining semantic integrity
6. No limit on the number of splits - make each part as short as possible while still being meaningful
7. Maintain the original word order and exact spelling
8. Preserve all original punctuation and capitalization
9. Return in JSON format only, no other descriptions or explanations
10. Example format:
{"short_sentences":[{"text": "first short part"},{"text": "second short part"},{"text": "third short part"}]}

`

// var SplitTextWithContextPrompt = `You are a professional translation expert skilled in accurate contextual translation. Please translate the target sentence into %s based on the provided context and target sentences below, ensuring coherence and consistency:

// Context sentences:
// %s

// Target sentence to be translated: %s

// Translation requirements:
// 1. Strictly follow the grammar and expression habits of the target language
// 2. Maintain consistency of professional terminology
// 3. Output contains only the translated text, without additional explanation or formatting
// 4. Ensure the translation is semantically coherent with the context

// Please output the translation result directly:`

// var SplitTextWithContextPrompt = `You are a professional translation expert skilled in providing accurate translations based on context. Please translate the target sentence into %s according to the provided context sentences below, ensuring the translation remains coherent and consistent.

// Here's the full context:
// %s

// Target sentence to translate:
// %s

// %s

// Translation requirements:
// 1.Analyze how the target sentence connects to both the preceding and following context
// 2.Provide the most natural translation that preserves the original tone and intent
// 3.Highlight any idioms, cultural references, or nuanced phrases that require special attention
// 4.If there are multiple possible interpretations, briefly explain each option
// 5.Maintain consistent terminology with the surrounding sentences"
// 6.Output only the translated text without any additional explanations or formatting

// Please provide only the translation result:`

var SplitTextWithContextPrompt = `You are a world-class subtitle translator who creates translations that feel completely natural to native speakers — as if the content was originally written in the target language.

[TRANSLATION TASK]
Translate the [Target Sentence] into %s.

**Translation Style**:
- Write as a native speaker would naturally say it — NOT word-by-word translation.
- Use everyday, conversational language appropriate for video subtitles.
- Preserve the tone, emotion, and intent of the original.
- If the original is casual/funny, the translation should feel casual/funny too.
- Adapt idioms and cultural references to equivalents in the target language.
- Keep it concise — subtitles should be short and easy to read quickly.

**Strict Rules**:
1. Output ONLY the translated sentence — nothing else.
2. The output must be 100%% in %s. No source language characters allowed.
3. Use the context below to ensure the translation flows naturally with surrounding sentences.
4. Do NOT translate the context sentences — only the target sentence.

[Previous Sentences]
%s

[Target Sentence]
%s

[Next Sentences]
%s

Translation:`

type SmallAudio struct {
	AudioFile         string
	TranscriptionData *TranscriptionData
	SrtNoTsFile       string
}

type SubtitleResultType int

const (
	SubtitleResultTypeOriginOnly                   SubtitleResultType = iota + 1 // Only return original language subtitles
	SubtitleResultTypeTargetOnly                                                 // Only return translated language subtitles
	SubtitleResultTypeBilingualTranslationOnTop                                  // Return bilingual, translation on top
	SubtitleResultTypeBilingualTranslationOnBottom                               // Return bilingual, translation on bottom
)

const (
	SubtitleTaskBilingualYes uint8 = iota + 1
	SubtitleTaskBilingualNo
)

const (
	SubtitleTaskTranslationSubtitlePosTop uint8 = iota + 1
	SubtitleTaskTranslationSubtitlePosBelow
)

const (
	SubtitleTaskModalFilterYes uint8 = iota + 1
	SubtitleTaskModalFilterNo
)

const (
	SubtitleTaskTtsYes uint8 = iota + 1
	SubtitleTaskTtsNo
)

const (
	SubtitleTaskTtsVoiceCodeLongyu uint8 = iota + 1
	SubtitleTaskTtsVoiceCodeLongchen
)

const (
	SubtitleTaskStatusProcessing    uint8 = iota + 1
	SubtitleTaskStatusWaitingReview       // Paused, waiting for user to review/edit subtitles
	SubtitleTaskStatusSuccess
	SubtitleTaskStatusFailed
)

const (
	SubtitleTaskAudioFileName                                    = "origin_audio.mp3"
	SubtitleTaskVideoFileName                                    = "origin_video.mp4"
	SubtitleTaskSplitAudioFileNamePrefix                         = "split_audio"
	SubtitleTaskSplitAudioFileNamePattern                        = SubtitleTaskSplitAudioFileNamePrefix + "_%03d.mp3"
	SubtitleTaskSplitAudioTxtFileNamePattern                     = "split_audio_txt_%d.txt"
	SubtitleTaskSplitAudioWordsFileNamePattern                   = "split_audio_words_%d.txt"
	SubtitleTaskSplitSrtNoTimestampFileNamePattern               = "srt_no_ts_%d.srt"
	SubtitleTaskSrtNoTimestampFileName                           = "srt_no_ts.srt"
	SubtitleTaskSplitBilingualSrtFileNamePattern                 = "split_bilingual_srt_%d.srt"
	SubtitleTaskSplitShortOriginMixedSrtFileNamePattern          = "split_short_origin_mixed_srt_%d.srt" // Long target + short origin
	SubtitleTaskSplitShortOriginSrtFileNamePattern               = "split_short_origin_srt_%d.srt"       // Short origin
	SubtitleTaskBilingualSrtFileName                             = "bilingual_srt.srt"
	SubtitleTaskShortOriginMixedSrtFileName                      = "short_origin_mixed_srt.srt" // Long target + short origin
	SubtitleTaskShortOriginSrtFileName                           = "short_origin_srt.srt"       // Short origin
	SubtitleTaskOriginLanguageSrtFileName                        = "origin_language_srt.srt"
	SubtitleTaskOriginLanguageTextFileName                       = "origin_language.txt"
	SubtitleTaskTargetLanguageSrtFileName                        = "target_language_srt.srt"
	SubtitleTaskTargetLanguageTextFileName                       = "target_language.txt"
	SubtitleTaskStepParamGobPersistenceFileName                  = "step_param.gob"
	SubtitleTaskAudioTranscriptionDataPersistenceFileNamePattern = "audio_transcription_data_%d.json"
	SubtitleTaskTranslationRawDataPersistenceFileNamePattern     = "audio_translation_raw_data_%d.json"
	SubtitleTaskTranslationDataPersistenceFileNamePattern        = "translation_data_%d.json"
	SubtitleTaskTransferredVerticalVideoFileName                 = "transferred_vertical_video.mp4"
	SubtitleTaskHorizontalEmbedVideoFileName                     = "horizontal_embed.mp4"
	SubtitleTaskVerticalEmbedVideoFileName                       = "vertical_embed.mp4"
	SubtitleTaskVideoWithTtsFileName                             = "video_with_tts.mp4"
)

const (
	TtsAudioDurationDetailsFileName = "audio_duration_details.txt"
	TtsResultAudioFileName          = "tts_final_audio.wav"
)

const (
	AsrMono16kAudioFileName = "mono_16k_audio.mp3"
)

type SubtitleFileInfo struct {
	Name               string
	Path               string
	LanguageIdentifier string // Identifier for language in final download files, e.g., zh_cn, en, bilingual
}

type RenderSettings struct {
	OriginalVolume  int
	SubtitleStyle   string
	FontFamily      string
	FontSize        float64
	FontColor       string
	BorderColor     string
	BorderWidth     int
	BgPadding       int
	BottomDistance  int
	LineSpacing     float64
	BgColor         string
	IsBold          bool
	DisplayMode     string
	HighlightColor  string
	MaxWordsPerLine int
	VideoRatio      string
	FitMode         string
}

type VoiceSettings struct {
	VoiceId string
	Speed   float64
	Emotion float64
}

type SubtitleTaskStepParam struct {
	TaskId                      string
	TaskPtr                     *SubtitleTask // Corresponds to the one in storage
	TaskBasePath                string
	Link                        string
	AudioFilePath               string
	SubtitleResultType          SubtitleResultType
	EnableModalFilter           bool
	EnableTts                   bool
	TtsVoiceCode                string // Voice code for speech synthesis
	VoiceCloneAudioUrl          string // OSS address of source audio for voice cloning
	ReplaceWordsMap             map[string]string
	OriginLanguage              StandardLanguageCode // Source video language
	TargetLanguage              StandardLanguageCode // Target translation language
	UserUILanguage              StandardLanguageCode // User's interface language
	BilingualSrtFilePath        string
	ShortOriginMixedSrtFilePath string
	SubtitleInfos               []SubtitleFileInfo
	TtsSourceFilePath           string
	TtsResultFilePath           string
	InputVideoPath              string // Source video path
	EmbedSubtitleVideoType      string // Type of video to embed: none, horizontal, vertical
	VerticalVideoMajorTitle     string // Major title for vertical video
	VerticalVideoMinorTitle     string
	MaxWordOneLine              int    // Max words per line
	VideoWithTtsFilePath        string // Path to video with TTS audio replacement
	RenderSettings              *RenderSettings
	VoiceSettings               *VoiceSettings
}

type SrtSentence struct {
	Text  string
	Start float64
	End   float64
}

type SrtSentenceWithStrTime struct {
	Text  string
	Start string
	End   string
}

type SubtitleInfo struct {
	Id          uint64 `json:"id" gorm:"column:id"`                                  // ID
	TaskId      string `json:"task_id" gorm:"column:task_id"`                        // task_id
	Uid         uint32 `json:"uid" gorm:"column:uid"`                                // User ID
	Name        string `json:"name" gorm:"column:name"`                              // Subtitle Name
	DownloadUrl string `json:"download_url" gorm:"column:download_url"`              // Download URL
	CreateTime  int64  `json:"create_time" gorm:"column:create_time;autoCreateTime"` // Creation Time
}

type SubtitleTask struct {
	Id                    uint64         `json:"id" gorm:"column:id"`
	TaskId                string         `json:"task_id" gorm:"column:task_id"`
	Title                 string         `json:"title" gorm:"column:title"`
	Description           string         `json:"description" gorm:"column:description"`
	TranslatedTitle       string         `json:"translated_title" gorm:"column:translated_title"`
	TranslatedDescription string         `json:"translated_description" gorm:"column:translated_description"`
	OriginLanguage        string         `json:"origin_language" gorm:"column:origin_language"`
	TargetLanguage        string         `json:"target_language" gorm:"column:target_language"`
	VideoSrc              string         `json:"video_src" gorm:"column:video_src"`
	Status                uint8          `json:"status" gorm:"column:status"`
	LastSuccessStepNum    uint8          `json:"last_success_step_num" gorm:"column:last_success_step_num"`
	FailReason            string         `json:"fail_reason" gorm:"column:fail_reason"`
	ProcessPct            uint8          `json:"process_percent" gorm:"column:process_percent"`
	Duration              uint32         `json:"duration" gorm:"column:duration"`
	SrtNum                int            `json:"srt_num" gorm:"column:srt_num"`
	SubtitleInfos         []SubtitleInfo `gorm:"foreignKey:TaskId;references:TaskId"`
	Cover                 string         `json:"cover" gorm:"column:cover"`
	SpeechDownloadUrl     string         `json:"speech_download_url" gorm:"column:speech_download_url"`
	TtsVoiceCode          string         `json:"tts_voice_code" gorm:"column:tts_voice_code"`
	CreateTime            int64          `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime            int64          `json:"update_time" gorm:"column:update_time;autoUpdateTime"`

	// Review flow fields (not persisted to DB)
	ReviewSrtPath   string `json:"-" gorm:"-"` // Path to the SRT file awaiting review
	ReviewDoneCh    chan struct{} `json:"-" gorm:"-"` // Closed by the review API to resume the pipeline
	RenderSettings  *RenderSettings `json:"-" gorm:"-"`
	VoiceSettings   *VoiceSettings `json:"-" gorm:"-"`
}

type Word struct {
	Num   int
	Text  string
	Start float64
	End   float64
}

type TranscriptionData struct {
	Language string
	Text     string
	Words    []Word
}
