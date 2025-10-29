package installer

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"jv/internal/config"
	"jv/internal/env"
	"jv/internal/java"
)

// Installer handles the interactive Java installation process
type Installer struct {
	detector     *java.Detector
	config       *config.Config
	isAdmin      bool
	distributors map[int]Distributor
}

// NewInstaller creates a new Installer instance
func NewInstaller(isAdmin bool) (*Installer, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	distributors := make(map[int]Distributor)
	distributors[1] = NewAdoptiumDistributor()
	// Future: distributors[2] = NewAzulDistributor()
	// Future: distributors[3] = NewCorrettoDistributor()

	return &Installer{
		detector:     java.NewDetector(),
		config:       cfg,
		isAdmin:      isAdmin,
		distributors: distributors,
	}, nil
}

// Run starts the interactive installation process
func (i *Installer) Run() error {
	fmt.Println()
	fmt.Println("=====================================")
	fmt.Println("   Java Installation Manager")
	fmt.Println("=====================================")
	fmt.Println()

	if !i.isAdmin {
		fmt.Println("⚠  Not running as Administrator")
		fmt.Println("   Installation will be user-level only")
		fmt.Println("   JAVA_HOME cannot be set automatically")
		fmt.Println()
	}

	// Step 1: Select distributor
	distributor, err := i.ShowDistributorMenu()
	if err != nil {
		return err
	}

	// Step 2: Select version
	version, err := i.ShowVersionMenu(distributor)
	if err != nil {
		return err
	}

	// Step 3: Download and install
	installedPath, err := i.InstallVersion(distributor, version)
	if err != nil {
		return err
	}

	// Step 4: Add to config
	i.config.AddCustomPath(installedPath)

	// Track installed JDK
	installedJDK := config.InstalledJDK{
		Version:     version,
		Path:        installedPath,
		Distributor: distributor.Name(),
		InstalledAt: time.Now().Format(time.RFC3339),
	}
	i.config.AddInstalledJDK(installedJDK)

	if err := i.config.Save(); err != nil {
		fmt.Printf("Warning: Failed to save config: %v\n", err)
	}

	// Step 5: Configure environment if JAVA_HOME not set
	if err := i.ConfigureEnvironment(installedPath); err != nil {
		fmt.Printf("\nNote: %v\n", err)
	}

	fmt.Println()
	fmt.Println("=====================================")
	fmt.Println("   Installation Complete!")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Printf("Java %s installed to:\n", version)
	fmt.Printf("  %s\n", installedPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run: jv list")
	fmt.Println("  2. Run: jv use " + version)
	fmt.Println()

	return nil
}

// ShowDistributorMenu displays available distributors and returns the selected one
func (i *Installer) ShowDistributorMenu() (Distributor, error) {
	fmt.Println("Available Java Distributors:")
	fmt.Println()
	fmt.Println("  1. Eclipse Adoptium (Temurin)  [Active]")
	fmt.Println("  2. Azul Zulu                    [Coming Soon]")
	fmt.Println("  3. Amazon Corretto              [Coming Soon]")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Select distributor [1]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			input = "1"
		}

		selection, err := strconv.Atoi(input)
		if err != nil || selection < 1 || selection > 3 {
			fmt.Println("Invalid selection. Please enter 1-3")
			continue
		}

		if selection != 1 {
			fmt.Println("This distributor is not yet supported. Please select 1 (Adoptium)")
			continue
		}

		return i.distributors[selection], nil
	}
}

// ShowVersionMenu displays available versions and returns the selected one
func (i *Installer) ShowVersionMenu(distributor Distributor) (string, error) {
	fmt.Println()
	fmt.Printf("Fetching available versions from %s...\n", distributor.Name())

	releases, err := distributor.GetAvailableVersions()
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	// Get currently installed versions
	installedVersions, err := i.detector.FindAll()
	if err != nil {
		fmt.Printf("Warning: Failed to detect installed versions: %v\n", err)
		installedVersions = []java.Version{}
	}

	// Create map of installed versions for quick lookup
	installedMap := make(map[string]bool)
	for _, iv := range installedVersions {
		// Extract major version
		parts := strings.Split(iv.Version, ".")
		if len(parts) > 0 {
			installedMap[parts[0]] = true
		}
	}

	fmt.Println()
	fmt.Println("Available Java Versions:")
	fmt.Println()

	// Show LTS versions first
	fmt.Println("Long Term Support (LTS):")
	ltsCount := 0
	for _, release := range releases {
		if !release.IsLTS {
			continue
		}
		ltsCount++
		marker := ""
		if installedMap[release.Version] {
			marker = " [Installed]"
		}
		fmt.Printf("  %2s. Java %s%s\n", release.Version, release.Version, marker)
	}

	// Show feature versions
	fmt.Println()
	fmt.Println("Feature Releases:")
	for _, release := range releases {
		if release.IsLTS {
			continue
		}
		marker := ""
		if installedMap[release.Version] {
			marker = " [Installed]"
		}
		fmt.Printf("  %2s. Java %s%s\n", release.Version, release.Version, marker)
	}

	fmt.Println()
	fmt.Println("Note: LTS versions are recommended for production use")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Select version to install (or 'q' to quit): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" || input == "Q" {
			return "", fmt.Errorf("installation cancelled")
		}

		// Validate version exists
		found := false
		for _, release := range releases {
			if release.Version == input {
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Invalid version. Please select a valid version or 'q' to quit\n")
			continue
		}

		// Warn if already installed
		if installedMap[input] {
			fmt.Printf("\nJava %s appears to be already installed.\n", input)
			fmt.Print("Continue anyway? (y/N): ")
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))
			if confirm != "y" && confirm != "yes" {
				continue
			}
		}

		return input, nil
	}
}

// InstallVersion downloads and installs the selected version
func (i *Installer) InstallVersion(distributor Distributor, version string) (string, error) {
	fmt.Println()
	fmt.Printf("Installing Java %s from %s...\n", version, distributor.Name())
	fmt.Println()

	// Get system architecture
	arch := runtime.GOARCH

	// Get download URL
	fmt.Println("Fetching download information...")
	downloadInfo, err := distributor.GetDownloadURL(version, arch)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	fmt.Printf("Package: %s\n", downloadInfo.FileName)
	fmt.Printf("Size: %.2f MB\n", float64(downloadInfo.Size)/1024/1024)
	fmt.Println()

	// Install JDK
	installedPath, err := InstallJDK(downloadInfo, version, distributor.Name(), i.isAdmin)
	if err != nil {
		return "", fmt.Errorf("installation failed: %w", err)
	}

	return installedPath, nil
}

// ConfigureEnvironment sets JAVA_HOME if not already set
func (i *Installer) ConfigureEnvironment(jdkPath string) error {
	// Check if JAVA_HOME is already set
	currentJavaHome := os.Getenv("JAVA_HOME")
	if currentJavaHome != "" {
		fmt.Println()
		fmt.Println("JAVA_HOME is already set to:")
		fmt.Printf("  %s\n", currentJavaHome)
		fmt.Println()
		fmt.Printf("To use the newly installed Java, run:\n")
		fmt.Printf("  jv use <version>\n")
		return nil
	}

	// Need admin privileges to set system environment variables
	if !i.isAdmin {
		fmt.Println()
		fmt.Println("⚠  Cannot set JAVA_HOME automatically (requires administrator)")
		fmt.Println()
		fmt.Println("To configure Java, run as administrator:")
		fmt.Printf("  jv init\n")
		fmt.Println()
		fmt.Println("Or manually set JAVA_HOME to:")
		fmt.Printf("  %s\n", jdkPath)
		return nil
	}

	// Set JAVA_HOME
	fmt.Println()
	fmt.Println("Configuring JAVA_HOME...")
	if err := env.SetJavaHome(jdkPath); err != nil {
		return fmt.Errorf("failed to set JAVA_HOME: %w", err)
	}

	fmt.Println("✓ JAVA_HOME configured successfully")
	fmt.Printf("  JAVA_HOME = %s\n", jdkPath)
	fmt.Println("  Added %JAVA_HOME%\\bin to PATH")

	return nil
}
