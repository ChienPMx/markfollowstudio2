package service

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"markflow-studio/internal/types"
	"markflow-studio/log"
	"markflow-studio/pkg/util"
)

func (s Service) uploadSubtitles(ctx context.Context, stepParam *types.SubtitleTaskStepParam) error {
	subtitleInfos := make([]types.SubtitleInfo, 0)
	var err error
	for _, info := range stepParam.SubtitleInfos {
		resultPath := info.Path
		if len(stepParam.ReplaceWordsMap) > 0 { // Words replacement required
			replacedSrcFile := util.AddSuffixToFileName(resultPath, "_replaced")
			err = util.ReplaceFileContent(resultPath, replacedSrcFile, stepParam.ReplaceWordsMap)
			if err != nil {
				log.GetLogger().Error("uploadSubtitles ReplaceFileContent err", zap.Any("stepParam", stepParam), zap.Error(err))
				return fmt.Errorf("uploadSubtitles ReplaceFileContent err: %w", err)
			}
			resultPath = replacedSrcFile
		}
		subtitleInfos = append(subtitleInfos, types.SubtitleInfo{
			TaskId:      stepParam.TaskId,
			Name:        info.Name,
			DownloadUrl: "/api/file/" + resultPath,
		})
	}
	// Update subtitle task info
	taskPtr := stepParam.TaskPtr
	taskPtr.SubtitleInfos = subtitleInfos
	taskPtr.Status = types.SubtitleTaskStatusSuccess
	taskPtr.ProcessPct = 100
	// Voiceover file
	if stepParam.TtsResultFilePath != "" {
		taskPtr.SpeechDownloadUrl = "/api/file/" + stepParam.TtsResultFilePath
	}
	return nil
}
