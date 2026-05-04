package service

import (
	"fmt"
	"krillin-ai/config"
	"krillin-ai/log"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

func Test_isValidSplitContent(t *testing.T) {
	// Fixed test file path
	splitContentFile := "g:\\bin\\AI\\tasks\\gdQRrtQP\\srt_no_ts_1.srt"
	originalTextFile := "g:\\bin\\AI\\tasks\\gdQRrtQP\\output\\origin_1.txt"

	// Read split content file
	splitContent, err := os.ReadFile(splitContentFile)
	if err != nil {
		t.Fatalf("Failed to read split content file: %v", err)
	}

	// Read original text file
	originalText, err := os.ReadFile(originalTextFile)
	if err != nil {
		t.Fatalf("Failed to read original text file: %v", err)
	}

	// Execute test
	if _, err := parseAndCheckContent(string(splitContent), string(originalText)); err != nil {
		t.Errorf("parseAndCheckContent() error = %v, want nil", err)
	}
}

func loadTestConfig() bool {
	var err error
	configPath := "../../config/config.toml"
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		log.GetLogger().Info("Config file not found")
		return false
	} else {
		log.GetLogger().Info("Config file found, loading configuration from file")
		if _, err = toml.DecodeFile(configPath, &config.Conf); err != nil {
			log.GetLogger().Error("Failed to load config file", zap.Error(err))
			return false
		}
		return true
	}
}

func initService() *Service {
	log.InitLogger()
	loadTestConfig()
	return NewService()
}

func Test_splitOriginLongSentence(t *testing.T) {
	// Fixed test file path
	testText := "then one more thing is search for file count file explorer note count is the name of the plug in install it and once enabled you can see that now I can see how many files are in each are inside each individual folder even the nested folders are showing properly now how many files are in them"
	s := initService()
	// Execute test
	splitTextSentences, err := s.splitOriginLongSentence(testText)
	if err != nil {
		t.Errorf("splitOriginLongSentence() error = %v, want nil", err)
	}

	fmt.Println("testText:", testText)
	for i, sentence := range splitTextSentences {
		fmt.Printf("Sentence %d: %s\n", i+1, sentence)
	}
}
