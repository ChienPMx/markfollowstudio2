package desktop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

type FileManager struct {
	window        fyne.Window
	files         []string
	selectedFiles []string // Multiple file selection
}

func NewFileManager(window fyne.Window) *FileManager {
	return &FileManager{
		window:        window,
		files:         make([]string, 0),
		selectedFiles: make([]string, 0),
	}
}

func (fm *FileManager) ShowUploadDialog() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, fm.window)
			return
		}
		if reader == nil {
			return
		}

		// Get file path
		filePath := reader.URI().Path()
		fileName := filepath.Base(filePath)

		err = fm.uploadFile(filePath, fileName)
		if err != nil {
			dialog.ShowError(err, fm.window)
			return
		}

		dialog.ShowInformation("Success", "File uploaded successfully", fm.window)
	}, fm.window)

	fd.Show()
}

func (fm *FileManager) ShowMultiUploadDialog() {
	// Clear selected files
	fm.selectedFiles = make([]string, 0)

	fm.showAddFileDialog()
}

func (fm *FileManager) showAddFileDialog() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, fm.window)
			return
		}
		if reader == nil {
			return
		}

		// Get file path
		filePath := reader.URI().Path()
		fm.selectedFiles = append(fm.selectedFiles, filePath)
		reader.Close()

		// Continue adding files?
		continueDialog := dialog.NewConfirm(
			"Add more files",
			fmt.Sprintf("%d files selected. Continue adding?", len(fm.selectedFiles)),
			func(cont bool) {
				if cont {
					// Continue adding files
					fm.showAddFileDialog()
				} else {
					// Upload
					if len(fm.selectedFiles) > 0 {
						// File name list
						fileNames := make([]string, len(fm.selectedFiles))
						for i, path := range fm.selectedFiles {
							fileNames[i] = filepath.Base(path)
						}

						// Upload
						err := fm.uploadMultipleFiles(fm.selectedFiles, fileNames)
						if err != nil {
							dialog.ShowError(err, fm.window)
							return
						}

						dialog.ShowInformation("Success", fmt.Sprintf("%d files uploaded", len(fm.selectedFiles)), fm.window)
					}
				}
			},
			fm.window,
		)
		continueDialog.Show()
	}, fm.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".mp4", ".mov", ".avi", ".mkv", ".wmv"}))
	fd.Show()
}

func (fm *FileManager) uploadFile(filePath, fileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	writer.Close()

	// Send request
	resp, err := http.Post("http://localhost:8888/api/file", writer.FormDataContentType(), body)
	if err != nil {
		return err
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
		return err
	}

	if result.Error != 0 && result.Error != 200 {
		return fmt.Errorf(result.Msg)
	}

	fm.files = append(fm.files, result.Data.FilePath)
	return nil
}

func (fm *FileManager) uploadMultipleFiles(filePaths []string, fileNames []string) error {
	// Show upload progress dialog
	progress := dialog.NewProgress("Uploading", "Uploading files...", fm.window)
	progress.Show()
	defer progress.Hide()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for i, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}

		part, err := writer.CreateFormFile("file", fileNames[i])
		if err != nil {
			file.Close()
			return err
		}

		_, err = io.Copy(part, file)
		file.Close()
		if err != nil {
			return err
		}

		progress.SetValue(float64(i+1) / float64(len(filePaths)))
	}
	writer.Close()

	resp, err := http.Post("http://localhost:8888/api/file", writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  struct {
			FilePath []string `json:"file_path"`
		} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}

	if result.Error != 0 && result.Error != 200 {
		return fmt.Errorf(result.Msg)
	}

	fm.files = append(fm.files, result.Data.FilePath...)
	return nil
}

func (fm *FileManager) GetFileCount() int {
	return len(fm.files)
}

func (fm *FileManager) GetFileName(index int) string {
	if index < 0 || index >= len(fm.files) {
		return ""
	}
	return filepath.Base(fm.files[index])
}

func (fm *FileManager) DownloadFile(index int) {
	if index < 0 || index >= len(fm.files) {
		return
	}

	filePath := fm.files[index]

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, fm.window)
			return
		}
		if writer == nil {
			return
		}

		resp, err := http.Get("http://localhost:8888" + filePath)
		if err != nil {
			dialog.ShowError(err, fm.window)
			return
		}
		defer resp.Body.Close()

		_, err = io.Copy(writer, resp.Body)
		if err != nil {
			dialog.ShowError(err, fm.window)
			return
		}

		writer.Close()
		dialog.ShowInformation("Success", "File download complete", fm.window)
	}, fm.window)
}
