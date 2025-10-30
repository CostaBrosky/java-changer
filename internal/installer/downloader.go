package installer

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// DownloadFile downloads a file from URL with animated progress bar
func DownloadFile(url string, destPath string) error {
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength

	// Create progress model
	progressModel := NewProgressModel(totalSize)
	p := tea.NewProgram(progressModel)

	// Create progress writer
	pw := newProgressWriter(totalSize, p)

	// Start the progress UI in a goroutine
	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running progress: %v\n", err)
		}
	}()

	// Give the UI a moment to start
	time.Sleep(100 * time.Millisecond)

	// Create multi-writer: write to file AND progress tracker
	multiWriter := io.MultiWriter(out, pw)

	// Download with progress
	written, err := io.Copy(multiWriter, resp.Body)
	if err != nil {
		p.Send(progressErrMsg{err: err})
		p.Quit()
		return fmt.Errorf("failed to write file: %w", err)
	}

	if written != totalSize {
		err := fmt.Errorf("incomplete download: got %d bytes, expected %d", written, totalSize)
		p.Send(progressErrMsg{err: err})
		p.Quit()
		return err
	}

	// Signal completion
	p.Send(downloadCompleteMsg{})

	// Wait a moment for UI to finish
	time.Sleep(200 * time.Millisecond)

	return nil
}

// VerifyChecksum verifies the SHA256 checksum of a file
func VerifyChecksum(filePath string, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// ExtractZip extracts a ZIP file to the destination directory
func ExtractZip(zipPath string, destDir string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	// Find the root directory in the ZIP (usually jdk-xxx)
	var rootDir string
	for _, file := range reader.File {
		if file.FileInfo().IsDir() && strings.HasPrefix(file.Name, "jdk") {
			parts := strings.Split(file.Name, "/")
			if len(parts) > 0 {
				rootDir = parts[0]
				break
			}
		}
	}

	extractedPath := ""

	for _, file := range reader.File {
		filePath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", fmt.Errorf("failed to create file: %w", err)
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return "", fmt.Errorf("failed to open file in zip: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", fmt.Errorf("failed to extract file: %w", err)
		}
	}

	if rootDir != "" {
		extractedPath = filepath.Join(destDir, rootDir)
	}

	return extractedPath, nil
}

// InstallJDK orchestrates the download, verification, and extraction of a JDK
func InstallJDK(downloadInfo *DownloadInfo, version string, distributor string, isSystemWide bool) (string, error) {
	// Determine installation base directory
	var installBase string
	if isSystemWide {
		// Use absolute path for system-wide installation
		installBase = filepath.Join(`C:\Program Files`, distributor)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		installBase = filepath.Join(homeDir, ".jv")
	}

	// Create installation directory
	if err := os.MkdirAll(installBase, 0755); err != nil {
		return "", fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Create temp directory for download
	tempDir := filepath.Join(os.TempDir(), "jv-install")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download JDK
	zipPath := filepath.Join(tempDir, downloadInfo.FileName)
	fmt.Println("Downloading JDK...")
	if err := DownloadFile(downloadInfo.URL, zipPath); err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}

	// Verify checksum with spinner
	var checksumErr error
	spinnerErr := WithSpinner("Verifying checksum...", func() error {
		checksumErr = VerifyChecksum(zipPath, downloadInfo.Checksum)
		return nil
	})
	if spinnerErr != nil {
		return "", spinnerErr
	}
	if checksumErr != nil {
		return "", fmt.Errorf("checksum verification failed: %w", checksumErr)
	}
	fmt.Println("✓ Checksum verified successfully")

	// Extract to temp location with spinner
	var extractedPath string
	var extractErr error
	tempExtractDir := filepath.Join(tempDir, "extract")

	spinnerErr = WithSpinner("Extracting JDK...", func() error {
		var err error
		extractedPath, err = ExtractZip(zipPath, tempExtractDir)
		extractErr = err
		return nil
	})
	if spinnerErr != nil {
		return "", spinnerErr
	}
	if extractErr != nil {
		return "", fmt.Errorf("extraction failed: %w", extractErr)
	}
	fmt.Println("✓ JDK extracted successfully")

	// Verify java.exe exists
	javaExe := filepath.Join(extractedPath, "bin", "java.exe")
	if _, err := os.Stat(javaExe); os.IsNotExist(err) {
		return "", fmt.Errorf("invalid JDK structure: bin\\java.exe not found")
	}

	// Move to final location
	finalPath := filepath.Join(installBase, fmt.Sprintf("jdk-%s", version))

	// Remove old installation if exists
	if _, err := os.Stat(finalPath); err == nil {
		fmt.Printf("Removing existing installation at %s\n", finalPath)
		if err := os.RemoveAll(finalPath); err != nil {
			return "", fmt.Errorf("failed to remove old installation: %w", err)
		}
	}

	// Move extracted directory to final location
	if err := os.Rename(extractedPath, finalPath); err != nil {
		return "", fmt.Errorf("failed to move JDK to final location: %w", err)
	}

	fmt.Printf("JDK installed successfully to: %s\n", finalPath)
	return finalPath, nil
}
