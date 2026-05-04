package util

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"krillin-ai/config"
	"krillin-ai/log"
	"net/http"
	"os"
	"time"
)

// progressWriter displays download progress, implements io.Writer
type progressWriter struct {
	Total      uint64
	Downloaded uint64
	StartTime  time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Downloaded += uint64(n)

	// Initialize start time
	if pw.StartTime.IsZero() {
		pw.StartTime = time.Now()
	}

	percent := float64(pw.Downloaded) / float64(pw.Total) * 100
	elapsed := time.Since(pw.StartTime).Seconds()
	speed := float64(pw.Downloaded) / 1024 / 1024 / elapsed

	fmt.Printf("\rDownload Progress: %.2f%% (%.2f MB / %.2f MB) | Speed: %.2f MB/s",
		percent,
		float64(pw.Downloaded)/1024/1024,
		float64(pw.Total)/1024/1024,
		speed)

	return n, nil
}

// DownloadFile downloads a file and saves it to a specified path, supports proxy
func DownloadFile(urlStr, filepath, proxyAddr string) error {
	log.GetLogger().Info("Start downloading file", zap.String("url", urlStr))
	client := &http.Client{}
	if proxyAddr != "" {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(config.Conf.App.ParsedProxy),
		}
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size := resp.ContentLength
	fmt.Printf("File Size: %.2f MB\n", float64(size)/1024/1024)

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Reader with progress tracking
	progress := &progressWriter{
		Total: uint64(size),
	}
	reader := io.TeeReader(resp.Body, progress)

	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}
	fmt.Printf("\n") // End of progress info, newline

	log.GetLogger().Info("File download complete", zap.String("Path", filepath))
	return nil
}
