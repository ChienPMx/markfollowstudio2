package deps

import (
	"fmt"
	"markflow-studio/config"
	"markflow-studio/internal/storage"
	"markflow-studio/log"
	"markflow-studio/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
)

func CheckDependency() error {
	err := checkAndDownloadFfmpeg()
	if err != nil {
		log.GetLogger().Error("Failed to prepare ffmpeg environment", zap.Error(err))
		return err
	}
	err = checkAndDownloadFfprobe()
	if err != nil {
		log.GetLogger().Error("Failed to prepare ffprobe environment", zap.Error(err))
		return err
	}
	err = checkAndDownloadYtDlp()
	if err != nil {
		log.GetLogger().Error("Failed to prepare yt-dlp environment", zap.Error(err))
		return err
	}
	if config.Conf.Transcribe.Provider == "fasterwhisper" {
		err = checkFasterWhisper()
		if err != nil {
			log.GetLogger().Error("Failed to prepare fasterwhisper environment", zap.Error(err))
			return err
		}
		err = checkModel("fasterwhisper")
		if err != nil {
			log.GetLogger().Error("Failed to prepare local model environment", zap.Error(err))
			return err
		}
	}
	if config.Conf.Transcribe.Provider == "whisperkit" {
		if err = checkWhisperKit(); err != nil {
			log.GetLogger().Error("Failed to prepare whisperkit environment", zap.Error(err))
			return err
		}
		err = checkModel("whisperkit")
		if err != nil {
			log.GetLogger().Error("Failed to prepare local model environment", zap.Error(err))
			return err
		}
	}
	if config.Conf.Transcribe.Provider == "whisperx" {
		err = checkWhisperX()
		if err != nil {
			log.GetLogger().Error("Failed to prepare whisperx environment", zap.Error(err))
			return err
		}
		err = checkModel("whisperx")
		if err != nil {
			log.GetLogger().Error("Failed to prepare local model environment", zap.Error(err))
			return err
		}
	}
	if config.Conf.Transcribe.Provider == "whispercpp" {
		if err = checkWhispercpp(); err != nil {
			log.GetLogger().Error("Failed to prepare whispercpp environment", zap.Error(err))
			return err
		}
		err = checkModel("whispercpp")
		if err != nil {
			log.GetLogger().Error("Failed to prepare local whispercpp model environment", zap.Error(err))
			return err
		}
	}


	return nil
}

// Detect and install ffmpeg
func checkAndDownloadFfmpeg() error {
	// Check if ffmpeg is already installed
	_, err := exec.LookPath("ffmpeg")
	if err == nil {
		log.GetLogger().Info("Found ffmpeg")
		storage.FfmpegPath = "ffmpeg"
		return nil
	}

	ffmpegBinFilePath := "./bin/ffmpeg"
	if runtime.GOOS == "windows" {
		ffmpegBinFilePath += ".exe"
	}
	// Previous download
	if _, err = os.Stat(ffmpegBinFilePath); err == nil {
		log.GetLogger().Info("Found ffmpeg")
		storage.FfmpegPath = ffmpegBinFilePath
		return nil
	}

	log.GetLogger().Info("ffmpeg not found, starting automatic installation...")
	// Ensure ./bin directory exists
	err = os.MkdirAll("./bin", 0755)
	if err != nil {
		log.GetLogger().Error("Failed to create ./bin directory", zap.Error(err))
		return err
	}

	var ffmpegURL string
	if runtime.GOOS == "linux" {
		ffmpegURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/ffmpeg-6.1-linux-64.zip"
	} else if runtime.GOOS == "darwin" {
		ffmpegURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/ffmpeg-6.1-macos-64.zip"
	} else if runtime.GOOS == "windows" {
		ffmpegURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/ffmpeg-6.1-win-64.zip"
	} else {
		log.GetLogger().Error("Unsupported OS", zap.String("Current OS", runtime.GOOS))
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Download
	ffmpegDownloadPath := "./bin/ffmpeg.zip"
	err = util.DownloadFile(ffmpegURL, ffmpegDownloadPath, config.Conf.App.Proxy)
	if err != nil {
		log.GetLogger().Error("Failed to download ffmpeg", zap.Error(err))
		return err
	}
	err = util.Unzip(ffmpegDownloadPath, "./bin")
	if err != nil {
		log.GetLogger().Error("Failed to unzip ffmpeg", zap.Error(err))
		return err
	}
	log.GetLogger().Info("Successfully unzipped ffmpeg")

	if runtime.GOOS != "windows" {
		err = os.Chmod(ffmpegBinFilePath, 0755)
		if err != nil {
			log.GetLogger().Error("Failed to set file permissions", zap.Error(err))
			return err
		}
	}

	storage.FfmpegPath = ffmpegBinFilePath
	log.GetLogger().Info("ffmpeg installation complete", zap.String("Path", ffmpegBinFilePath))

	return nil
}

// Detect and install ffprobe
func checkAndDownloadFfprobe() error {
	// Check if ffprobe is already installed
	_, err := exec.LookPath("ffprobe")
	if err == nil {
		log.GetLogger().Info("Found ffprobe")
		storage.FfprobePath = "ffprobe"
		return nil
	}

	ffprobeBinFilePath := "./bin/ffprobe"
	if runtime.GOOS == "windows" {
		ffprobeBinFilePath += ".exe"
	}
	// Previous download
	if _, err = os.Stat(ffprobeBinFilePath); err == nil {
		log.GetLogger().Info("Found ffprobe")
		storage.FfprobePath = ffprobeBinFilePath
		return nil
	}

	log.GetLogger().Info("ffprobe not found, starting automatic installation...")
	// Ensure ./bin directory exists
	err = os.MkdirAll("./bin", 0755)
	if err != nil {
		log.GetLogger().Error("Failed to create ./bin directory", zap.Error(err))
		return err
	}

	var ffprobeURL string
	if runtime.GOOS == "linux" {
		ffprobeURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/ffprobe-6.1-linux-64.zip"
	} else if runtime.GOOS == "darwin" {
		ffprobeURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/ffprobe-6.1-macos-64.zip"
	} else if runtime.GOOS == "windows" {
		ffprobeURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/ffprobe-6.1-win-64.zip"
	} else {
		log.GetLogger().Error("Unsupported OS", zap.String("Current OS", runtime.GOOS))
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Download
	ffprobeDownloadPath := "./bin/ffprobe.zip"
	err = util.DownloadFile(ffprobeURL, ffprobeDownloadPath, config.Conf.App.Proxy)
	if err != nil {
		log.GetLogger().Error("Failed to download ffprobe", zap.Error(err))
		return err
	}
	err = util.Unzip(ffprobeDownloadPath, "./bin")
	if err != nil {
		log.GetLogger().Error("Failed to unzip ffprobe", zap.Error(err))
		return err
	}
	log.GetLogger().Info("Successfully unzipped ffprobe")

	if runtime.GOOS != "windows" {
		err = os.Chmod(ffprobeBinFilePath, 0755)
		if err != nil {
			log.GetLogger().Error("Failed to set file permissions", zap.Error(err))
			return err
		}
	}

	storage.FfprobePath = ffprobeBinFilePath
	log.GetLogger().Info("ffprobe installation complete", zap.String("Path", ffprobeBinFilePath))

	return nil
}

// Detect and install yt-dlp
func checkAndDownloadYtDlp() error {
	_, err := exec.LookPath("yt-dlp")
	if err == nil {
		log.GetLogger().Info("Found yt-dlp")
		storage.YtdlpPath = "yt-dlp"
		return nil
	}

	ytdlpBinFilePath := "./bin/yt-dlp"
	if runtime.GOOS == "windows" {
		ytdlpBinFilePath += ".exe"
	}
	// Previous download
	if _, err = os.Stat(ytdlpBinFilePath); err == nil {
		log.GetLogger().Info("Found yt-dlp")
		storage.YtdlpPath = ytdlpBinFilePath
		return nil
	}

	log.GetLogger().Info("yt-dlp not found, starting automatic installation...")
	err = os.MkdirAll("./bin", 0755)
	if err != nil {
		log.GetLogger().Error("Failed to create ./bin directory", zap.Error(err))
		return err
	}

	var ytDlpURL string
	if runtime.GOOS == "linux" {
		ytDlpURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/yt-dlp_linux"
	} else if runtime.GOOS == "darwin" {
		ytDlpURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/yt-dlp_macos"
	} else if runtime.GOOS == "windows" {
		ytDlpURL = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/yt-dlp.exe"
	} else {
		log.GetLogger().Error("Unsupported OS", zap.String("Current OS", runtime.GOOS))
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	err = util.DownloadFile(ytDlpURL, ytdlpBinFilePath, config.Conf.App.Proxy)
	if err != nil {
		log.GetLogger().Error("Failed to download yt-dlp", zap.Error(err))
		return err
	}

	if runtime.GOOS != "windows" {
		err = os.Chmod(ytdlpBinFilePath, 0755)
		if err != nil {
			log.GetLogger().Error("Failed to set file permissions", zap.Error(err))
			return err
		}
	}

	storage.YtdlpPath = ytdlpBinFilePath
	log.GetLogger().Info("yt-dlp installation complete", zap.String("Path", ytdlpBinFilePath))

	return nil
}

// Detect faster-whisper
func checkFasterWhisper() error {
	var (
		filePath string
		err      error
	)
	if runtime.GOOS == "windows" {
		filePath = "./bin/faster-whisper/Faster-Whisper-XXL/faster-whisper-xxl.exe"
	} else if runtime.GOOS == "linux" {
		filePath = "./bin/faster-whisper/Whisper-Faster-XXL/whisper-faster-xxl"
	} else {
		return fmt.Errorf("fasterwhisper does not support your OS: %s, please choose another provider", runtime.GOOS)
	}
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		log.GetLogger().Info("faster-whisper not found, starting automatic download (please wait...)")
		err = os.MkdirAll("./bin", 0755)
		if err != nil {
			log.GetLogger().Error("Failed to create ./bin directory", zap.Error(err))
			return err
		}
		var downloadUrl string
		if runtime.GOOS == "windows" {
			downloadUrl = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/Faster-Whisper-XXL_r194.5_windows.zip"
		} else {
			downloadUrl = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/Faster-Whisper-XXL_r192.3.1_linux.zip"
		}
		err = util.DownloadFile(downloadUrl, "./bin/faster-whisper.zip", config.Conf.App.Proxy)
		if err != nil {
			log.GetLogger().Error("Failed to download faster-whisper", zap.Error(err))
			return err
		}
		log.GetLogger().Info("Starting to unzip faster-whisper")
		err = util.Unzip("./bin/faster-whisper.zip", "./bin/faster-whisper/")
		if err != nil {
			log.GetLogger().Error("Failed to unzip faster-whisper", zap.Error(err))
			return err
		}
	}
	if runtime.GOOS != "windows" {
		err = os.Chmod(filePath, 0755)
		if err != nil {
			log.GetLogger().Error("Failed to set file permissions", zap.Error(err))
			return err
		}
	}
	storage.FasterwhisperPath = filePath
	log.GetLogger().Info("faster-whisper check complete", zap.String("Path", filePath))
	return nil
}

// Detect whisperkit
func checkWhisperKit() error {
	cmd := exec.Command("which", "whisperkit-cli")
	err := cmd.Run()
	if err != nil {
		log.GetLogger().Info("whisperkit-cli not found, starting automatic installation...")
		cmd = exec.Command("brew", "install", "whisperkit-cli")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.GetLogger().Error("Failed to install whisperkit-cli", zap.String("Info", string(output)), zap.Error(err))
			return err
		}
		log.GetLogger().Info("Successfully installed whisperkit-cli")
	}
	storage.WhisperKitPath = "whisperkit-cli"
	log.GetLogger().Info("Detected whisperkit-cli installed")
	return nil
}

// Detect whisperx
func checkWhisperX() error {
	var (
		filePath  string
		_filePath string
		err       error
	)
	if runtime.GOOS == "windows" {
		filePath = "whisperx"
		_filePath = ".\\bin\\whisperx\\.venv\\Scripts\\whisperx.exe"
	} else if runtime.GOOS == "linux" {
		filePath = "./bin/whisperx/.venv/bin/whisperx"
		_filePath = "./bin/whisperx/.venv/bin/whisperx"
	} else {
		return fmt.Errorf("WhisperX does not support your OS: %s, please choose WhisperKit", runtime.GOOS)
	}

	if _, err = os.Stat(_filePath); os.IsNotExist(err) {
		var cmd *exec.Cmd
		// TODO: Download compressed package
		// log.GetLogger().Info("WhisperX not found, starting automatic download, large file please wait...")
		// err = os.MkdirAll("./bin", 0755)
		// if err != nil {
		// 	log.GetLogger().Error("Failed to create ./bin directory", zap.Error(err))
		// 	return err
		// }
		// var downloadUrl string
		// if runtime.GOOS == "windows" {
		// 	downloadUrl = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/WhisperX_win.zip"
		// } else if runtime.GOOS == "darwin" {
		// 	downloadUrl = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/WhisperX_linux.zip"
		// } else {
		// 	downloadUrl = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/WhisperX_mac.zip"
		// }
		// err = util.DownloadFile(downloadUrl, "./bin/WhisperX.zip", config.Conf.App.Proxy)
		// if err != nil {
		// 	log.GetLogger().Error("Failed to download WhisperX", zap.Error(err))
		// 	return err
		// }
		log.GetLogger().Info("Starting to unzip WhisperX")
		err = util.Unzip("./bin/WhisperX.zip", "./bin/whisperx/")
		if err != nil {
			log.GetLogger().Error("Failed to unzip WhisperX", zap.Error(err))
			return err
		}
		if runtime.GOOS == "windows" {
			cmd = exec.Command(".\\bin\\whisperx\\python\\python.exe", "-m", "venv", ".\\bin\\whisperx\\.venv")
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.GetLogger().Error("Failed to create Python virtual environment", zap.String("Info", string(output)), zap.Error(err))
				return err
			}
			cmd = exec.Command(".\\bin\\whisperx\\.venv\\Scripts\\activate", "&&", "pip", "install", "-r", ".\\bin\\whisperx\\requirements_win.txt")
			cmd.CombinedOutput()
		} else {
			os.Chmod("./bin/whisperx/python/bin/python3.12", 0755)
			os.Chmod("./bin/whisperx/install.sh", 0755)
			log.GetLogger().Info("Starting WhisperX installation")
			cmd = exec.Command("bash", "./bin/whisperx/install.sh")
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.GetLogger().Error("Failed to install WhisperX", zap.String("Info", string(output)), zap.Error(err))
				return err
			}
		}
		log.GetLogger().Info("Successfully installed WhisperX")
	}

	storage.WhisperXPath = filePath
	log.GetLogger().Info("WhisperX check complete", zap.String("Path", _filePath))
	return nil
}

// Detect whispercpp
func checkWhispercpp() error {
	var (
		filePath string
		err      error
	)
	if runtime.GOOS == "windows" {
		filePath = filepath.Join("bin", "whispercpp", "whisper-cli.exe")
	} else {
		return fmt.Errorf("whisper.cpp does not support your OS: %s, please choose another provider", runtime.GOOS)
	}
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		log.GetLogger().Info("whispercpp not found, starting automatic download (please wait...)")
		err = os.MkdirAll("bin", 0755)
		if err != nil {
			log.GetLogger().Error("Failed to create ./bin directory", zap.Error(err))
			return err
		}
		var downloadUrl string
		if runtime.GOOS == "windows" {
			downloadUrl = "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/whispercpp-windows-cuda.zip"
		}
		zipFilePath := filepath.Join("bin", "whispercpp-windows-cuda.zip")
		err = util.DownloadFile(downloadUrl, zipFilePath, config.Conf.App.Proxy)
		if err != nil {
			log.GetLogger().Error("Failed to download whispercpp", zap.Error(err))
			return err
		}
		log.GetLogger().Info("Starting to unzip whispercpp")
		err = util.Unzip(zipFilePath, filepath.Join("bin", "whispercpp")+string(filepath.Separator))
		if err != nil {
			log.GetLogger().Error("Failed to unzip whispercpp", zap.Error(err))
			return err
		}
	}
	if runtime.GOOS != "windows" {
		err = os.Chmod(filePath, 0755)
		if err != nil {
			log.GetLogger().Error("Failed to set file permissions", zap.Error(err))
			return err
		}
	}
	storage.WhispercppPath = filePath
	log.GetLogger().Info("whispercpp check complete", zap.String("Path", filePath))
	return nil
}

// Detect local model
func checkModel(whisperType string) error {
	var err error
	if _, err = os.Stat("./models/whisperkit"); os.IsNotExist(err) {
		err = os.MkdirAll("./models/whisperkit", 0755)
		if err != nil {
			log.GetLogger().Error("Failed to create ./models directory", zap.Error(err))
			return err
		}
	}
	// Model files
	var model string
	var modelPath string // model path used in cli
	switch whisperType {
	case "fasterwhisper":
		model = config.Conf.Transcribe.Fasterwhisper.Model
		modelPath = fmt.Sprintf("./models/faster-whisper-%s/model.bin", model)
		if _, err = os.Stat(modelPath); os.IsNotExist(err) {
			// Download
			log.GetLogger().Info(fmt.Sprintf("Model file %s not found, starting automatic download...", modelPath))
			downloadUrl := fmt.Sprintf("https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/faster-whisper-%s.zip", model)
			err = util.DownloadFile(downloadUrl, fmt.Sprintf("./models/faster-whisper-%s.zip", model), config.Conf.App.Proxy)
			if err != nil {
				log.GetLogger().Error("Failed to download fasterwhisper model", zap.Error(err))
				return err
			}
			err = util.Unzip(fmt.Sprintf("./models/faster-whisper-%s.zip", model), fmt.Sprintf("./models/faster-whisper-%s/", model))
			if err != nil {
				log.GetLogger().Error("Failed to unzip model", zap.Error(err))
				return err
			}
			log.GetLogger().Info("Model download complete", zap.String("Path", modelPath))
		}
	//case "whisperx":
	//	// TODO: upload models
	//	model = config.Conf.Transcribe.Whisperx.Model
	//	modelDir := fmt.Sprintf("./models/whisperx/models--Systran--faster-whisper-%s", model)
	//	if _, err = os.Stat(modelDir); os.IsNotExist(err) {
	//		log.GetLogger().Info(fmt.Sprintf("WhisperX model %s not found, starting automatic download", modelDir))
	//		// downloadUrl := fmt.Sprintf("https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/WhisperX_models_%s.zip", model)
	//		// err = util.DownloadFile(downloadUrl, fmt.Sprintf("./models/WhisperX_models_%s.zip", model), config.Conf.App.Proxy)
	//		// if err != nil {
	//		// 	log.GetLogger().Info("Failed to download WhisperX model", zap.Error(err))
	//		// 	return err
	//		// }
	//		err = util.Unzip(fmt.Sprintf("./models/WhisperX_models_%s.zip", model), "./models/whisperx/")
	//		if err != nil {
	//			log.GetLogger().Error("Failed to unzip model", zap.Error(err))
	//			return err
	//		}
	//		log.GetLogger().Info("WhisperX model download complete", zap.String("Path", modelDir))
	//	}
	case "whispercpp":
		model = config.Conf.Transcribe.Whispercpp.Model
		modelPath = fmt.Sprintf("./models/whispercpp/ggml-%s.bin", model)
		if _, err = os.Stat(modelPath); os.IsNotExist(err) {
			log.GetLogger().Info(fmt.Sprintf("whisper.cpp model %s not found, starting automatic download...", modelPath))
			downloadUrl := fmt.Sprintf("https://gitcode.com/hf_mirrors/ai-gitcode/whisper.cpp/blob/main/ggml-%s.bin", model)
			err = util.DownloadFile(downloadUrl, fmt.Sprintf("./models/whispercpp/ggml-%s.bin", model), config.Conf.App.Proxy)
			if err != nil {
				log.GetLogger().Info("Failed to download whisper.cpp model", zap.Error(err))
				return err
			}
			log.GetLogger().Info("whisper.cpp model download complete", zap.String("Path", modelPath))
		}
	case "whisperkit":
		model = config.Conf.Transcribe.Whisperkit.Model
		modelPath = fmt.Sprintf("./models/whisperkit/openai_whisper-%s", model)
		files, _ := os.ReadDir(modelPath)
		if len(files) == 0 {
			log.GetLogger().Info("whisperkit model not found, starting automatic download...")
			downloadUrl := "https://modelscope.cn/models/Maranello/MarkFlow Studio_dependency_cn/resolve/master/whisperkit-large-v2.zip"
			err = util.DownloadFile(downloadUrl, "./models/whisperkit/openai_whisper-large-v2.zip", config.Conf.App.Proxy)
			if err != nil {
				log.GetLogger().Info("Failed to download whisperkit model", zap.Error(err))
				return err
			}
			err = util.Unzip("./models/whisperkit/openai_whisper-large-v2.zip", "./models/whisperkit/")
			if err != nil {
				log.GetLogger().Error("Failed to unzip whisperkit model", zap.Error(err))
				return err
			}
			log.GetLogger().Info("whisperkit model download complete", zap.String("Path", modelPath))
		}
	}

	log.GetLogger().Info("Model check complete", zap.String("Path", modelPath))
	return nil
}
