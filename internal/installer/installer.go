package installer

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"jv/internal/config"
	"jv/internal/env"
	"jv/internal/java"
	"jv/internal/theme"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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
	// Styled header with JV theme
	title := theme.Title.Padding(0, 2).Render("Java Installation Manager")
	fmt.Println()
	fmt.Println(theme.TitleBox.Render(title))
	fmt.Println()

	if !i.isAdmin {
		fmt.Println(theme.WarningMessage("Not running as Administrator"))
		fmt.Println(theme.Faint.Render("   Installation will be user-level only"))
		fmt.Println(theme.Faint.Render("   JAVA_HOME cannot be set automatically"))
		fmt.Println()
	}

	// Step 1: Select distributor
	distributor, err := i.ShowDistributorMenu()
	if err != nil {
		return err
	}

	// Step 1.5: Select installation mode
	mode, err := i.SelectInstallMode()
	if err != nil {
		return err
	}

	if mode == "multi" {
		return i.RunMultiInstall(distributor)
	}

	// Single install (existing flow)
	return i.RunSingleInstall(distributor)
}

// RunSingleInstall handles single version installation
func (i *Installer) RunSingleInstall(distributor Distributor) error {
	// Step 2: Select version
	version, err := i.ShowVersionMenu(distributor)
	if err != nil {
		return err
	}

	// Step 3: Select scope
	scope, err := i.SelectInstallScope()
	if err != nil {
		return err
	}

	// Step 4: Install
	installedPath, err := i.InstallVersion(distributor, version, scope)
	if err != nil {
		return err
	}

	// Step 5: Configure and save
	return i.finalizeInstallation([]string{installedPath}, []string{version}, scope, distributor.Name())
}

// RunMultiInstall handles multiple versions installation
func (i *Installer) RunMultiInstall(distributor Distributor) error {
	// Step 2: Select multiple versions
	versions, err := i.SelectMultipleVersions(distributor)
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		return fmt.Errorf("no versions selected")
	}

	// Step 3: Select scope (same for all)
	scope, err := i.SelectInstallScope()
	if err != nil {
		return err
	}

	// Step 4: Install each version
	fmt.Println()
	fmt.Printf("Installing %d Java versions...\n", len(versions))
	fmt.Println()

	installedPaths := []string{}
	for idx, version := range versions {
		fmt.Printf("[%d/%d] Installing Java %s...\n", idx+1, len(versions), version)

		installedPath, err := i.InstallVersion(distributor, version, scope)
		if err != nil {
			fmt.Printf("❌ Failed to install Java %s: %v\n", version, err)
			continue
		}

		installedPaths = append(installedPaths, installedPath)
		fmt.Printf("✓ Java %s installed successfully\n\n", version)
	}

	// Step 5: Configure and save
	return i.finalizeInstallation(installedPaths, versions, scope, distributor.Name())
}

// finalizeInstallation handles config saving and environment setup
func (i *Installer) finalizeInstallation(paths []string, versions []string, scope string, distributorName string) error {
	// Add to config
	for idx, path := range paths {
		if strings.EqualFold(scope, "user") {
			i.config.AddCustomPath(path)
		}

		installedJDK := config.InstalledJDK{
			Version:     versions[idx],
			Path:        path,
			Distributor: distributorName,
			InstalledAt: time.Now().Format(time.RFC3339),
			Scope:       scope,
		}
		i.config.AddInstalledJDK(installedJDK)
	}

	if err := i.config.Save(); err != nil {
		fmt.Printf("Warning: Failed to save config: %v\n", err)
	}

	// Configure environment for first installation if JAVA_HOME not set
	if len(paths) > 0 {
		if err := i.ConfigureEnvironment(paths[0]); err != nil {
			fmt.Printf("\nNote: %v\n", err)
		}
	}

	// Success message with JV theme
	fmt.Println()
	title := theme.SuccessStyle.Padding(0, 2).Render("✓ Installation Complete!")
	fmt.Println(theme.SuccessBox.Render(title))
	fmt.Println()

	// Installation details
	if len(paths) == 1 {
		fmt.Println(theme.LabelStyle.Render(fmt.Sprintf("Java %s installed to:", versions[0])))
		fmt.Printf("  %s\n", theme.PathStyle.Render(paths[0]))
	} else {
		fmt.Println(theme.LabelStyle.Render(fmt.Sprintf("Installed %d Java versions:", len(paths))))
		for idx, version := range versions {
			fmt.Printf("  • %s → %s\n",
				theme.SuccessStyle.Render("Java "+version),
				theme.PathStyle.Render(paths[idx]))
		}
	}

	fmt.Println()

	// Next steps
	fmt.Println(theme.Subtitle.Render("Next steps:"))
	fmt.Printf("  %s %s\n", theme.StepStyle.Render("1."), theme.Code.Render("jv list"))
	fmt.Printf("  %s %s\n", theme.StepStyle.Render("2."), theme.Code.Render("jv use <version>"))
	fmt.Println()

	return nil
}

// SelectInstallScope asks user to choose installation scope (admin only)
func (i *Installer) SelectInstallScope() (string, error) {
	if !i.isAdmin {
		// No choice for non-admin users
		fmt.Println()
		fmt.Println(theme.InfoMessage("Installing to user directory (administrator required for system-wide)"))
		return "user", nil
	}

	var scope string

	err := huh.NewSelect[string]().
		Title(theme.Subtitle.Render("Select Installation Scope")).
		Description(theme.Faint.Render("System-wide requires admin privileges")).
		Options(
			huh.NewOption(theme.CurrentStyle.Render("System-wide")+" (recommended) - C\\Program Files\\...", "system"),
			huh.NewOption(theme.CurrentStyle.Render("User-only")+" - %USERPROFILE%\\.jv\\...", "user"),
		).
		Value(&scope).
		Run()

	if err != nil {
		return "", err
	}

	return scope, nil
}

// ShowDistributorMenu displays available distributors and returns the selected one
func (i *Installer) ShowDistributorMenu() (Distributor, error) {
	var selection string

	err := huh.NewSelect[string]().
		Title(theme.Subtitle.Render("Select Java Distributor")).
		Description(theme.Faint.Render("More distributors coming soon")).
		Options(
			huh.NewOption(theme.CurrentStyle.Render("Eclipse Adoptium")+" (Temurin)", "adoptium"),
			// Coming soon: Azul Zulu, Amazon Corretto
		).
		Value(&selection).
		Run()

	if err != nil {
		return nil, err
	}

	// Return the distributor based on selection
	return i.distributors[1], nil // Adoptium for now
}

// ShowVersionMenu displays available versions and returns the selected one
func (i *Installer) ShowVersionMenu(distributor Distributor) (string, error) {
	var releases []JavaRelease
	var fetchErr error

	// Fetch with spinner
	spinnerErr := WithSpinner(
		fmt.Sprintf("Fetching available versions from %s...", distributor.Name()),
		func() error {
			var err error
			releases, err = distributor.GetAvailableVersions()
			fetchErr = err
			return nil // Don't propagate error, just store it
		},
	)

	if spinnerErr != nil {
		return "", spinnerErr
	}

	if fetchErr != nil {
		fmt.Printf("Warning: %v\n", fetchErr)
	}

	// Get currently installed versions
	installedVersions, _ := i.detector.FindAll()

	// Create map of installed versions for quick lookup
	installedMap := make(map[string]bool)
	for _, iv := range installedVersions {
		// Extract major version
		parts := strings.Split(iv.Version, ".")
		if len(parts) > 0 {
			installedMap[parts[0]] = true
		}
	}

	// Build options grouped by LTS/Feature with themed tags and aligned columns
	var ltsOptions []huh.Option[string]
	var featureOptions []huh.Option[string]

	// determine max width for version column (visual)
	maxW := 0
	for _, r := range releases {
		w := lipgloss.Width("Java " + r.Version)
		if w > maxW {
			maxW = w
		}
	}

	for _, release := range releases {
		base := fmt.Sprintf("Java %s", release.Version)
		vis := lipgloss.Width(base)
		pad := ""
		if vis < maxW {
			pad = strings.Repeat(" ", maxW-vis)
		}

		// Fixed tag columns: [LTS] and [Installed]
		ltsCol := strings.Repeat(" ", len("[LTS]"))
		if release.IsLTS {
			ltsCol = theme.SuccessStyle.Render("[LTS]")
		}
		instCol := strings.Repeat(" ", len("[Installed]"))
		if installedMap[release.Version] {
			instCol = theme.InfoStyle.Render("[Installed]")
		}

		left := theme.CurrentStyle.Render("Java") + " " + release.Version
		// one space before tags, two spaces between tag columns
		label := left + pad + " " + ltsCol + "  " + instCol
		option := huh.NewOption(label, release.Version)

		if release.IsLTS {
			ltsOptions = append(ltsOptions, option)
		} else {
			featureOptions = append(featureOptions, option)
		}
	}

	// Combine all options
	allOptions := append(ltsOptions, featureOptions...)

	var selected string
	err := huh.NewSelect[string]().
		Title(theme.Subtitle.Render("Select Java Version")).
		Description(theme.Faint.Render("Use arrow keys to navigate, Enter to select")).
		Options(allOptions...).
		Value(&selected).
		Run()

	if err != nil {
		return "", err
	}

	return selected, nil
}

// SelectMultipleVersions allows installing multiple Java versions at once
func (i *Installer) SelectMultipleVersions(distributor Distributor) ([]string, error) {
	var releases []JavaRelease
	var fetchErr error

	// Fetch with spinner
	spinnerErr := WithSpinner(
		fmt.Sprintf("Fetching available versions from %s...", distributor.Name()),
		func() error {
			var err error
			releases, err = distributor.GetAvailableVersions()
			fetchErr = err
			return nil
		},
	)

	if spinnerErr != nil {
		return nil, spinnerErr
	}

	if fetchErr != nil {
		return nil, fetchErr
	}

	// Get installed versions
	installedVersions, _ := i.detector.FindAll()
	installedMap := make(map[string]bool)
	for _, iv := range installedVersions {
		parts := strings.Split(iv.Version, ".")
		if len(parts) > 0 {
			installedMap[parts[0]] = true
		}
	}

	// Build options (aligned, with orange prefix and colored tags)
	var options []huh.Option[string]
	// determine max width
	maxW := 0
	for _, r := range releases {
		w := lipgloss.Width("Java " + r.Version)
		if w > maxW {
			maxW = w
		}
	}
	for _, release := range releases {
		base := fmt.Sprintf("Java %s", release.Version)
		vis := lipgloss.Width(base)
		pad := ""
		if vis < maxW {
			pad = strings.Repeat(" ", maxW-vis)
		}

		// Fixed tag columns
		ltsCol := strings.Repeat(" ", len("[LTS]"))
		if release.IsLTS {
			ltsCol = theme.SuccessStyle.Render("[LTS]")
		}
		instCol := strings.Repeat(" ", len("[Installed]"))
		if installedMap[release.Version] {
			instCol = theme.InfoStyle.Render("[Installed]")
		}

		left := theme.CurrentStyle.Render("Java") + " " + release.Version
		label := left + pad + " " + ltsCol + "  " + instCol
		options = append(options, huh.NewOption(label, release.Version))
	}

	var selected []string

	err := huh.NewMultiSelect[string]().
		Title(theme.Subtitle.Render("Select Java Versions to Install")).
		Description(theme.Faint.Render("Use Space to select, Enter to confirm")).
		Options(options...).
		Value(&selected).
		Limit(10). // Show 10 items at a time
		Run()

	if err != nil {
		return nil, err
	}

	return selected, nil
}

// SelectInstallMode allows choosing between single and multi install
func (i *Installer) SelectInstallMode() (string, error) {
	var mode string

	err := huh.NewSelect[string]().
		Title(theme.Subtitle.Render("Installation Mode")).
		Options(
			huh.NewOption(theme.CurrentStyle.Render("Install")+" single version", "single"),
			huh.NewOption(theme.CurrentStyle.Render("Install")+" multiple versions (batch)", "multi"),
		).
		Value(&mode).
		Run()

	if err != nil {
		return "", err
	}

	return mode, nil
}

// InstallVersion downloads and installs the selected version
func (i *Installer) InstallVersion(distributor Distributor, version string, scope string) (string, error) {
	// Installation header with JV theme
	fmt.Println()
	fmt.Println(theme.Subtitle.Render(fmt.Sprintf("Installing Java %s from %s", version, distributor.Name())))
	fmt.Println()

	// Get system architecture
	arch := runtime.GOARCH

	// Get download URL with spinner
	var downloadInfo *DownloadInfo
	var fetchErr error

	spinnerErr := WithSpinner(
		"Fetching download information...",
		func() error {
			var err error
			downloadInfo, err = distributor.GetDownloadURL(version, arch)
			fetchErr = err
			return nil
		},
	)

	if spinnerErr != nil {
		return "", spinnerErr
	}

	if fetchErr != nil {
		return "", fmt.Errorf("failed to get download URL: %w", fetchErr)
	}

	// Styled package info with JV theme
	fmt.Printf("%s %s\n", theme.LabelStyle.Render("Package:"), theme.ValueStyle.Render(downloadInfo.FileName))
	sizeMB := float64(downloadInfo.Size) / 1024 / 1024
	fmt.Printf("%s %s\n", theme.LabelStyle.Render("Size:   "), theme.ValueStyle.Render(fmt.Sprintf("%.2f MB", sizeMB)))
	fmt.Println()

	// Determine isSystemWide based on scope
	isSystemWide := (scope == "system" && i.isAdmin)

	// Install JDK
	installedPath, err := InstallJDK(downloadInfo, version, distributor.Name(), isSystemWide)
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
		fmt.Println(theme.InfoStyle.Render("JAVA_HOME is already set to:"))
		fmt.Printf("  %s\n", theme.PathStyle.Render(currentJavaHome))
		fmt.Println()
		fmt.Println(theme.Faint.Render("To use the newly installed Java, run:"))
		fmt.Printf("  %s\n", theme.Code.Render("jv use <version>"))
		return nil
	}

	// Need admin privileges to set system environment variables
	if !i.isAdmin {
		fmt.Println()
		fmt.Println(theme.WarningMessage("Cannot set JAVA_HOME automatically (requires administrator)"))
		fmt.Println()
		fmt.Println(theme.InfoStyle.Render("To configure Java, run as administrator:"))
		fmt.Printf("  %s\n", theme.Code.Render("jv use <version>"))
		fmt.Println()
		fmt.Println(theme.Faint.Render("Or manually set JAVA_HOME to:"))
		fmt.Printf("  %s\n", theme.PathStyle.Render(jdkPath))
		return nil
	}

	// Set JAVA_HOME
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("Configuring JAVA_HOME..."))
	if err := env.SetJavaHome(jdkPath); err != nil {
		return fmt.Errorf("failed to set JAVA_HOME: %w", err)
	}

	fmt.Println(theme.SuccessMessage("JAVA_HOME configured successfully"))
	fmt.Printf("  JAVA_HOME = %s\n", theme.PathStyle.Render(jdkPath))
	fmt.Println(theme.Faint.Render("  Added %JAVA_HOME%\\bin to PATH"))

	return nil
}
