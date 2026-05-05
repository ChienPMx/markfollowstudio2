package util

import (
	"bufio"
	"fmt"
	"markflow-studio/internal/storage"
	"markflow-studio/internal/types"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Process each subtitle block
func ProcessBlock(block []string, targetLanguageFile, targetLanguageTextFile, originLanguageFile, originLanguageTextFile *os.File, isTargetOnTop bool) {
	var targetLines, originLines []string
	// Regex pattern for matching timestamps
	timePattern := regexp.MustCompile(`\d{2}:\d{2}:\d{2},\d{3} --> \d{2}:\d{2}:\d{2},\d{3}`)
	for _, line := range block {
		if timePattern.MatchString(line) || IsNumber(line) {
			// Retain timestamp and index lines in both files
			targetLines = append(targetLines, line)
			originLines = append(originLines, line)
			continue
		}
		if len(targetLines) == 2 && len(originLines) == 2 { // Just finished index and timestamp, reached the upper text line
			if isTargetOnTop {
				targetLines = append(targetLines, line)
				targetLanguageTextFile.WriteString(line + " ") // Transcript file
			} else {
				originLines = append(originLines, line)
				originLanguageTextFile.WriteString(line + " ")
			}
			continue
		}
		// Reached the lower text line
		if isTargetOnTop {
			originLines = append(originLines, line)
			originLanguageTextFile.WriteString(line + " ")
		} else {
			targetLines = append(targetLines, line)
			targetLanguageTextFile.WriteString(line + " ")
		}
	}

	if len(targetLines) > 2 {
		// Write to target language file
		for _, line := range targetLines {
			targetLanguageFile.WriteString(line + "\n")
		}
		targetLanguageFile.WriteString("\n")
	}

	if len(originLines) > 2 {
		// Write to source language file
		for _, line := range originLines {
			originLanguageFile.WriteString(line + "\n")
		}
		originLanguageFile.WriteString("\n")
	}
}

// IsSubtitleText checks if a line is a subtitle text line
func IsSubtitleText(line string) bool {
	if line == "" {
		return false
	}
	if IsNumber(line) {
		return false
	}
	timelinePattern := regexp.MustCompile(`\d{2}:\d{2}:\d{2},\d{3} --> \d{2}:\d{2}:\d{2},\d{3}`)
	return !timelinePattern.MatchString(line)
}

type Format struct {
	Duration string `json:"duration"`
}

type ProbeData struct {
	Format Format `json:"format"`
}

type SrtBlock struct {
	Index                  int
	Timestamp              string
	TargetLanguageSentence string
	OriginLanguageSentence string
}

func TrimString(s string) string {
	s = strings.Replace(s, "[Translated Sentence]", "", -1)
	s = strings.Replace(s, "[Original Sentence]", "", -1)
	s = strings.Replace(s, "[Chinese Translation]", "", -1)
	s = strings.Replace(s, "[English Sentence]", "", -1)
	// Remove leading spaces and '['
	s = strings.TrimLeft(s, " [")

	// Remove trailing spaces and ']'
	s = strings.TrimRight(s, " ]")

	// Replace CJK single quotes
	s = strings.ReplaceAll(s, "’", "'")

	return s
}

func SplitSentence(sentence string) []string {
	// Use regex to remove punctuation and special characters (preserving alphanumeric and spaces across languages)
	re := regexp.MustCompile(`[^\p{L}\p{N}\s']+`)
	cleanedSentence := re.ReplaceAllString(sentence, " ")

	// Use strings.Fields to split into words by whitespace
	words := strings.Fields(cleanedSentence)

	return words
}

func MergeFile(finalFile string, files ...string) error {
	// Create final file
	final, err := os.Create(finalFile)
	if err != nil {
		return err
	}

	// Read files one by one and write to the final file
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			final.WriteString(line + "\n")
		}
	}

	return nil
}

func MergeSrtFiles(finalFile string, files ...string) error {
	output, err := os.Create(finalFile)
	if err != nil {
		return err
	}
	defer output.Close()
	writer := bufio.NewWriter(output)
	lineNumber := 0
	for _, file := range files {
		// Skip if a file does not exist
		if _, err = os.Stat(file); os.IsNotExist(err) {
			continue
		}
		// Open current subtitle file
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		// Process current subtitle file
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "```") {
				continue
			}

			if IsNumber(line) {
				lineNumber++
				line = strconv.Itoa(lineNumber)
			}

			writer.WriteString(line + "\n")
		}
	}
	writer.Flush()

	return nil
}

// Replace all keys with values in the given file using the replacement map
func ReplaceFileContent(srcFile, dstFile string, replacements map[string]string) error {
	file, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer file.Close()

	outFile, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(outFile) // For performance improvement
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		for before, after := range replacements {
			line = strings.ReplaceAll(line, before, after)
		}
		_, _ = writer.WriteString(line + "\n")
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Generate a new filename by adding a suffix before the extension, e.g., /path/abc.srt becomes /path/abc_suffix.srt
func AddSuffixToFileName(filePath, suffix string) string {
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), ext)
	newName := fmt.Sprintf("%s%s%s", name, suffix, ext)
	return filepath.Join(dir, newName)
}

// Remove punctuation and other symbols to ensure text is recognizable by Whisper models, facilitating timestamp alignment
func GetRecognizableString(s string) string {
	var result []rune
	for _, v := range s {
		// English letters and numbers
		if unicode.Is(unicode.Latin, v) || unicode.Is(unicode.Number, v) {
			result = append(result, v)
		}
		// Chinese
		if unicode.Is(unicode.Han, v) {
			result = append(result, v)
		}
		// Korean
		if unicode.Is(unicode.Hangul, v) {
			result = append(result, v)
		}
		// Japanese Hiragana/Katakana
		if unicode.Is(unicode.Hiragana, v) || unicode.Is(unicode.Katakana, v) {
			result = append(result, v)
		}
	}
	return string(result)
}

func GetAudioDuration(inputFile string) (float64, error) {
	// Use ffprobe to get precise duration
	cmd := exec.Command(storage.FfprobePath, "-i", inputFile, "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0")
	cmdOutput, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("GetAudioDuration failed to get audio duration: %w", err)
	}

	// Parse duration
	duration, err := strconv.ParseFloat(strings.TrimSpace(string(cmdOutput)), 64)
	if err != nil {
		return 0, fmt.Errorf("GetAudioDuration failed to parse audio duration: %w", err)
	}

	return duration, nil
}

// todo: Add more later
func IsAsianLanguage(code types.StandardLanguageCode) bool {
	return code == types.LanguageNameSimplifiedChinese || code == types.LanguageNameTraditionalChinese || code == types.LanguageNameJapanese || code == types.LanguageNameKorean || code == types.LanguageNameThai
}

func BeautifyAsianLanguageSentence(input string) string {
	if len(input) == 0 {
		return input
	}

	// Punctuations not to process (balanced pairs)
	pairPunctuations := map[rune]rune{
		'「': '」', '『': '』', '“': '”', '‘': '’',
		'《': '》', '<': '>', '【': '】', '〔': '〕',
		'(': ')', '[': ']', '{': '}',
	}

	// Single punctuations to be processed
	singlePunctuations := ",.;:!?~，、。！？；：…"

	// Process punctuations at the end of the string first
	runes := []rune(input)
	i := len(runes) - 1
	for i >= 0 {
		r := runes[i]
		// If it's a space, check the previous character
		if unicode.IsSpace(r) {
			i--
			continue
		}
		// If it's a single punctuation, remove it
		if strings.ContainsRune(singlePunctuations, r) {
			runes = runes[:i]
			i--
		} else {
			// Stop when encountering non-punctuation or paired punctuation
			break
		}
	}

	// Replace single punctuations in the middle with spaces
	var inPair bool
	var expectedClose rune
	var result []rune

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Check if inside paired punctuations
		if inPair {
			if r == expectedClose {
				inPair = false
			}
			result = append(result, r)
			continue
		}

		// Check if it's the start of a paired punctuation
		if close, isPair := pairPunctuations[r]; isPair {
			inPair = true
			expectedClose = close
			result = append(result, r)
			continue
		}

		// Check if it's a decimal point in a number
		if r == '.' && i > 0 && i < len(runes)-1 {
			prev := runes[i-1]
			next := runes[i+1]
			if unicode.IsDigit(prev) && unicode.IsDigit(next) {
				result = append(result, r)
				continue
			}
		}

		// Handle single punctuation
		if strings.ContainsRune(singlePunctuations, r) {
			// Replace with space, avoiding consecutive spaces
			if len(result) > 0 && !unicode.IsSpace(result[len(result)-1]) {
				result = append(result, ' ')
			}
		} else {
			result = append(result, r)
		}
	}

	return strings.TrimSpace(string(result))
}

// SplitTextSentences splits text into sentences by common full/half-width delimiters, considering special cases that shouldn't be split
// maxChars: Minimum characters; full sentences smaller than this are not split, otherwise even commas trigger splits
// Usage example:
//
//	SplitTextSentences("Hello, World!", 5)  // Returns: ["Hello, World!"] (Not split because total characters < 5)
//	SplitTextSentences("This is a very long sentence, containing a lot of content.", 10) // Returns: ["This is a very long sentence", "containing a lot of content."] (Split by comma)
func SplitTextSentences(text string, maxChars int) []string {
	if strings.TrimSpace(text) == "" {
		return []string{}
	}

	// Step 1: Protect special patterns (numbers, times, abbreviations, etc.)
	text = protectSpecialNumbers(text)

	// Step 2: Intelligent splitting - First split by complete sentences
	completeSentences := splitByCompleteSentences(text)

	var result []string
	for _, sentence := range completeSentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Count effective characters (excluding punctuation and spaces)
		effectiveChars := CountEffectiveChars(sentence)

		// If the complete sentence is smaller than the minimum character count, don't split
		if effectiveChars < maxChars {
			cleaned := restoreProtectedPatterns(sentence)
			result = append(result, strings.TrimSpace(cleaned))
		} else {
			// Full sentence is too long, need to split further by commas or other punctuation
			subSentences := splitByAllPunctuation(sentence)
			merged := mergeShortSentences(subSentences, 20, maxChars)

			for _, subSentence := range merged {
				cleaned := restoreProtectedPatterns(subSentence)
				cleaned = strings.TrimSpace(cleaned)
				if cleaned != "" {
					result = append(result, cleaned)
				}
			}
		}
	}

	return result
}

// protectedPatterns stores protected patterns
var protectedPatterns map[string]string

// protectSpecialNumbers protects numbers, times, abbreviations, etc., from being mis-cut
func protectSpecialNumbers(text string) string {
	protectedPatterns = make(map[string]string)

	// Use a more direct method to protect list numbering patterns
	// Process specific patterns first, like "1.value", "2.be", "3.give", etc.
	listNumberPattern := regexp.MustCompile(`\b\d+\.[a-zA-Z]`)
	text = listNumberPattern.ReplaceAllStringFunc(text, func(match string) string {
		placeholder := fmt.Sprintf("\uE000%d\uE000", len(protectedPatterns))
		protectedPatterns[placeholder] = match
		return placeholder
	})

	patterns := []struct {
		regex *regexp.Regexp
		name  string
	}{
		// Protect domains and URLs (e.g., .com, .org, .net, etc.)
		{regexp.MustCompile(`\b[a-zA-Z0-9-]+\.(?:com|org|net|edu|gov|mil|int|co|io|ai|me|tv|fm|am|pm|uk|cn|jp|de|fr|it|es|ru|in|au|ca|br|mx|ar|cl|pe|ve|ec|py|uy|bo|gf|sr|gy|fk|gs|sh|ac|ad|ae|af|ag|al|am|an|ao|aq|as|at|aw|ax|az|ba|bb|bd|be|bf|bg|bh|bi|bj|bm|bn|bo|br|bs|bt|bv|bw|by|bz|cc|cd|cf|cg|ch|ci|ck|cm|co|cr|cs|cu|cv|cx|cy|cz|dj|dk|dm|do|dz|eg|eh|er|et|eu|fi|fj|fk|fo|ga|gb|gd|ge|gf|gg|gh|gi|gl|gm|gn|gp|gq|gr|gs|gt|gu|gw|gy|hk|hm|hn|hr|ht|hu|id|ie|il|im|iq|ir|is|je|jm|jo|ke|kg|kh|ki|km|kn|kp|kr|kw|ky|kz|la|lb|lc|li|lk|lr|ls|lt|lu|lv|ly|ma|mc|md|me|mg|mh|mk|ml|mm|mn|mo|mp|mq|mr|ms|mt|mu|mv|mw|my|mz|na|nc|ne|nf|ng|ni|nl|no|np|nr|nu|nz|om|pa|pg|ph|pk|pl|pm|pn|pr|ps|pt|pw|qa|re|ro|rs|rw|sa|sb|sc|sd|se|sg|si|sj|sk|sl|sm|sn|so|st|su|sv|sy|sz|tc|td|tf|tg|th|tj|tk|tl|tm|tn|to|tp|tr|tt|tz|ua|ug|um|us|uy|uz|va|vc|vg|vi|vn|vu|wf|ws|ye|yt|za|zm|zw)\b`), "domain"},
		// Protect abbreviations like a.m., p.m., A.M., P.M.
		{regexp.MustCompile(`(?i)\b[ap]\.m\.`), "ampm"},
		// Time format
		{regexp.MustCompile(`\b\d{1,2}[:\.]\d{2}\s*(?:[ap]\.?m\.?|AM|PM)?\b`), "time"},
		// Decimals (including multi-digit decimals)
		{regexp.MustCompile(`\b\d+\.\d+\b`), "decimal"},
		// Thousands separator
		{regexp.MustCompile(`\b\d{1,3}(?:,\d{3})+(?:\.\d+)?\b`), "thousands"},
		// Version numbers (e.g., 1.0, 2.5.1, etc.)
		{regexp.MustCompile(`\b\d+(?:\.\d+)+\b`), "version"},
		// English abbreviations
		{regexp.MustCompile(`\b(?:[A-Z][a-z]*\.){2,}|(?:[A-Z]\.){2,}[A-Z]?\b`), "abbrev"},
		// Titles like Mr., Mrs., Dr., etc.
		{regexp.MustCompile(`\b(?:Mr|Mrs|Ms|Dr|Prof|Sr|Jr)\.`), "title"},
		// List numbering (e.g., 1., 2., 3., etc.) - Number + Dot + Space
		{regexp.MustCompile(`\b\d+\.\s`), "list_number_with_space"},
		// Letter numbering (e.g., a., b., c., etc.)
		{regexp.MustCompile(`\b[a-zA-Z]\.\s`), "letter_number_with_space"},
	}

	for _, pattern := range patterns {
		text = pattern.regex.ReplaceAllStringFunc(text, func(match string) string {
			placeholder := fmt.Sprintf("\uE000%d\uE000", len(protectedPatterns))
			protectedPatterns[placeholder] = match
			return placeholder
		})
	}

	return text
}

// splitByCompleteSentences splits by complete sentence punctuation (period, exclamation, question mark, etc.)
func splitByCompleteSentences(text string) []string {
	// Split only by end-of-sentence punctuation, excluding commas
	completeSentenceMarkers := []string{
		".", "!", "?", "。", "！", "？", "；", "\n", "\r\n",
	}

	// Create regex pattern
	var patterns []string
	for _, marker := range completeSentenceMarkers {
		patterns = append(patterns, regexp.QuoteMeta(marker))
	}

	// Match consecutive end-of-sentence punctuation marks
	regexPattern := fmt.Sprintf(`([%s]+)`, strings.Join(patterns, ""))
	regex := regexp.MustCompile(regexPattern)

	// Add separator after punctuation marks
	text = regex.ReplaceAllString(text, "${1}\uE001")

	// Split by separator
	parts := strings.Split(text, "\uE001")

	var segments []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			segments = append(segments, trimmed)
		}
	}

	return segments
}

// countEffectiveChars counts effective characters (excluding punctuation and spaces)
func CountEffectiveChars(text string) int {
	effectiveText := regexp.MustCompile(`[^\p{L}\p{N}]`).ReplaceAllString(text, "")
	return len([]rune(effectiveText))
}

// splitByAllPunctuation splits text by all punctuation marks
func splitByAllPunctuation(text string) []string {
	// Note: text here has already been protected in SplitTextSentences, no need to protect again

	// Define splitting punctuation marks (including CJK and Latin)
	punctuationMarkers := []string{
		// End-of-sentence punctuation
		".", "!", "?", "；", "。", "！", "？", "；",
		// Intra-sentence punctuation (to be split)
		",", "，", ";",
		// Newline characters
		"\n", "\r\n",
	}

	// Create regex pattern
	var patterns []string
	for _, marker := range punctuationMarkers {
		patterns = append(patterns, regexp.QuoteMeta(marker))
	}

	// Match consecutive punctuation marks
	regexPattern := fmt.Sprintf(`([%s]+)`, strings.Join(patterns, ""))
	regex := regexp.MustCompile(regexPattern)

	// Add separator after punctuation marks
	text = regex.ReplaceAllString(text, "${1}\uE001")

	// Split by separator
	parts := strings.Split(text, "\uE001")

	var segments []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			segments = append(segments, trimmed)
		}
	}

	return segments
}

// mergeShortSentences merges sentences that are too short
// minChars: Minimum characters; consider merging if below this value
// maxChars: Maximum characters; merged sentence cannot exceed this value
func mergeShortSentences(segments []string, minChars, maxChars int) []string {
	if len(segments) == 0 {
		return segments
	}

	var result []string
	var current strings.Builder

	for i, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		// Add to current sentence
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(segment)

		currentText := current.String()
		currentEffectiveChars := CountEffectiveChars(currentText)

		// Check if the next segment should be merged
		shouldMerge := false
		if i < len(segments)-1 { // Still more segments
			nextSegment := strings.TrimSpace(segments[i+1])
			if nextSegment != "" {
				// Calculate merged length
				potentialMerged := currentText + " " + nextSegment
				mergedEffectiveChars := CountEffectiveChars(potentialMerged)

				// Merge only if current sentence is below minChars and merged result is within maxChars
				shouldMerge = currentEffectiveChars < minChars && mergedEffectiveChars <= maxChars
			}
		}

		if !shouldMerge {
			// No merge; output current sentence and reset
			result = append(result, strings.TrimSpace(currentText))
			current.Reset()
		}
		// If shouldMerge is true, continue to the next segment for merging
	}

	// Handle the final segment
	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// isTooShort determines if a sentence is too short and needs merging
func isTooShort(text string, maxChars int) bool {
	text = strings.TrimSpace(text)

	// Count effective characters (excluding punctuation and spaces)
	effectiveChars := CountEffectiveChars(text)

	// If effective characters are fewer than minChars, it's considered too short
	if effectiveChars < maxChars {
		return true
	}

	// If there's only one word, it's also considered too short (unless minChars is reached)
	words := strings.Fields(text)
	return len(words) <= 1 && effectiveChars < maxChars
}

// restoreProtectedPatterns restores protected patterns
func restoreProtectedPatterns(text string) string {
	for placeholder, original := range protectedPatterns {
		text = strings.ReplaceAll(text, placeholder, original)
	}
	return text
}

// Convert start and end to specified format
func ConvertTimes(start, end float32) string {
	startTime := FormatTime(start)
	endTime := FormatTime(end)
	return fmt.Sprintf("%s --> %s", startTime, endTime)
}
