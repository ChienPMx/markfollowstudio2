package desktop

import (
	"fmt"
	"image/color"
	"krillin-ai/config"
	"krillin-ai/internal/deps"
	"krillin-ai/internal/server"
	"krillin-ai/internal/types"
	"krillin-ai/log"
	"krillin-ai/static"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.uber.org/zap"
)

// Create config tab
func CreateConfigTab(window fyne.Window) fyne.CanvasObject {
	pageTitle := TitleText("App Configuration")

	appGroup := createAppConfigGroup()
	serverGroup := createServerConfigGroup()
	transcribeGroup := createTranscribeConfigGroup()
	ttsGroup := createTtsConfigGroup()

	var background *canvas.LinearGradient
	if GetCurrentThemeIsDark() {
		background = canvas.NewLinearGradient(
			color.NRGBA{R: 15, G: 23, B: 42, A: 255},
			color.NRGBA{R: 30, G: 41, B: 59, A: 255},
			0.0,
		)
	} else {
		background = canvas.NewLinearGradient(
			color.NRGBA{R: 248, G: 250, B: 252, A: 255},
			color.NRGBA{R: 241, G: 245, B: 249, A: 255},
			0.0,
		)
	}

	spacer1 := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacer1.SetMinSize(fyne.NewSize(0, 15))
	spacer2 := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacer2.SetMinSize(fyne.NewSize(0, 15))
	spacer3 := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacer3.SetMinSize(fyne.NewSize(0, 20))

	saveButton := PrimaryButton("💾  Save Configuration", theme.DocumentSaveIcon(), func() {
		if err := config.SaveConfig(); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save configuration: %v", err), window)
			return
		}
		dialog.ShowInformation("Saved", "Configuration saved successfully.\nRestart the app for changes to take effect.", window)
	})
	saveButtonContainer := container.NewHBox(layout.NewSpacer(), saveButton, layout.NewSpacer())

	configContainer := container.NewVBox(
		container.NewPadded(pageTitle),
		spacer1,
		container.NewPadded(appGroup),
		container.NewPadded(serverGroup),
		container.NewPadded(transcribeGroup),
		container.NewPadded(ttsGroup),
		spacer2,
		container.NewPadded(saveButtonContainer),
		spacer3,
	)

	scroll := container.NewScroll(configContainer)

	configStack := container.NewStack(background, scroll)

	return container.NewPadded(configStack)
}


// LLM configuration widget references for provider card interaction
var llmBaseUrlEntryRef *widget.Entry
var llmModelEntryRef *widget.Entry
var llmModelSelectRef *widget.Select

func CreateLlmTab() fyne.CanvasObject {
	pageTitle := TitleText("LLM Configuration")

	// Create LLM config form
	llmConfigCard := createLlmConfigGroup()

	// Create API provider shortcut settings area (depends on widget references above)
	providersCard := createApiProvidersCard()

	// Create usage guide card
	guideCard := createLlmGuideCard()

	var background *canvas.LinearGradient
	if GetCurrentThemeIsDark() {
		background = canvas.NewLinearGradient(
			color.NRGBA{R: 15, G: 23, B: 42, A: 255},
			color.NRGBA{R: 30, G: 41, B: 59, A: 255},
			0.0,
		)
	} else {
		background = canvas.NewLinearGradient(
			color.NRGBA{R: 248, G: 250, B: 252, A: 255},
			color.NRGBA{R: 241, G: 245, B: 249, A: 255},
			0.0,
		)
	}

	spacer1 := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacer1.SetMinSize(fyne.NewSize(0, 15))
	spacer2 := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacer2.SetMinSize(fyne.NewSize(0, 15))

	llmContainer := container.NewVBox(
		container.NewPadded(pageTitle),
		spacer1,
		container.NewPadded(providersCard),
		container.NewPadded(llmConfigCard),
		container.NewPadded(guideCard),
		spacer2,
	)

	scroll := container.NewScroll(llmContainer)
	llmStack := container.NewStack(background, scroll)

	return container.NewPadded(llmStack)
}

// Create API provider shortcut card
func createApiProvidersCard() *fyne.Container {
	// Internal tool: set BaseURL and recommended models
	setProvider := func(baseURL string, models []string) {
		if llmBaseUrlEntryRef != nil {
			llmBaseUrlEntryRef.SetText(baseURL)
		}
		if llmModelSelectRef != nil {
			llmModelSelectRef.Options = models
			llmModelSelectRef.Refresh()
			if len(models) > 0 {
				llmModelSelectRef.SetSelected(models[0])
				if llmModelEntryRef != nil {
					llmModelEntryRef.SetText(models[0])
				}
			} else {
				if llmModelEntryRef != nil {
					llmModelEntryRef.SetText("")
				}
			}
		}
	}
	// Qwen Card
	qwenCard := createProviderCard(
		"Aliyun Qwen",
		"Aliyun Large Model Service",
		"https://bailian.console.aliyun.com/",
		color.NRGBA{R: 99, G: 54, B: 231, A: 255}, // Qwen purple
		"qwen",
		func() {
			setProvider("https://dashscope.aliyuncs.com/compatible-mode/v1", []string{
				"qwen-turbo", "qwen-plus", "qwen-max",
			})
		},
	)

	// OpenAI Card
	openaiCard := createProviderCard(
		"OpenAI",
		"GPT Model API Service",
		"https://platform.openai.com/",
		color.NRGBA{R: 116, G: 195, B: 101, A: 255}, // OpenAI green
		"openai",
		func() {
			setProvider("https://api.openai.com/v1", []string{
				"gpt-4o-mini", "gpt-4o", "gpt-4.1-mini", "o3-mini",
			})
		},
	)

	// DeepSeek Card
	deepseekCard := createProviderCard(
		"DeepSeek",
		"High Performance AI Model",
		"https://platform.deepseek.com/",
		color.NRGBA{R: 77, G: 107, B: 254, A: 255}, // DeepSeek blue
		"deepseek",
		func() {
			setProvider("https://api.deepseek.com/v1", []string{
				"deepseek-chat", "deepseek-coder", "DeepSeek-V3", "DeepSeek-R1",
			})
		},
	)

	// Add custom provider card
	addProviderCard := createProviderCard(
		"Add",
		"Add custom provider",
		"https://example.com/krillinai/add-provider", // Placeholder link, can be replaced later
		color.NRGBA{R: 14, G: 165, B: 233, A: 255},   // Cyan accent
		"add",
		func() {
			setProvider("", []string{})
		},
	)

	providersGrid := container.New(
		layout.NewGridLayoutWithColumns(2),
		qwenCard,
		openaiCard,
		deepseekCard,
		addProviderCard,
	)

	return GlassmorphismCard(
		"API Providers",
		"Click cards below to visit platforms and purchase API keys",
		providersGrid,
		GetCurrentThemeIsDark(),
	)
}

// Get provider icon
func getProviderIcon(provider string) fyne.CanvasObject {
	var pngPath string
	switch provider {
	case "qwen":
		pngPath = "source/qwen-color.png"
	case "openai":
		pngPath = "source/openai.png"
	case "deepseek":
		pngPath = "source/deepseek-color.png"
	// case "siliconcloud":
	// 	pngPath = "source/siliconcloud-color.png"
	default:
		return container.NewWithoutLayout()
	}

	data, err := static.EmbeddedFiles.ReadFile(pngPath)
	if err != nil {
		log.GetLogger().Error("Failed to load PNG icon", zap.String("path", pngPath), zap.Error(err))
		return container.NewWithoutLayout()
	}

	res := fyne.NewStaticResource(pngPath, data)
	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(24, 24))
	img.Resize(fyne.NewSize(24, 24))
	return img
}

// Create single provider card
func createProviderCard(name, description, url string, accentColor color.Color, provider string, onTap func()) *fyne.Container {
	isDark := GetCurrentThemeIsDark()

	var bgColor color.Color
	var textColor color.Color
	var descColor color.Color
	var shadowColor color.Color
	var hoverBgColor color.Color

	if isDark {
		bgColor = color.NRGBA{R: 51, G: 65, B: 85, A: 120}
		hoverBgColor = color.NRGBA{R: 71, G: 85, B: 105, A: 150}
		textColor = color.NRGBA{R: 248, G: 250, B: 252, A: 255}
		descColor = color.NRGBA{R: 148, G: 163, B: 184, A: 255}
		shadowColor = color.NRGBA{R: 0, G: 0, B: 0, A: 60}
	} else {
		bgColor = color.NRGBA{R: 255, G: 255, B: 255, A: 200}
		hoverBgColor = color.NRGBA{R: 245, G: 247, B: 250, A: 220}
		textColor = color.NRGBA{R: 17, G: 24, B: 39, A: 255}
		descColor = color.NRGBA{R: 107, G: 114, B: 128, A: 255}
		shadowColor = color.NRGBA{R: 0, G: 0, B: 0, A: 30}
	}

	// Create shadow effect
	shadow := canvas.NewRectangle(shadowColor)
	shadow.CornerRadius = 12
	shadow.Move(fyne.NewPos(2, 2))

	// Background
	background := canvas.NewRectangle(bgColor)
	background.CornerRadius = 12
	background.StrokeColor = accentColor
	background.StrokeWidth = 2

	// Icon
	icon := getProviderIcon(provider)
	// Top padding to avoid icon sticking to edge
	topPadding := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	topPadding.SetMinSize(fyne.NewSize(0, 12))
	// Create container for icon to ensure centering
	iconContainer := container.NewCenter(icon)

	// Title
	nameLabel := canvas.NewText(name, textColor)
	nameLabel.TextSize = 16
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}
	nameLabel.Alignment = fyne.TextAlignCenter

	// Description
	descLabel := canvas.NewText(description, descColor)
	descLabel.TextSize = 12
	descLabel.Alignment = fyne.TextAlignCenter

	// Create clickable container
	content := container.NewVBox(
		topPadding,
		iconContainer,
		container.NewPadded(nameLabel),
		container.NewPadded(descLabel),
	)

	// Create card container with shadow and background
	card := container.NewStack(shadow, background, content)
	card.Resize(fyne.NewSize(200, 100)) // Increase height to accommodate icon

	// Create transparent clickable area
	clickableArea := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	clickableArea.Resize(fyne.NewSize(200, 100))

	// Create custom tappable object
	tappable := &tappableObject{
		rect: clickableArea,
		onTap: func() {
			// Click effect: pressed animation
			originalPos := card.Position()
			originalShadowPos := shadow.Position()

			// Pressed effect: card moves down, shadow shrinks
			card.Move(fyne.NewPos(originalPos.X+1, originalPos.Y+1))
			shadow.Move(fyne.NewPos(originalShadowPos.X+1, originalShadowPos.Y+1))

			// Background color change
			background.FillColor = hoverBgColor
			background.Refresh()

			// Execute callback or open URL if no callback provided
			if onTap != nil {
				onTap()
			} else {
				if app := fyne.CurrentApp(); app != nil && url != "" {
					app.OpenURL(parseURL(url))
				}
			}

			// Restore original position and color
			go func() {
				time.Sleep(150 * time.Millisecond)
				card.Move(fyne.NewPos(0, 0))
				shadow.Move(fyne.NewPos(2, 2))
				background.FillColor = bgColor
				background.Refresh()
			}()
		},
		onHover: func(hovering bool) {
			if hovering {
				// Hover: color/shadow change only to avoid layout jitter
				background.FillColor = hoverBgColor
				background.StrokeWidth = 3
				shadow.Move(fyne.NewPos(3, 3))
				background.Refresh()
			} else {
				background.FillColor = bgColor
				background.StrokeWidth = 2
				shadow.Move(fyne.NewPos(2, 2))
				background.Refresh()
			}
		},
	}

	// Create final container
	finalContainer := container.NewStack(card, tappable)

	return finalContainer
}

// Create LLM usage guide card
func createLlmGuideCard() *fyne.Container {
	guideText := `# LLM Usage Guide:  

## API Base URL: (Choose according to platform)  
   - OpenAI Official: https://api.openai.com/v1  
   - Aliyun Qwen: https://dashscope.aliyuncs.com/compatible-mode/v1  
   - DeepSeek: https://api.deepseek.com/v1  

## API Key:  
   - Obtain from the respective platform's console  
   - Keep it secure and avoid leaking  

## Model Name:  
   - OpenAI: gpt-3.5-turbo, gpt-4, gpt-4-turbo...
   - Aliyun: qwen-turbo, qwen-plus, qwen-max...
   - DeepSeek: deepseek-chat, deepseek-coder...

## Usage Advice:
   - Choose the appropriate model based on your needs
   - Pay attention to API usage costs`
	guideLabel := widget.NewRichTextFromMarkdown(guideText)
	guideLabel.Wrapping = fyne.TextWrapWord

	return GlassmorphismCard(
		"Usage Guide",
		"LLM API configuration instructions",
		guideLabel,
		GetCurrentThemeIsDark(),
	)
}

// Helper function to parse URL
func parseURL(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.GetLogger().Error("Failed to parse URL", zap.Error(err))
		return nil
	}
	return u
}

// Create subtitle task tab
func CreateSubtitleTab(window fyne.Window) fyne.CanvasObject {
	sm := NewSubtitleManager(window)

	title1 := TitleText("Video Translation & Dubbing")
	title2 := TitleText("Video Translate & Dubbing")
	titleContainer := container.NewVBox(title1, title2)

	videoInputContainer := createVideoInputContainer(sm)
	subtitleSettingsCard := createSubtitleSettingsCard(sm)
	voiceSettingsCard := createVoiceSettingsCard(sm)
	embedSettingsCard := createEmbedSettingsCard(sm)

	progress, downloadContainer, tipsLabel := createProgressAndDownloadArea(sm)

	startButton := createStartButton(window, sm, videoInputContainer, embedSettingsCard, progress, downloadContainer)
	startButtonContainer := container.NewHBox(layout.NewSpacer(), startButton, layout.NewSpacer())

	var background *canvas.LinearGradient
	if GetCurrentThemeIsDark() {
		background = canvas.NewLinearGradient(
			color.NRGBA{R: 15, G: 23, B: 42, A: 255},
			color.NRGBA{R: 30, G: 41, B: 59, A: 255},
			0.0,
		)
	} else {
		background = canvas.NewLinearGradient(
			color.NRGBA{R: 248, G: 250, B: 252, A: 255},
			color.NRGBA{R: 241, G: 245, B: 249, A: 255},
			0.0,
		)
	}

	spacer1 := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacer1.SetMinSize(fyne.NewSize(0, 15))
	spacer2 := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacer2.SetMinSize(fyne.NewSize(0, 15))
	spacer3 := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacer3.SetMinSize(fyne.NewSize(0, 15))

	progressArea := container.NewVBox(progress)

	mainContent := container.NewVBox(
		container.NewPadded(titleContainer),
		spacer1,
		container.NewPadded(videoInputContainer),
		container.NewPadded(subtitleSettingsCard),
		container.NewPadded(voiceSettingsCard),
		container.NewPadded(embedSettingsCard),
		spacer2,
		container.NewPadded(startButtonContainer),
		spacer3,
		progressArea,
		downloadContainer,
		tipsLabel,
	)

	scroll := container.NewScroll(mainContent)

	// Use a Stack to combine background and scroll content
	contentStack := container.NewStack(background, scroll)

	return container.NewPadded(contentStack)
}

// Create app config group
func createAppConfigGroup() *fyne.Container {
	appSegmentDurationEntry := StyledEntry("Segment duration (minutes)")
	appSegmentDurationEntry.Bind(binding.IntToString(binding.BindInt(&config.Conf.App.SegmentDuration)))
	appSegmentDurationEntry.Validator = func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("Please enter a number")
		}
		if val < 1 || val > 30 {
			return fmt.Errorf("Please enter a number between 1-30")
		}
		return nil
	}

	appTranscribeParallelNumEntry := StyledEntry("Transcribe Parallel Num")
	appTranscribeParallelNumEntry.Bind(binding.IntToString(binding.BindInt(&config.Conf.App.TranscribeParallelNum)))
	appTranscribeParallelNumEntry.Validator = func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("Please enter a number")
		}
		if val < 1 || val > 10 {
			return fmt.Errorf("Please enter a number between 1-10")
		}
		return nil
	}

	appTranslateParallelNumEntry := StyledEntry("Translate Parallel Num")
	appTranslateParallelNumEntry.Bind(binding.IntToString(binding.BindInt(&config.Conf.App.TranslateParallelNum)))
	appTranslateParallelNumEntry.Validator = func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("Please enter a number")
		}
		if val < 1 || val > 20 {
			return fmt.Errorf("Please enter a number between 1-20")
		}
		return nil
	}

	appTranscribeMaxAttemptsEntry := StyledEntry("Transcribe Max Attempts")
	appTranscribeMaxAttemptsEntry.Bind(binding.IntToString(binding.BindInt(&config.Conf.App.TranscribeMaxAttempts)))
	appTranscribeMaxAttemptsEntry.Validator = func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("Please enter a number")
		}
		if val < 1 || val > 10 {
			return fmt.Errorf("Please enter a number between 1-10")
		}
		return nil
	}

	appTranslateMaxAttemptsEntry := StyledEntry("Translate Max Attempts")
	appTranslateMaxAttemptsEntry.Bind(binding.IntToString(binding.BindInt(&config.Conf.App.TranslateMaxAttempts)))
	appTranslateMaxAttemptsEntry.Validator = func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("Please enter a number")
		}
		if val < 1 || val > 20 {
			return fmt.Errorf("Please enter a number between 1-20")
		}
		return nil
	}

	appMaxSentenceLengthEntry := StyledEntry("Max sentence length")
	appMaxSentenceLengthEntry.Bind(binding.IntToString(binding.BindInt(&config.Conf.App.MaxSentenceLength)))
	appMaxSentenceLengthEntry.Validator = func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("Please enter a number")
		}
		if val < 1 || val > 200 {
			return fmt.Errorf("Please enter a number between 1-200")
		}
		return nil
	}

	appProxyEntry := StyledEntry("Proxy Address")
	appProxyEntry.Bind(binding.BindString(&config.Conf.App.Proxy))

	form := widget.NewForm(
		widget.NewFormItem("Segment duration (minutes)", appSegmentDurationEntry),
		widget.NewFormItem("Transcribe parallel num", appTranscribeParallelNumEntry),
		widget.NewFormItem("Translate parallel num", appTranslateParallelNumEntry),
		widget.NewFormItem("Transcribe max attempts", appTranscribeMaxAttemptsEntry),
		widget.NewFormItem("Translate max attempts", appTranslateMaxAttemptsEntry),
		widget.NewFormItem("Max sentence length", appMaxSentenceLengthEntry),
		widget.NewFormItem("Proxy address", appProxyEntry),
	)

	return GlassmorphismCard("App Configuration", "Basic Parameters", form, GetCurrentThemeIsDark())
}

// Create server config group
func createServerConfigGroup() *fyne.Container {
	serverHostEntry := StyledEntry("Server address")
	serverHostEntry.Bind(binding.BindString(&config.Conf.Server.Host))

	serverPortEntry := StyledEntry("Server port")
	serverPortEntry.Bind(binding.IntToString(binding.BindInt(&config.Conf.Server.Port)))
	serverPortEntry.Validator = func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("Please enter a number")
		}
		if val < 1 || val > 65535 {
			return fmt.Errorf("Please enter a valid port (1-65535)")
		}
		return nil
	}

	form := widget.NewForm(
		widget.NewFormItem("Server Address", serverHostEntry),
		widget.NewFormItem("Server Port", serverPortEntry),
	)

	return GlassmorphismCard("Server Configuration", "API Server Settings", form, GetCurrentThemeIsDark())
}

// Create LLM config group
func createLlmConfigGroup() *fyne.Container {
	baseUrlEntry := StyledEntry("API Base URL")
	baseUrlEntry.Bind(binding.BindString(&config.Conf.Llm.BaseUrl))
	llmBaseUrlEntryRef = baseUrlEntry

	apiKeyEntry := StyledPasswordEntry("API Key")
	apiKeyEntry.Bind(binding.BindString(&config.Conf.Llm.ApiKey))

	modelEntry := StyledEntry("Model Name")
	modelEntry.Bind(binding.BindString(&config.Conf.Llm.Model))
	llmModelEntryRef = modelEntry

	// Recommended model dropdown (display only, sync to text box on selection)
	modelSelect := StyledSelect([]string{}, func(v string) {
		if v != "" && llmModelEntryRef != nil {
			llmModelEntryRef.SetText(v)
		}
	})
	modelSelect.PlaceHolder = "Select recommended model (optional)"
	llmModelSelectRef = modelSelect

	form := widget.NewForm(
		widget.NewFormItem("API Base URL", baseUrlEntry),
		widget.NewFormItem("API Key", apiKeyEntry),
		widget.NewFormItem("Model Name", modelEntry),
		widget.NewFormItem("Supported Models", modelSelect),
	)
	return GlassmorphismCard("LLM Configuration", "LLM Settings", form, GetCurrentThemeIsDark())
}

// Create transcription config group
func createTranscribeConfigGroup() *fyne.Container {
	providerOptions := []string{"openai", "fasterwhisper", "whisperkit", "whispercpp"}
	providerSelect := widget.NewSelect(providerOptions, func(value string) {
		config.Conf.Transcribe.Provider = value
	})
	providerSelect.SetSelected(config.Conf.Transcribe.Provider)

	openaiBaseUrlEntry := StyledEntry("API Base URL")
	openaiBaseUrlEntry.Bind(binding.BindString(&config.Conf.Transcribe.Openai.BaseUrl))
	openaiApiKeyEntry := StyledPasswordEntry("API Key")
	openaiApiKeyEntry.Bind(binding.BindString(&config.Conf.Transcribe.Openai.ApiKey))
	openaiModelEntry := StyledEntry("Model Name")
	openaiModelEntry.Bind(binding.BindString(&config.Conf.Transcribe.Openai.Model))

	fasterWhisperModelEntry := StyledEntry("Model Name")
	fasterWhisperModelEntry.Bind(binding.BindString(&config.Conf.Transcribe.Fasterwhisper.Model))

	whisperKitModelEntry := StyledEntry("Model Name")
	whisperKitModelEntry.Bind(binding.BindString(&config.Conf.Transcribe.Whisperkit.Model))

	whisperCppModelEntry := StyledEntry("Model Name")
	whisperCppModelEntry.Bind(binding.BindString(&config.Conf.Transcribe.Whispercpp.Model))

	form := widget.NewForm(
		widget.NewFormItem("Provider", providerSelect),
		widget.NewFormItem("GPU Acceleration", widget.NewCheckWithData("Enable", binding.BindBool(&config.Conf.Transcribe.EnableGpuAcceleration))),

		widget.NewFormItem("OpenAI Base URL", openaiBaseUrlEntry),
		widget.NewFormItem("OpenAI API Key", openaiApiKeyEntry),
		widget.NewFormItem("OpenAI Model", openaiModelEntry),

		widget.NewFormItem("FasterWhisper Model", fasterWhisperModelEntry),

		widget.NewFormItem("WhisperKit Model", whisperKitModelEntry),

		widget.NewFormItem("WhisperCpp Model", whisperCppModelEntry),
	)

	return GlassmorphismCard("Transcription Configuration", "Transcription Settings", form, GetCurrentThemeIsDark())
}

// Create TTS config group
func createTtsConfigGroup() *fyne.Container {
	providerOptions := []string{"openai", "edge-tts", "vclip"}
	providerSelect := widget.NewSelect(providerOptions, func(value string) {
		config.Conf.Tts.Provider = value
	})
	providerSelect.SetSelected(config.Conf.Tts.Provider)

	openaiBaseUrlEntry := StyledEntry("API Base URL")
	openaiBaseUrlEntry.Bind(binding.BindString(&config.Conf.Tts.Openai.BaseUrl))
	openaiApiKeyEntry := StyledPasswordEntry("API Key")
	openaiApiKeyEntry.Bind(binding.BindString(&config.Conf.Tts.Openai.ApiKey))
	openaiModelEntry := StyledEntry("Model Name (e.g., tts-1)")
	openaiModelEntry.Bind(binding.BindString(&config.Conf.Tts.Openai.Model))
	openaiVoiceEntry := StyledEntry("Default Voice (e.g., alloy)")
	openaiVoiceEntry.Bind(binding.BindString(&config.Conf.Tts.Openai.Voice))

	edgeTtsVoiceEntry := StyledEntry("Edge-TTS Voice (e.g., en-US-AndrewNeural)")
	edgeTtsVoiceEntry.Bind(binding.BindString(&config.Conf.Tts.EdgeTts.Voice))

	vclipApiKeyEntry := StyledPasswordEntry("VClip API Key")
	vclipApiKeyEntry.Bind(binding.BindString(&config.Conf.Tts.VClip.ApiKey))
	vclipVoiceIdEntry := StyledEntry("VClip Voice ID (userVoiceId)")
	vclipVoiceIdEntry.Bind(binding.BindString(&config.Conf.Tts.VClip.VoiceID))
	vclipSpeedEntry := StyledEntry("Speed (0.5 - 2.0, default 1.0)")
	vclipSpeedEntry.Bind(binding.FloatToString(binding.BindFloat(&config.Conf.Tts.VClip.Speed)))

	form := widget.NewForm(
		widget.NewFormItem("Provider", providerSelect),

		widget.NewFormItem("OpenAI Base URL", openaiBaseUrlEntry),
		widget.NewFormItem("OpenAI API Key", openaiApiKeyEntry),
		widget.NewFormItem("OpenAI Model", openaiModelEntry),
		widget.NewFormItem("OpenAI Voice", openaiVoiceEntry),

		widget.NewFormItem("Edge-TTS Voice", edgeTtsVoiceEntry),

		widget.NewFormItem("VClip API Key", vclipApiKeyEntry),
		widget.NewFormItem("VClip Voice ID", vclipVoiceIdEntry),
		widget.NewFormItem("VClip Speed", vclipSpeedEntry),
	)

	return GlassmorphismCard("TTS Configuration", "TTS Settings", form, GetCurrentThemeIsDark())
}


// Create video input container
func createVideoInputContainer(sm *SubtitleManager) *fyne.Container {
	inputTypeRadio := widget.NewRadioGroup([]string{"Local Upload", "Paste Link"}, nil)
	inputTypeRadio.Horizontal = true
	inputTypeContainer := container.NewHBox(
		inputTypeRadio,
	)

	urlEntry := StyledEntry("Paste video link here")
	urlEntry.Hide()
	urlEntry.OnChanged = func(text string) {
		sm.SetVideoUrl(text)
	}

	selectButton := PrimaryButton("Select Video Files", theme.FolderOpenIcon(), sm.ShowFileDialog)

	selectedVideoLabel := widget.NewLabel("")
	selectedVideoLabel.Hide()

	sm.SetVideoSelectedCallback(func(path string) { // Set video URL and control display info
		if path != "" {
			sm.SetVideoUrl(path)
			selectedVideoLabel.SetText("Selected: " + filepath.Base(path))
			selectedVideoLabel.Show()
		} else {
			selectedVideoLabel.Hide()
		}
	})

	sm.SetVideosSelectedCallback(func(paths []string) {
		if len(paths) > 0 {
			sm.SetVideoUrl(paths[0])

			fileNames := make([]string, 0, len(paths))
			for _, path := range paths {
				fileNames = append(fileNames, filepath.Base(path))
			}

			displayText := fmt.Sprintf("Selected %d files:\n", len(paths))
			for i, name := range fileNames {
				displayText += fmt.Sprintf("%d. %s\n", i+1, name)
			}

			selectedVideoLabel.SetText(displayText)
			selectedVideoLabel.Show()
		} else {
			selectedVideoLabel.Hide()
		}
	})

	videoInputContainer := container.NewVBox()
	videoInputContainer.Objects = []fyne.CanvasObject{selectButton, selectedVideoLabel}

	inputTypeRadio.SetSelected("Local Upload")
	inputTypeRadio.OnChanged = func(value string) {
		if value == "Local Upload" {
			urlEntry.Hide()
			selectButton.Show()
			selectedVideoLabel.Show()
			videoInputContainer.Objects = []fyne.CanvasObject{selectButton, selectedVideoLabel}
			sm.SetVideoUrl("")
		} else {
			selectButton.Hide()
			selectedVideoLabel.Hide()
			urlEntry.Show()
			videoInputContainer.Objects = []fyne.CanvasObject{urlEntry}
		}
		videoInputContainer.Refresh()
	}

	content := container.NewVBox(
		container.NewPadded(inputTypeContainer),
		container.NewPadded(videoInputContainer),
	)

	return GlassmorphismCard("1. Select Video", "", content, GetCurrentThemeIsDark())
}

// Create subtitle settings card
func createSubtitleSettingsCard(sm *SubtitleManager) *fyne.Container {
	positionSelect := widget.NewSelect([]string{
		"Translation Above",
		"Translation Below",
	}, func(value string) {
		if value == "Translation Above" {
			sm.SetBilingualPosition(1)
		} else {
			sm.SetBilingualPosition(2)
		}
	})
	positionSelect.SetSelected("Translation Above")

	bilingualCheck := widget.NewCheck("Bilingual Subtitles", func(checked bool) {
		sm.SetBilingualEnabled(checked)
		if checked {
			positionSelect.Enable()
		} else {
			positionSelect.Disable()
		}
	})
	bilingualCheck.SetChecked(true)

	var targetSelectOptions []string
	targetLangMap := make(map[string]string)
	for code, name := range types.StandardLanguageCode2Name {
		targetSelectOptions = append(targetSelectOptions, name)
		targetLangMap[name] = string(code)
	}
	targetLangSelector := StyledSelect(targetSelectOptions, func(value string) {
		sm.SetTargetLang(targetLangMap[value])
	})

	langContainer := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Original Language:"),
			StyledSelect([]string{
				"Simplified Chinese", "English", "Japanese", "Turkish", "German", "Korean", "Russian", "Malay",
			}, func(value string) {
				sourceLangMap := map[string]string{
					"Simplified Chinese": "zh_cn", "English": "en", "Japanese": "ja",
					"Turkish": "tr", "German": "de", "Korean": "ko", "Russian": "ru",
					"Malay": "ms",
				}
				sm.SetSourceLang(sourceLangMap[value])
			}),
		),
		container.NewHBox(
			widget.NewLabel("Translate To:"),
			targetLangSelector,
		),
	)

	// Set default languages
	langContainer.Objects[0].(*fyne.Container).Objects[1].(*widget.Select).SetSelected("English")
	langContainer.Objects[1].(*fyne.Container).Objects[1].(*widget.Select).SetSelected("Simplified Chinese")

	fillerCheck := widget.NewCheck("Tone Word Filtering", func(checked bool) {
		sm.SetFillerFilter(checked)
	})
	fillerCheck.SetChecked(true)

	reviewCheck := widget.NewCheck("Review Subtitles before TTS", func(checked bool) {
		sm.SetEnableReview(checked)
	})
	reviewCheck.SetChecked(true)

	content := container.NewVBox(
		container.NewHBox(bilingualCheck, fillerCheck, reviewCheck),
		langContainer,
		positionSelect,
	)

	return ModernCard("2. Subtitle Settings", content, GetCurrentThemeIsDark())
}

// Create dubbing settings card
func createVoiceSettingsCard(sm *SubtitleManager) *fyne.Container {
	voiceCodeEntry := widget.NewEntry()
	voiceCodeEntry.SetPlaceHolder("Enter voice code")

	// Set default voice from config based on provider
	defaultVoice := ""
	switch config.Conf.Tts.Provider {
	case "openai":
		defaultVoice = config.Conf.Tts.Openai.Voice
	case "edge-tts":
		defaultVoice = config.Conf.Tts.EdgeTts.Voice
	case "vclip":
		defaultVoice = config.Conf.Tts.VClip.VoiceID
	}
	voiceCodeEntry.SetText(defaultVoice)
	sm.SetTtsVoiceCode(defaultVoice)

	voiceCodeEntry.OnChanged = func(text string) {
		sm.SetTtsVoiceCode(text)
	}
	voiceCodeEntry.Disable()

	voiceoverCheck := widget.NewCheck("Apply Dubbing", func(checked bool) {
		sm.SetVoiceoverEnabled(checked)
		if checked {
			voiceCodeEntry.Enable()
		} else {
			voiceCodeEntry.Disable()
		}
	})

	grid := container.NewVBox(
		container.NewHBox(voiceoverCheck),
		container.NewPadded(voiceCodeEntry),
	)

	return ModernCard("3. Dubbing Settings", grid, GetCurrentThemeIsDark())
}


// Video composition card
func createEmbedSettingsCard(sm *SubtitleManager) *fyne.Container {
	embedCheck := widget.NewCheck("Composite", nil)

	embedTypeSelect := StyledSelect([]string{
		"Landscape (16:9)", "Portrait (9:16)", "Landscape + Portrait",
	}, nil)
	embedTypeSelect.Disable()

	mainTitleEntry := StyledEntry("Enter main title")
	subTitleEntry := StyledEntry("Enter sub title")

	titleInputContainer := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("Main Title:"),
			mainTitleEntry,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Sub Title:"),
			subTitleEntry,
		),
	)
	titleInputContainer.Hide()

	embedCheck.OnChanged = func(checked bool) {
		if checked {
			embedTypeSelect.Enable()
			embedTypeSelect.SetSelected("Landscape (16:9)")
		} else {
			embedTypeSelect.Disable()
			sm.SetEmbedSubtitle("none")
		}
	}

	embedTypeSelect.OnChanged = func(value string) {
		switch value {
		case "Landscape (16:9)":
			titleInputContainer.Hide()
			sm.SetEmbedSubtitle("horizontal")
		case "Portrait (9:16)":
			titleInputContainer.Show()
			sm.SetEmbedSubtitle("vertical")
		case "Landscape + Portrait":
			titleInputContainer.Show()
			sm.SetEmbedSubtitle("all")
		}
	}

	topContainer := container.NewHBox(embedCheck, embedTypeSelect)

	mainContainer := container.NewVBox(
		topContainer,
		container.NewPadded(titleInputContainer),
	)

	return ModernCard("Composition Settings", mainContainer, GetCurrentThemeIsDark())
}

// Create progress and download area
func createProgressAndDownloadArea(sm *SubtitleManager) (*widget.ProgressBar, *fyne.Container, *fyne.Container) {
	progress := widget.NewProgressBar()
	progress.Hide()

	percentLabel := widget.NewLabel("0%")
	percentLabel.Hide()
	percentLabel.Alignment = fyne.TextAlignTrailing

	progressContainer := container.NewBorder(nil, nil, nil, percentLabel, progress)
	progressContainer.Hide()

	progressBg := canvas.NewRectangle(color.NRGBA{R: 240, G: 245, B: 250, A: 230})
	progressBg.SetMinSize(fyne.NewSize(0, 40))
	progressBg.CornerRadius = 8

	progressShadow := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 20})
	progressShadow.Move(fyne.NewPos(2, 2))
	progressShadow.SetMinSize(fyne.NewSize(0, 40))
	progressShadow.CornerRadius = 8

	progressWithBg := container.NewStack(
		progressShadow,
		progressBg,
		container.NewPadded(progressContainer),
	)
	progressWithBg.Hide()

	sm.SetProgressBar(progress)
	sm.SetProgressLabel(percentLabel)

	downloadBg := canvas.NewRectangle(color.NRGBA{R: 240, G: 250, B: 255, A: 230})
	downloadBg.CornerRadius = 10

	downloadContainer := container.NewVBox()
	downloadContainer.Hide()
	sm.SetDownloadContainer(downloadContainer)

	downloadWithBg := container.NewStack(
		downloadBg,
		container.NewPadded(downloadContainer),
	)
	downloadWithBg.Hide()

	tipsLabel := widget.NewLabel("")
	tipsLabel.Hide()
	tipsLabel.Alignment = fyne.TextAlignCenter
	tipsLabel.Wrapping = fyne.TextWrapWord
	sm.SetTipsLabel(tipsLabel)

	tipsBg := canvas.NewRectangle(color.NRGBA{R: 255, G: 250, B: 230, A: 200})
	tipsBg.CornerRadius = 6

	tipsWithBg := container.NewStack(
		tipsBg,
		container.NewPadded(tipsLabel),
	)
	tipsWithBg.Hide()

	return progress, downloadWithBg, tipsWithBg
}

// Start button
func createStartButton(window fyne.Window, sm *SubtitleManager, videoInputContainer *fyne.Container, embedSettingsCard *fyne.Container, progress *widget.ProgressBar, downloadContainer *fyne.Container) *widget.Button {
	btn := widget.NewButtonWithIcon("Start Translating", theme.MediaPlayIcon(), nil)
	btn.Importance = widget.HighImportance

	btn.OnTapped = func() {
		originalImportance := btn.Importance
		btn.Importance = widget.DangerImportance
		btn.Refresh()

		go func() {
			time.Sleep(300 * time.Millisecond)
			btn.Importance = originalImportance
			btn.Refresh()
		}()

		var mainTitle, subTitle string

		if embedSettingsCard != nil && len(embedSettingsCard.Objects) > 1 {
			if titleContainer, ok := embedSettingsCard.Objects[1].(*fyne.Container); ok && titleContainer != nil && len(titleContainer.Objects) >= 2 {
				if mainTitleRow, ok := titleContainer.Objects[0].(*fyne.Container); ok && mainTitleRow != nil && len(mainTitleRow.Objects) >= 2 {
					if mainTitleEntry, ok := mainTitleRow.Objects[1].(*widget.Entry); ok {
						mainTitle = mainTitleEntry.Text
					}
				}

				if subTitleRow, ok := titleContainer.Objects[1].(*fyne.Container); ok && subTitleRow != nil && len(subTitleRow.Objects) >= 2 {
					if subTitleEntry, ok := subTitleRow.Objects[1].(*widget.Entry); ok {
						subTitle = subTitleEntry.Text
					}
				}
			}
		}

		sm.SetVerticalTitles(mainTitle, subTitle)

		progress.Show()
		sm.progressBar.SetValue(0)
		downloadContainer.Hide()

		if sm.GetVideoUrl() == "" {
			inputType := "Local Video"

			if videoInputContainer != nil && len(videoInputContainer.Objects) > 0 {
				for i := 0; i < len(videoInputContainer.Objects); i++ {
					// If object is Container, find RadioGroup within it
					if container, ok := videoInputContainer.Objects[i].(*fyne.Container); ok {
						for j := 0; j < len(container.Objects); j++ {
							if radio, ok := container.Objects[j].(*widget.RadioGroup); ok {
								inputType = radio.Selected
								break
							}
						}
					}
				}
			}

			if inputType == "Local Video" {
				dialog.ShowError(fmt.Errorf("Please select video files first"), window)
			} else {
				dialog.ShowError(fmt.Errorf("Please enter a video link"), window)
			}
			progress.Hide()
			return
		}

		err := config.CheckConfig()
		if err != nil {
			dialog.ShowError(fmt.Errorf("Incorrect configuration: %v", err), window)
			log.GetLogger().Error("Incorrect configuration", zap.Error(err))
			progress.Hide()
			return
		}

		err = deps.CheckDependency()
		if err != nil {
			dialog.ShowError(fmt.Errorf("Dependency environment preparation failed: %v", err), window)
			log.GetLogger().Error("Dependency environment preparation failed", zap.Error(err))
			progress.Hide()
			return
		}
		btn.Hide()

		if config.ConfigBackup != config.Conf {
			if err = server.StopBackend(); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to stop backend service: %v", err), window)
				log.GetLogger().Error("Failed to stop backend service", zap.Error(err))
				progress.Hide()
				return
			}

			go func() {
				err := server.StartBackend()
				if err != nil {
					dialog.ShowError(fmt.Errorf("Failed to start backend service: %v", err), window)
					log.GetLogger().Error("Failed to start backend service", zap.Error(err))
					progress.Hide()
					return
				}
			}()

			time.Sleep(1 * time.Second)
			config.ConfigBackup = config.Conf
		}

		if err = sm.StartTask(); err != nil {
			dialog.ShowError(err, window)
			progress.Hide()
			return
		}

		go func() {
			for {
				time.Sleep(1 * time.Second)
				if sm.progressBar.Value < 1 {
					continue
				}
				time.Sleep(1 * time.Second)
				if sm.progressBar.Value < 1 {
					continue
				}
				break
			}
			btn.Show()
			downloadContainer.Show()
		}()
		sm.progressBar.Refresh()
	}

	return btn
}
