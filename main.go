package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jv/internal/config"
	"jv/internal/env"
	"jv/internal/installer"
	"jv/internal/java"
	"jv/internal/theme"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Version is set during build time via ldflags
var Version = "dev"

// Use JV custom theme
var (
	successStyle = theme.SuccessStyle
	errorStyle   = theme.ErrorStyle
	warningStyle = theme.WarningStyle
	infoStyle    = theme.InfoStyle
	titleStyle   = theme.Title
	boxStyle     = theme.Box
	currentStyle = theme.CurrentStyle
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		handleList()
	case "use":
		handleUse()
	case "current":
		handleCurrent()
	case "add":
		handleAdd()
	case "remove":
		handleRemove()
	case "add-path":
		handleAddPath()
	case "remove-path":
		handleRemovePath()
	case "list-paths":
		handleListPaths()
	case "install":
		handleInstall()
	case "switch":
		handleSwitch()
	case "doctor":
		handleDoctor()
	case "repair":
		handleRepair()
	case "version", "-v", "--version":
		printVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleList() {
	detector := java.NewDetector()

	var versions []java.Version
	var scanErr error

	// Scan with spinner
	java.WithScanner(func() error {
		var err error
		versions, err = detector.FindAll()
		scanErr = err
		return nil
	})

	if scanErr != nil {
		fmt.Println(errorStyle.Render("Error finding Java versions: " + scanErr.Error()))
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println(warningStyle.Render("No Java installations found."))
		fmt.Println(infoStyle.Render("Run 'jv install' to install Java."))
		return
	}

	// Load config to get scope info
	cfg, _ := config.Load()
	scopeMap := make(map[string]string)
	for _, jdk := range cfg.InstalledJDKs {
		scopeMap[jdk.Path] = jdk.Scope
	}

	// Prefer system-wide JAVA_HOME (registry), fallback to process env
	current, _ := env.GetJavaHome()
	if current == "" {
		current = os.Getenv("JAVA_HOME")
	}

	fmt.Println(titleStyle.Render("Available Java Versions:"))
	fmt.Println()

	for _, v := range versions {
		marker := "  "
		versionStr := v.Version
		if strings.EqualFold(v.Path, current) {
			marker = "→ "
			versionStr = currentStyle.Render(v.Version)
		}

		source := "auto"
		sourceStyle := theme.Faint

		if v.IsCustom {
			source = "custom"
		}

		// Add scope info if available
		if scope, found := scopeMap[v.Path]; found {
			switch scope {
			case "system":
				source = "system-wide"
				sourceStyle = successStyle
			case "user":
				source = "user-only"
				sourceStyle = infoStyle
			}
		}

		// Align version column to width 15 considering visual width
		visW := lipgloss.Width(versionStr)
		pad := 0
		if visW < 15 {
			pad = 15 - visW
		}
		fmt.Printf("%s%s%s %s %s\n", marker, versionStr, strings.Repeat(" ", pad), v.Path, sourceStyle.Render("("+source+")"))
	}

	fmt.Println()

	if current == "" {
		// If system-wide JAVA_HOME exists in registry, suggest restart instead
		if sysJavaHome, err := env.GetJavaHome(); err == nil && sysJavaHome != "" {
			fmt.Println(theme.InfoMessage(" JAVA_HOME is set system-wide, but not visible in this session"))
			fmt.Println(theme.Faint.Render("  Restart your terminal to pick up environment changes"))
		} else {
			fmt.Println(theme.WarningMessage(" JAVA_HOME is not set"))
			fmt.Println(theme.Faint.Render("  Run 'jv use <version>' or 'jv switch' to configure"))
		}
	}
}

func handleUse() {
	detector := java.NewDetector()
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error finding Java versions: %v", err)))
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println(warningStyle.Render("No Java installations found."))
		fmt.Println(infoStyle.Render("Run 'jv install' to install Java."))
		os.Exit(1)
	}

	var target *java.Version

	// Interactive mode if no version specified
	if len(os.Args) < 3 {
		selected, err := selectJavaVersion(versions)
		if err != nil {
			fmt.Println(warningStyle.Render(fmt.Sprintf("Selection cancelled: %v", err)))
			os.Exit(1)
		}
		// If selected is already current, no-op
		current, _ := env.GetJavaHome()
		if current == "" {
			current = os.Getenv("JAVA_HOME")
		}
		if strings.EqualFold(selected.Path, current) {
			fmt.Println(infoStyle.Render(fmt.Sprintf("Already using Java %s. No changes needed.", selected.Version)))
			os.Exit(0)
		}
		target = selected
	} else {
		// Direct mode with version argument
		version := os.Args[2]
		for i, v := range versions {
			if strings.Contains(v.Version, version) {
				target = &versions[i]
				break
			}
		}

		if target == nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Java version '%s' not found.", version)))
			fmt.Println(infoStyle.Render("Use 'jv list' to see available versions."))
			os.Exit(1)
		}

		// If specified version is already current, no-op
		current, _ := env.GetJavaHome()
		if current == "" {
			current = os.Getenv("JAVA_HOME")
		}
		if strings.EqualFold(target.Path, current) {
			fmt.Println(infoStyle.Render(fmt.Sprintf("Already using Java %s. No changes needed.", target.Version)))
			os.Exit(0)
		}
	}

	// Confirm switch
	confirmed, err := confirmAction(
		fmt.Sprintf("Switch to Java %s?", target.Version),
		fmt.Sprintf("Path: %s", target.Path),
	)
	if err != nil || !confirmed {
		fmt.Println(warningStyle.Render("Operation cancelled."))
		os.Exit(0)
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("Switching to Java %s...", target.Version)))

	if err := env.SetJavaHome(target.Path); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		fmt.Println()
		fmt.Println(warningStyle.Render("Note: This command requires administrator privileges."))
		fmt.Println(theme.Faint.Render("Please run your terminal as Administrator and try again."))
		os.Exit(1)
	}

	fmt.Println(successStyle.Render("✓ Successfully updated JAVA_HOME!"))
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Faint(true).Render("Note: You may need to restart your terminal or applications for changes to take effect."))
}

func handleCurrent() {
	// Prefer system-wide JAVA_HOME (registry), fallback to process env
	javaHome, _ := env.GetJavaHome()
	if javaHome == "" {
		javaHome = os.Getenv("JAVA_HOME")
	}

	fmt.Println(titleStyle.Render("Current Java"))
	fmt.Println()

	if javaHome == "" {
		fmt.Println(warningStyle.Render("JAVA_HOME is not set"))
		fmt.Println(theme.Faint.Render("Run 'jv use <version>' or 'jv switch' to configure"))
		return
	}

	detector := java.NewDetector()
	version := detector.GetVersion(javaHome)
	isValid := detector.IsValidJavaPath(javaHome)

	// Labeled fields with theme
	fmt.Printf("%s %s\n", theme.LabelStyle.Render("Version:"), currentStyle.Render(version))
	fmt.Printf("%s %s\n", theme.LabelStyle.Render("JAVA_HOME:"), theme.PathStyle.Render(javaHome))

	if !isValid {
		fmt.Println()
		fmt.Println(warningStyle.Render("JAVA_HOME path looks invalid"))
		fmt.Println(theme.Faint.Render("Use 'jv use' to fix it or 'jv repair' for assistance"))
	}
}

func handleAdd() {
	if len(os.Args) < 3 {
		fmt.Println(errorStyle.Render("Usage: jv add <path>"))
		fmt.Println(infoStyle.Render("Example: jv add C:\\custom\\jdk-21"))
		os.Exit(1)
	}

	path := os.Args[2]

	detector := java.NewDetector()
	if !detector.IsValidJavaPath(path) {
		fmt.Printf("Invalid Java installation path: %s\n", path)
		fmt.Println("Make sure the path contains bin\\java.exe")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.HasCustomPath(path) {
		fmt.Println(warningStyle.Render("This path is already in the custom paths list."))
		return
	}

	version := detector.GetVersion(path)

	// Confirm addition
	confirmed, err := confirmAction(
		fmt.Sprintf("Add Java %s?", version),
		fmt.Sprintf("Path: %s", path),
	)
	if err != nil || !confirmed {
		fmt.Println("Operation cancelled.")
		return
	}

	cfg.AddCustomPath(path)
	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Added Java %s to custom paths.\n", version)
}

func handleRemove() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(errorStyle.Render("Error loading config: " + err.Error()))
		os.Exit(1)
	}

	var pathToRemove string

	// Interactive mode if no path specified
	if len(os.Args) < 3 {
		if len(cfg.CustomPaths) == 0 {
			fmt.Println(theme.InfoMessage("No custom Java installations to remove"))
			fmt.Println("  " + theme.Faint.Render("Use ") + theme.Code.Render("jv add <path>") + theme.Faint.Render(" to add one"))
			return
		}

		// Build options with aligned version and current highlighting
		detector := java.NewDetector()

		// precompute version widths (styled)
		maxW := 0
		versions := make([]string, len(cfg.CustomPaths))
		for i, p := range cfg.CustomPaths {
			v := detector.GetVersion(p)
			versions[i] = v
			rv := theme.CurrentStyle.Render(v)
			if w := lipgloss.Width(rv); w > maxW {
				maxW = w
			}
		}

		options := make([]huh.Option[string], len(cfg.CustomPaths))
		for i, p := range cfg.CustomPaths {
			v := versions[i]
			// Always render version in Java orange (match use/switch visual identity)
			ver := theme.CurrentStyle.Render(v)
			// pad to fixed 15 cols based on rendered width
			vis := lipgloss.Width(ver)
			pad := ""
			if vis < 15 {
				pad = strings.Repeat(" ", 15-vis)
			}
			label := fmt.Sprintf("%s%s %s %s", ver, pad, p, theme.Faint.Render("(custom)"))
			options[i] = huh.NewOption(label, p)
		}

		err := huh.NewSelect[string]().
			Title(theme.Subtitle.Render("Select Java Installation to Remove")).
			Description(theme.Faint.Render("Use arrow keys to navigate, Enter to select")).
			Options(options...).
			Value(&pathToRemove).
			Run()

		if err != nil {
			fmt.Println(warningStyle.Render(fmt.Sprintf("Selection cancelled: %v", err)))
			os.Exit(1)
		}
	} else {
		pathToRemove = os.Args[2]
	}

	if !cfg.HasCustomPath(pathToRemove) {
		fmt.Println(warningStyle.Render("This path is not in the custom paths list."))
		return
	}

	// Confirm removal
	detector := java.NewDetector()
	version := detector.GetVersion(pathToRemove)
	confirmed, err := confirmAction(
		fmt.Sprintf("Remove Java %s?", version),
		fmt.Sprintf("Path: %s", pathToRemove),
	)
	if err != nil || !confirmed {
		fmt.Println(warningStyle.Render("Operation cancelled."))
		return
	}

	cfg.RemoveCustomPath(pathToRemove)
	cfg.RemoveInstalledJDK(pathToRemove) // Also remove from installed JDKs if present

	if err := cfg.Save(); err != nil {
		fmt.Println(errorStyle.Render("Error saving config: " + err.Error()))
		os.Exit(1)
	}

	fmt.Println(successStyle.Render("✓ Removed from custom paths."))
}

func handleAddPath() {
	if len(os.Args) < 3 {
		fmt.Println(errorStyle.Render("Usage: jv add-path <directory>"))
		fmt.Println(infoStyle.Render("Example: jv add-path C:\\DevTools\\Java"))
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Faint(true).Render("This adds a directory where the detector will search for Java installations."))
		os.Exit(1)
	}

	path := os.Args[2]

	detector := java.NewDetector()
	if !detector.IsValidSearchPath(path) {
		fmt.Printf("Invalid directory path: %s\n", path)
		fmt.Println("Make sure the path exists and is a directory.")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.HasSearchPath(path) {
		fmt.Println(warningStyle.Render("This search path is already configured."))
		return
	}

	// Confirm addition
	confirmed, err := confirmAction(
		"Add search path?",
		fmt.Sprintf("Path: %s\n\nThe detector will scan this directory for Java installations.", path),
	)
	if err != nil || !confirmed {
		fmt.Println("Operation cancelled.")
		return
	}

	cfg.AddSearchPath(path)
	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(theme.SuccessMessage("Added search path:"))
	fmt.Println("  " + theme.PathStyle.Render(path))
	fmt.Println(theme.Faint.Render("Run ") + theme.Code.Render("jv list") + theme.Faint.Render(" to see detected versions"))
}

func handleRemovePath() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(errorStyle.Render("Error loading config: " + err.Error()))
		os.Exit(1)
	}

	var pathToRemove string

	// Interactive mode if no path specified
	if len(os.Args) < 3 {
		if len(cfg.SearchPaths) == 0 {
			fmt.Println(theme.InfoMessage("No custom search paths to remove"))
			fmt.Println("  " + theme.Faint.Render("Use ") + theme.Code.Render("jv add-path <directory>") + theme.Faint.Render(" to add one"))
			return
		}

		// Build options with status tag and alignment
		detector := java.NewDetector()
		maxW := 0
		for _, p := range cfg.SearchPaths {
			rp := theme.CurrentStyle.Render(p)
			if w := lipgloss.Width(rp); w > maxW {
				maxW = w
			}
		}

		options := make([]huh.Option[string], len(cfg.SearchPaths))
		for i, p := range cfg.SearchPaths {
			renderedPath := theme.CurrentStyle.Render(p)
			pad := ""
			if w := lipgloss.Width(renderedPath); w < maxW {
				pad = strings.Repeat(" ", maxW-w)
			}
			status := theme.Faint.Render("Not found")
			if detector.IsValidSearchPath(p) {
				status = theme.SuccessStyle.Render("✓ Exists")
			}
			label := fmt.Sprintf("%s%s  %s", renderedPath, pad, status)
			options[i] = huh.NewOption(label, p)
		}

		err := huh.NewSelect[string]().
			Title(theme.Subtitle.Render("Select Search Path to Remove")).
			Description(theme.Faint.Render("Use arrow keys to navigate, Enter to select")).
			Options(options...).
			Value(&pathToRemove).
			Run()

		if err != nil {
			fmt.Println(warningStyle.Render(fmt.Sprintf("Selection cancelled: %v", err)))
			os.Exit(1)
		}
	} else {
		pathToRemove = os.Args[2]
	}

	if !cfg.HasSearchPath(pathToRemove) {
		fmt.Println(warningStyle.Render("This path is not in the search paths list."))
		return
	}

	// Confirm removal
	confirmed, err := confirmAction(
		"Remove search path?",
		fmt.Sprintf("Path: %s", pathToRemove),
	)
	if err != nil || !confirmed {
		fmt.Println(warningStyle.Render("Operation cancelled."))
		return
	}

	cfg.RemoveSearchPath(pathToRemove)
	if err := cfg.Save(); err != nil {
		fmt.Println(errorStyle.Render("Error saving config: " + err.Error()))
		os.Exit(1)
	}

	fmt.Println(successStyle.Render("✓ Removed search path."))
}

func handleListPaths() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(errorStyle.Render("Error loading config: " + err.Error()))
		os.Exit(1)
	}

	detector := java.NewDetector()

	fmt.Println(titleStyle.Render("Java Search Paths"))
	fmt.Println()

	// Table styles from theme
	headerStyle := theme.TableHeader
	cellStyle := theme.TableCell
	existsStyle := theme.SuccessStyle.Padding(0, 1)
	notFoundStyle := theme.ErrorStyle.Padding(0, 1)
	tableStyle := theme.TableStyle

	// Standard paths table
	fmt.Println(theme.LabelStyle.Render("Standard Paths (built-in):"))
	fmt.Println()

	standardPaths := []string{
		"C:\\Program Files\\Java",
		"C:\\Program Files (x86)\\Java",
		"C:\\Program Files\\Eclipse Adoptium",
		"C:\\Program Files\\Eclipse Foundation",
		"C:\\Program Files\\Zulu",
		"C:\\Program Files\\Amazon Corretto",
		"C:\\Program Files\\Microsoft",
	}

	var rows []string
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
		headerStyle.Render("Path"),
		headerStyle.Width(50).Render(""),
		headerStyle.Render("Status"),
	))

	for _, p := range standardPaths {
		exists := detector.IsValidSearchPath(p)
		status := ""
		if exists {
			status = existsStyle.Render("✓ Exists")
		} else {
			status = cellStyle.Faint(true).Render("Not found")
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			cellStyle.Width(58).Render(p),
			status,
		))
	}

	table := lipgloss.JoinVertical(lipgloss.Left, rows...)
	fmt.Println(tableStyle.Render(table))
	fmt.Println()

	// Custom paths table
	if len(cfg.SearchPaths) > 0 {
		fmt.Println(theme.LabelStyle.Render("Custom Search Paths:"))
		fmt.Println()

		var customRows []string
		customRows = append(customRows, lipgloss.JoinHorizontal(lipgloss.Left,
			headerStyle.Render("Path"),
			headerStyle.Width(50).Render(""),
			headerStyle.Render("Status"),
		))

		for _, p := range cfg.SearchPaths {
			exists := detector.IsValidSearchPath(p)
			status := ""
			if exists {
				status = existsStyle.Render("✓ Exists")
			} else {
				status = notFoundStyle.Render("✗ Not found")
			}

			customRows = append(customRows, lipgloss.JoinHorizontal(lipgloss.Left,
				cellStyle.Width(58).Render(p),
				status,
			))
		}

		customTable := lipgloss.JoinVertical(lipgloss.Left, customRows...)
		fmt.Println(tableStyle.Render(customTable))
	} else {
		fmt.Println(infoStyle.Render("No custom search paths configured."))
		fmt.Println(theme.Faint.Render("Use 'jv add-path <directory>' to add one."))
	}
	fmt.Println()
}

func handleInstall() {
	// Check admin privileges
	isAdmin := env.IsAdmin()

	// Create installer
	inst, err := installer.NewInstaller(isAdmin)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Run interactive installation
	if err := inst.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleSwitch() {
	// Always interactive - ignore any arguments
	detector := java.NewDetector()
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error finding Java versions: %v", err)))
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println(warningStyle.Render("No Java installations found."))
		fmt.Println(infoStyle.Render("Run 'jv install' to install Java."))
		os.Exit(1)
	}

	// Show interactive selector
	target, err := selectJavaVersion(versions)
	if err != nil {
		fmt.Println(warningStyle.Render(fmt.Sprintf("Selection cancelled: %v", err)))
		os.Exit(1)
	}

	// If selected is already current, no-op
	current, _ := env.GetJavaHome()
	if current == "" {
		current = os.Getenv("JAVA_HOME")
	}
	if strings.EqualFold(target.Path, current) {
		fmt.Println(infoStyle.Render(fmt.Sprintf("Already using Java %s. No changes needed.", target.Version)))
		os.Exit(0)
	}

	// Confirm switch
	confirmed, err := confirmAction(
		fmt.Sprintf("Switch to Java %s?", target.Version),
		fmt.Sprintf("Path: %s", target.Path),
	)
	if err != nil || !confirmed {
		fmt.Println(warningStyle.Render("Operation cancelled."))
		os.Exit(0)
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("Switching to Java %s...", target.Version)))

	if err := env.SetJavaHome(target.Path); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		fmt.Println()
		fmt.Println(warningStyle.Render("Note: This command requires administrator privileges."))
		fmt.Println(theme.Faint.Render("Please run your terminal as Administrator and try again."))
		os.Exit(1)
	}

	fmt.Println(successStyle.Render("✓ Successfully updated JAVA_HOME!"))
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Faint(true).Render("Note: You may need to restart your terminal or applications for changes to take effect."))
}

func handleDoctor() {
	fmt.Println(titleStyle.Render("Java Version Switcher - System Diagnostics"))
	fmt.Println()

	issues := []string{}
	warnings := []string{}

	// Detector and current JAVA_HOME
	detector := java.NewDetector()
	currentJavaHome, _ := env.GetJavaHome()
	if currentJavaHome == "" {
		currentJavaHome = os.Getenv("JAVA_HOME")
	}

	// 1. Check JAVA_HOME
	fmt.Println(theme.LabelStyle.Render("Checking JAVA_HOME..."))
	if currentJavaHome == "" {
		fmt.Println("  " + theme.ErrorMessage("JAVA_HOME is not set"))
		issues = append(issues, "JAVA_HOME is not set")
	} else if detector.IsValidJavaPath(currentJavaHome) {
		fmt.Printf("  %s %s\n", theme.SuccessMessage("JAVA_HOME is set and valid:"), theme.PathStyle.Render(currentJavaHome))
	} else {
		fmt.Printf("  %s %s\n", theme.ErrorStyle.Render("✗ JAVA_HOME is set but invalid:"), theme.PathStyle.Render(currentJavaHome))
		issues = append(issues, fmt.Sprintf("JAVA_HOME points to invalid location: %s", currentJavaHome))
	}
	fmt.Println()

	// 2. Check PATH
	fmt.Println(theme.LabelStyle.Render("Checking Path..."))
	pathEnv := os.Getenv("Path")

	// Expected entry: <JAVA_HOME>\bin (resolved from registry/env)
	expectedBin := ""
	if currentJavaHome != "" {
		expectedBin = strings.TrimRight(filepath.Clean(filepath.Join(currentJavaHome, "bin")), "\\")
	}

	hasJavaHomeInPath := false
	if expectedBin != "" {
		expectedLower := strings.ToLower(strings.TrimRight(expectedBin, "\\"))
		for _, entry := range strings.Split(pathEnv, ";") {
			e := strings.TrimSpace(strings.Trim(entry, "\""))
			if e == "" {
				continue
			}
			eLower := strings.ToLower(strings.TrimRight(e, "\\"))
			if eLower == expectedLower {
				hasJavaHomeInPath = true
				break
			}
		}
	}

	if hasJavaHomeInPath {
		fmt.Println("  " + theme.SuccessMessage("%JAVA_HOME%\\bin is in Path"))
	} else {
		fmt.Println("  " + theme.ErrorMessage("No Java found in Path"))
		issues = append(issues, "%JAVA_HOME%\\bin is not in Path")
	}
	fmt.Println()

	// 3. Check Java installations with table
	fmt.Println(theme.LabelStyle.Render("Checking Java installations..."))
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Printf("  %s %v\n", theme.ErrorStyle.Render("✗ Error finding Java versions:"), err)
		issues = append(issues, fmt.Sprintf("Error detecting Java installations: %v", err))
	} else if len(versions) == 0 {
		fmt.Println(theme.WarningMessage("No Java installations found"))
		warnings = append(warnings, "No Java installations detected. Run 'jv install' to install Java.")
	} else {
		fmt.Printf("  %s %d\n", theme.SuccessMessage("Found installations:"), len(versions))

		// Build table
		headerStyle := theme.TableHeader
		cellStyle := theme.TableCell
		tableStyle := theme.TableStyle

		var rows []string
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
			headerStyle.Width(9).Render("Current"),
			headerStyle.Width(12).Render("Version"),
			headerStyle.Width(58).Render("Path"),
			headerStyle.Render("Source"),
		))

		for _, v := range versions {
			currentMark := ""
			versionStr := v.Version
			source := "auto"
			sourceStyle := theme.Faint
			if v.IsCustom {
				source = "custom"
				sourceStyle = infoStyle
			}
			if strings.EqualFold(v.Path, currentJavaHome) {
				currentMark = theme.SuccessMessage("")
				versionStr = currentStyle.Render(versionStr)
			}

			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left,
				cellStyle.Width(9).Align(lipgloss.Center).Render(currentMark),
				cellStyle.Width(12).Render(versionStr),
				cellStyle.Width(58).Render(v.Path),
				sourceStyle.Render(source),
			))
		}
		table := lipgloss.JoinVertical(lipgloss.Left, rows...)
		fmt.Println(tableStyle.Render(table))
	}
	fmt.Println()

	// 4. Check configuration file
	fmt.Println(theme.LabelStyle.Render("Checking configuration..."))
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("  ✗ Error loading config: %v\n", err)
		issues = append(issues, fmt.Sprintf("Configuration file error: %v", err))
	} else {
		homeDir, _ := os.UserHomeDir()
		configPath := homeDir + "\\.config\\jv\\jv.json"
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			configPath = os.Getenv("XDG_CONFIG_HOME") + "\\jv\\jv.json"
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Println("  " + theme.WarningMessage("Configuration file does not exist (will be created when needed)"))
		} else {
			fmt.Println("  " + theme.SuccessMessage("Configuration file exists and is valid"))
		}
		if len(cfg.CustomPaths) > 0 {
			fmt.Println("  " + theme.SuccessMessage(fmt.Sprintf("Custom paths configured: %d", len(cfg.CustomPaths))))
		}
		if len(cfg.SearchPaths) > 0 {
			fmt.Println("  " + theme.SuccessMessage(fmt.Sprintf("Search paths configured: %d", len(cfg.SearchPaths))))
		}
		if len(cfg.InstalledJDKs) > 0 {
			fmt.Println("  " + theme.SuccessMessage(fmt.Sprintf("Tracked JDKs: %d", len(cfg.InstalledJDKs))))
		}
	}
	fmt.Println()

	// 5. Check administrator privileges
	fmt.Println(theme.LabelStyle.Render("Checking privileges..."))
	isAdmin := env.IsAdmin()
	if isAdmin {
		fmt.Println("  " + theme.SuccessMessage("Running with administrator privileges"))
	} else {
		fmt.Println("  " + theme.WarningMessage("Not running as administrator (some operations require admin)"))
		warnings = append(warnings, "Administrator privileges may be required for 'jv use' and 'jv repair'")
	}
	fmt.Println()

	// 6. Check if jv.exe is accessible
	fmt.Println(theme.LabelStyle.Render("Checking jv tool..."))
	if _, err := os.Executable(); err != nil {
		fmt.Println("  " + theme.WarningMessage("Could not determine jv executable path"))
	} else {
		fmt.Println("  " + theme.SuccessMessage("jv tool is accessible"))
	}
	fmt.Println()

	// Summary
	fmt.Println()
	fmt.Println(titleStyle.Render("Diagnostics Summary"))
	fmt.Println()

	if len(issues) == 0 && len(warnings) == 0 {
		successBox := theme.SuccessBox.Render(theme.SuccessMessage("All checks passed!") + "\n\nYour Java environment is properly configured.")
		fmt.Println(successBox)
		return
	}

	// Build summary content
	var summaryContent string

	if len(issues) > 0 {
		summaryContent += errorStyle.Render(fmt.Sprintf("Issues Found: %d", len(issues))) + "\n\n"
		for _, issue := range issues {
			summaryContent += theme.ErrorMessage(issue) + "\n"
		}
	}

	if len(warnings) > 0 {
		if len(issues) > 0 {
			summaryContent += "\n"
		}
		summaryContent += warningStyle.Render(fmt.Sprintf("Warnings: %d", len(warnings))) + "\n\n"
		for _, warning := range warnings {
			summaryContent += theme.WarningMessage(warning) + "\n"
		}
	}

	if len(issues) > 0 {
		summaryContent += "\n" + theme.InfoMessage(" Run 'jv repair' to fix issues")
		summaryContent += "\n" + theme.Faint.Render("  (Note: requires administrator privileges)")
	}

	fmt.Println(boxStyle.Render(summaryContent))
}

type RepairIssue struct {
	ID            string
	Description   string
	RequiresAdmin bool
	CanFix        bool
}

func handleRepair() {
	// Themed header
	header := theme.Title.Padding(0, 2).Render("Java Version Switcher - Auto Repair")
	fmt.Println(theme.TitleBox.Render(header))
	fmt.Println()

	isAdmin := env.IsAdmin()
	if !isAdmin {
		fmt.Println(theme.WarningMessage("Not running as Administrator"))
		fmt.Println(theme.Faint.Render("   Some repairs require administrator privileges."))
		fmt.Println()
	}

	detector := java.NewDetector()

	// Step 1: Detect all issues
	fmt.Println(theme.LabelStyle.Render("Scanning for issues..."))
	fmt.Println()

	// Find Java installations first
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Println(theme.ErrorMessage(fmt.Sprintf("Error finding Java versions: %v", err)))
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println(theme.ErrorMessage("No Java installations found."))
		fmt.Println(theme.Faint.Render("Please install Java first: jv install"))
		os.Exit(1)
	}

	// Detect all issues
	issues := []RepairIssue{}
	currentJavaHome := os.Getenv("JAVA_HOME")

	// Issue 1: JAVA_HOME not set or invalid
	if currentJavaHome == "" {
		issues = append(issues, RepairIssue{
			ID:            "java_home_not_set",
			Description:   "JAVA_HOME is not set",
			RequiresAdmin: true,
			CanFix:        isAdmin,
		})
	} else if !detector.IsValidJavaPath(currentJavaHome) {
		issues = append(issues, RepairIssue{
			ID:            "java_home_invalid",
			Description:   fmt.Sprintf("JAVA_HOME is invalid: %s", currentJavaHome),
			RequiresAdmin: true,
			CanFix:        isAdmin,
		})
	}

	// Issue 2: PATH doesn't contain %JAVA_HOME%\bin (check resolved <JAVA_HOME>\bin exactly)
	pathEnv := os.Getenv("Path")
	hasJavaHomeInPath := false
	if currentJavaHome != "" {
		expectedBin := strings.ToLower(strings.TrimRight(filepath.Join(currentJavaHome, "bin"), "\\"))
		for _, entry := range strings.Split(pathEnv, ";") {
			e := strings.TrimSpace(strings.Trim(entry, "\""))
			if e == "" {
				continue
			}
			eLower := strings.ToLower(strings.TrimRight(e, "\\"))
			if eLower == expectedBin {
				hasJavaHomeInPath = true
				break
			}
		}
	}
	if !hasJavaHomeInPath && currentJavaHome != "" {
		issues = append(issues, RepairIssue{
			ID:            "path_missing_java_home",
			Description:   "%JAVA_HOME%\\bin is not in PATH",
			RequiresAdmin: true,
			CanFix:        isAdmin,
		})
	}

	// Issue 3: Config file problems
	if _, err := config.Load(); err != nil {
		issues = append(issues, RepairIssue{
			ID:            "config_error",
			Description:   fmt.Sprintf("Configuration file error: %v", err),
			RequiresAdmin: false,
			CanFix:        true,
		})
	}

	// No issues found
	if len(issues) == 0 {
		fmt.Println(successStyle.Render("✓ No issues found - your environment is properly configured!"))
		return
	}

	// Show issues
	fmt.Printf("%s %d issue(s):\n", theme.LabelStyle.Render("Found"), len(issues))
	for i, issue := range issues {
		adminMarker := ""
		if issue.RequiresAdmin {
			adminMarker = " [Requires Admin]"
		}
		if !issue.CanFix {
			adminMarker += " [Cannot Fix]"
		}
		fmt.Printf("  %d. %s%s\n", i+1, theme.WarningStyle.Render(issue.Description), adminMarker)
	}
	fmt.Println()

	// Filter fixable issues
	fixableIssues := []RepairIssue{}
	for _, issue := range issues {
		if issue.CanFix {
			fixableIssues = append(fixableIssues, issue)
		}
	}

	if len(fixableIssues) == 0 {
		fmt.Println(theme.ErrorMessage("No fixable issues (some require administrator privileges)"))
		fmt.Println(theme.Faint.Render("  Run as administrator to fix all issues"))
		os.Exit(1)
	}

	// Interactive selection of issues to fix
	options := make([]huh.Option[string], len(fixableIssues))
	for i, issue := range fixableIssues {
		options[i] = huh.NewOption(theme.WarningStyle.Render(issue.Description), issue.ID)
	}

	var selectedIssues []string
	err = huh.NewMultiSelect[string]().
		Title(theme.Subtitle.Render("Select Issues to Fix")).
		Description(theme.Faint.Render("Use Space to select, Enter to confirm")).
		Options(options...).
		Value(&selectedIssues).
		Run()

	if err != nil || len(selectedIssues) == 0 {
		fmt.Println("No issues selected. Repair cancelled.")
		return
	}

	// Perform repairs
	fmt.Println()
	fmt.Println(theme.LabelStyle.Render("Performing repairs..."))
	fmt.Println()

	repaired := []string{}
	for _, issueID := range selectedIssues {
		switch issueID {
		case "java_home_not_set", "java_home_invalid":
			// Let user select which Java to use (themed preamble)
			fmt.Println(theme.LabelStyle.Render("Select Java to set as JAVA_HOME"))
			fmt.Println(theme.Faint.Render("Use arrow keys to navigate, Enter to select"))
			target, err := selectJavaVersion(versions)
			if err != nil {
				fmt.Printf("  %s %v\n", theme.ErrorMessage("Skipped JAVA_HOME repair:"), err)
				continue
			}

			if err := env.SetJavaHome(target.Path); err != nil {
				fmt.Printf("  %s %v\n", theme.ErrorMessage("Failed to set JAVA_HOME:"), err)
				continue
			}

			repaired = append(repaired, fmt.Sprintf("Set JAVA_HOME to %s", target.Path))
			fmt.Println("  " + theme.SuccessMessage(fmt.Sprintf("JAVA_HOME set to Java %s", target.Version)))

		case "path_missing_java_home":
			// Ensure PATH has %JAVA_HOME%\\bin by reapplying SetJavaHome
			targetPath := currentJavaHome
			if targetPath == "" {
				// pick version interactively (themed preamble)
				fmt.Println(theme.LabelStyle.Render("Select Java to set as JAVA_HOME"))
				fmt.Println(theme.Faint.Render("Use arrow keys to navigate, Enter to select"))
				t, err := selectJavaVersion(versions)
				if err != nil {
					fmt.Printf("  %s %v\n", theme.ErrorMessage("Skipped PATH repair:"), err)
					continue
				}
				targetPath = t.Path
			}
			if err := env.SetJavaHome(targetPath); err != nil {
				fmt.Printf("  %s %v\n", theme.ErrorMessage("Failed to update PATH:"), err)
				continue
			}
			repaired = append(repaired, "Added %JAVA_HOME%\\bin to PATH")
			fmt.Println(theme.SuccessMessage("PATH updated"))

		case "config_error":
			cfg, err := config.Load()
			if err == nil {
				if err := cfg.Save(); err == nil {
					repaired = append(repaired, "Repaired configuration file")
					fmt.Println(theme.SuccessMessage("Configuration file repaired"))
				}
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Println(theme.Title.Render("Repair Complete"))
	fmt.Println()

	if len(repaired) == 0 {
		fmt.Println(theme.ErrorMessage("No repairs were successful"))
		return
	}

	fmt.Println(theme.LabelStyle.Render("Repairs performed:"))
	for _, repair := range repaired {
		fmt.Println("  " + theme.SuccessMessage(repair))
	}
	fmt.Println()
	fmt.Println(theme.Faint.Render("Note: You may need to restart your terminal for changes to take effect."))
}

func printVersion() {
	linkStyle := lipgloss.NewStyle().
		Foreground(theme.Info).
		Underline(true)

	fmt.Printf("%s %s %s\n",
		theme.Subtitle.Render("Java Version Switcher (jv)"),
		theme.Faint.Render("version"),
		theme.HighlightText(Version))
	fmt.Println(linkStyle.Render("https://github.com/CostaBrosky/jv"))
	fmt.Println()

	// Add features badge
	fmt.Println(theme.SuccessStyle.Italic(true).Render("✨ Interactive TUI powered by Huh! and Lip Gloss"))
}

func printUsage() {
	// ASCII Art Banner with JV theme
	banner := `     ██╗██╗   ██╗
     ██║██║   ██║
     ██║╚██╗ ██╔╝
██   ██║ ╚████╔╝ 
╚█████╔╝  ╚██╔╝  
 ╚════╝    ╚═╝   `

	fmt.Println(theme.Banner.Render(banner))
	fmt.Println(theme.Subtitle.Render("Java Version Switcher"))
	fmt.Println(theme.Faint.Render("Easy Java version management for Windows"))
	fmt.Println()

	// Usage section
	fmt.Println(theme.Title.Render("USAGE"))
	fmt.Println(theme.Faint.Render("  jv <command> [arguments]"))
	fmt.Println()

	// Command categories use theme
	categoryStyle := theme.Subtitle
	commandStyle := theme.CommandStyle
	descStyle := theme.Faint

	fmt.Println(categoryStyle.Render("INSTALLATION & SETUP"))
	fmt.Printf("  %s            %s\n",
		commandStyle.Render("install"),
		descStyle.Render("Install Java from open-source distributors"))
	fmt.Printf("  %s             %s\n",
		commandStyle.Render("doctor"),
		descStyle.Render("Run diagnostics on your Java environment"))
	fmt.Printf("  %s             %s\n",
		commandStyle.Render("repair"),
		descStyle.Render("Automatically fix configuration issues"))
	fmt.Println()

	fmt.Println(categoryStyle.Render("VERSION MANAGEMENT"))
	fmt.Printf("  %s               %s\n",
		commandStyle.Render("list"),
		descStyle.Render("List all available Java versions"))
	fmt.Printf("  %s [version]      %s\n",
		commandStyle.Render("use"),
		descStyle.Render("Switch to Java version"))
	fmt.Printf("  %s             %s\n",
		commandStyle.Render("switch"),
		descStyle.Render("Quick interactive version switcher"))
	fmt.Printf("  %s            %s\n",
		commandStyle.Render("current"),
		descStyle.Render("Show current Java version"))
	fmt.Println()

	fmt.Println(categoryStyle.Render("CUSTOM INSTALLATIONS"))
	fmt.Printf("  %s <path>         %s\n",
		commandStyle.Render("add"),
		descStyle.Render("Add a specific Java installation"))
	fmt.Printf("  %s [path]      %s\n",
		commandStyle.Render("remove"),
		descStyle.Render("Remove a custom installation"))
	fmt.Println()

	fmt.Println(categoryStyle.Render("SEARCH PATHS"))
	fmt.Printf("  %s <dir>     %s\n",
		commandStyle.Render("add-path"),
		descStyle.Render("Add directory to scan for Java installations"))
	fmt.Printf("  %s [dir]  %s\n",
		commandStyle.Render("remove-path"),
		descStyle.Render("Remove directory from search paths"))
	fmt.Printf("  %s         %s\n",
		commandStyle.Render("list-paths"),
		descStyle.Render("Show all search paths (standard + custom)"))
	fmt.Println()

	fmt.Println(categoryStyle.Render("OTHER"))
	fmt.Printf("  %s            %s\n",
		commandStyle.Render("version"),
		descStyle.Render("Show version information"))
	fmt.Printf("  %s               %s\n",
		commandStyle.Render("help"),
		descStyle.Render("Show this help message"))
	fmt.Println()

	// Examples section
	fmt.Println(theme.Title.Render("EXAMPLES"))
	fmt.Println("  " + theme.Code.Render("jv list") + "                  # List Java versions")
	fmt.Println("  " + theme.Code.Render("jv switch") + "                # Interactive switcher")
	fmt.Println("  " + theme.Code.Render("jv use 17") + "                # Switch to Java 17")
	fmt.Println("  " + theme.Code.Render("jv install") + "               # Install Java interactively")
	fmt.Println("  " + theme.Code.Render("jv add C:\\custom\\jdk-21") + "  # Add custom installation")
	fmt.Println("  " + theme.Code.Render("jv doctor") + "                # Check system health")
	fmt.Println()

	// Autocomplete note
	fmt.Println(theme.InfoStyle.Italic(true).Render("💡 Tip: PowerShell autocomplete is installed automatically by the installer"))

	fmt.Println()

	// Note section with theme
	note := theme.WarningBox.Render("⚠  Administrator privileges required for: use, switch, install, repair")
	fmt.Println(note)
	fmt.Println()

	// Footer with theme
	fmt.Println(theme.Faint.Italic(true).Render("For more information: https://github.com/CostaBrosky/jv"))
}

// selectJavaVersion shows an interactive selector for Java versions
func selectJavaVersion(versions []java.Version) (*java.Version, error) {
	// Load config to show scope info
	cfg, _ := config.Load()
	scopeMap := make(map[string]string)
	for _, jdk := range cfg.InstalledJDKs {
		scopeMap[jdk.Path] = jdk.Scope
	}

	// Prefer system-wide JAVA_HOME (registry), fallback to process env
	current, _ := env.GetJavaHome()
	if current == "" {
		current = os.Getenv("JAVA_HOME")
	}

	// Reorder: put current first
	ordered := make([]java.Version, 0, len(versions))
	for _, v := range versions {
		if strings.EqualFold(v.Path, current) {
			ordered = append(ordered, v)
		}
	}
	for _, v := range versions {
		if !strings.EqualFold(v.Path, current) {
			ordered = append(ordered, v)
		}
	}

	// Build options with themed parts (same as use/switch)
	options := make([]huh.Option[int], len(ordered))
	for i, v := range ordered {
		// Version part (highlight current or all when no current is set)
		versionPart := v.Version
		if current == "" {
			versionPart = currentStyle.Render(v.Version)
		} else if strings.EqualFold(v.Path, current) {
			versionPart = currentStyle.Render(v.Version)
		}

		// Compute padding based on visual width to align columns
		versionWidth := lipgloss.Width(versionPart)
		pad := 0
		if versionWidth < 15 {
			pad = 15 - versionWidth
		}
		padSpaces := strings.Repeat(" ", pad)

		// Path part (leave unstyled to let focused highlight be visible)
		pathPart := v.Path

		// Scope/info part
		scopeTag := "(auto)"
		scopeStyle := theme.Faint
		if scope, found := scopeMap[v.Path]; found {
			switch scope {
			case "system":
				scopeTag = "(system-wide)"
				scopeStyle = successStyle
			case "user":
				scopeTag = "(user-only)"
				scopeStyle = infoStyle
			}
		} else if v.IsCustom {
			scopeTag = "(custom)"
			scopeStyle = theme.Bold
		}

		label := fmt.Sprintf("%s%s %s %s", versionPart, padSpaces, pathPart, scopeStyle.Render(scopeTag))
		// Mark current explicitly
		if strings.EqualFold(v.Path, current) {
			label += " " + theme.Faint.Render("[current]")
		}

		options[i] = huh.NewOption(label, i)
	}

	var selectedIdx int

	err := huh.NewSelect[int]().
		Title(theme.Subtitle.Render("Select Java Version")).
		Description(theme.Faint.Render("Use arrow keys to navigate, Enter to select")).
		Options(options...).
		Value(&selectedIdx).
		Run()

	if err != nil {
		return nil, err
	}

	return &ordered[selectedIdx], nil
}

// confirmAction shows a confirmation prompt
func confirmAction(title, description string) (bool, error) {
	var confirmed bool

	err := huh.NewConfirm().
		Title(theme.Subtitle.Render(title)).
		Description(theme.Faint.Render(description)).
		Affirmative(theme.SuccessStyle.Render("Yes")).
		Negative(theme.ErrorStyle.Render("No")).
		Value(&confirmed).
		Run()

	return confirmed, err
}
