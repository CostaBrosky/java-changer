package java

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"jv/internal/config"
)

// Detector finds Java installations on the system
type Detector struct {
	standardPaths []string
}

// NewDetector creates a new Java detector
func NewDetector() *Detector {
	return &Detector{
		standardPaths: []string{
			"C:\\Program Files\\Java",
			"C:\\Program Files (x86)\\Java",
			"C:\\Program Files\\Eclipse Adoptium",
			"C:\\Program Files\\Eclipse Foundation",
			"C:\\Program Files\\Zulu",
			"C:\\Program Files\\Amazon Corretto",
			"C:\\Program Files\\Microsoft",
		},
	}
}

// FindAll finds all Java installations (auto-detected + custom)
func (d *Detector) FindAll() ([]Version, error) {
	versions := make([]Version, 0)

	// Load config first to get additional search paths
	cfg, err := config.Load()
	searchPaths := d.standardPaths

	// Add custom search paths from config
	if err == nil && len(cfg.SearchPaths) > 0 {
		searchPaths = append(searchPaths, cfg.SearchPaths...)
	}

	// Use a map to deduplicate by path (case-insensitive)
	type item struct{ v Version }
	seen := make(map[string]item)

	// Auto-detect from all search paths (standard + custom)
	for _, basePath := range searchPaths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			javaPath := filepath.Join(basePath, entry.Name())
			if d.IsValidJavaPath(javaPath) {
				version := d.GetVersion(javaPath)
				key := strings.ToLower(filepath.Clean(javaPath))
				seen[key] = item{v: Version{Version: version, Path: filepath.Clean(javaPath), IsCustom: false}}
			}
		}
	}

	// Add specific custom installation paths
	if err == nil {
		for _, customPath := range cfg.CustomPaths {
			if d.IsValidJavaPath(customPath) {
				norm := filepath.Clean(customPath)
				key := strings.ToLower(norm)
				version := d.GetVersion(norm)
				// If already seen as auto, upgrade to custom; else add as custom
				seen[key] = item{v: Version{Version: version, Path: norm, IsCustom: true}}
			}
		}
	}

	// Materialize map to slice
	for _, it := range seen {
		versions = append(versions, it.v)
	}

	return versions, nil
}

// IsValidJavaPath checks if a path is a valid Java installation
func (d *Detector) IsValidJavaPath(path string) bool {
	javaExe := filepath.Join(path, "bin", "java.exe")
	_, err := os.Stat(javaExe)
	return err == nil
}

// IsValidSearchPath checks if a path is a valid directory to search for Java installations
func (d *Detector) IsValidSearchPath(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetVersion extracts the version from a Java installation path
func (d *Detector) GetVersion(javaPath string) string {
	// First, try to get version by running java -version
	javaExe := filepath.Join(javaPath, "bin", "java.exe")
	cmd := exec.Command(javaExe, "-version")
	output, err := cmd.CombinedOutput()
	if err == nil {
		version := d.parseVersionOutput(string(output))
		if version != "" {
			return version
		}
	}

	// Fallback: extract from directory name
	dirName := filepath.Base(javaPath)
	return d.parseVersionFromDirName(dirName)
}

// parseVersionOutput parses the output of 'java -version'
func (d *Detector) parseVersionOutput(output string) string {
	// Look for version patterns like: version "17.0.1"
	re := regexp.MustCompile(`version\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	// Alternative pattern: openjdk version "11.0.12"
	re = regexp.MustCompile(`(?:openjdk|java)\s+version\s+"([^"]+)"`)
	matches = re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// parseVersionFromDirName extracts version from directory names like "jdk-17" or "jdk1.8.0_322"
func (d *Detector) parseVersionFromDirName(dirName string) string {
	dirName = strings.ToLower(dirName)

	// Pattern: jdk-17, jdk-17.0.1, etc.
	re := regexp.MustCompile(`jdk-?(\d+(?:\.\d+)*(?:_\d+)?)`)
	matches := re.FindStringSubmatch(dirName)
	if len(matches) > 1 {
		return matches[1]
	}

	// Pattern: jdk1.8.0_322
	re = regexp.MustCompile(`jdk(1\.\d+\.\d+_\d+)`)
	matches = re.FindStringSubmatch(dirName)
	if len(matches) > 1 {
		return matches[1]
	}

	// Pattern: java-17, java-11, etc.
	re = regexp.MustCompile(`java-?(\d+(?:\.\d+)*)`)
	matches = re.FindStringSubmatch(dirName)
	if len(matches) > 1 {
		return matches[1]
	}

	// Return dir name as-is if no pattern matches
	return dirName
}
