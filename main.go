package main

import (
	"fmt"
	"os"
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

	current := os.Getenv("JAVA_HOME")

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

		fmt.Printf("%s%-15s %s %s\n", marker, versionStr, v.Path, sourceStyle.Render("("+source+")"))
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
		fmt.Printf("Error finding Java versions: %v\n", err)
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println("No Java installations found.")
		fmt.Println("Run 'jv install' to install Java.")
		os.Exit(1)
	}

	var target *java.Version

	// Interactive mode if no version specified
	if len(os.Args) < 3 {
		selected, err := selectJavaVersion(versions)
		if err != nil {
			fmt.Printf("Selection cancelled: %v\n", err)
			os.Exit(1)
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
			fmt.Printf("Java version '%s' not found.\n", version)
			fmt.Println("Use 'jv list' to see available versions.")
			os.Exit(1)
		}
	}

	// Confirm switch
	confirmed, err := confirmAction(
		fmt.Sprintf("Switch to Java %s?", target.Version),
		fmt.Sprintf("Path: %s", target.Path),
	)
	if err != nil || !confirmed {
		fmt.Println("Operation cancelled.")
		os.Exit(0)
	}

	fmt.Printf("Switching to Java %s...\n", target.Version)

	if err := env.SetJavaHome(target.Path); err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("\nNote: This command requires administrator privileges.")
		fmt.Println("Please run your terminal as Administrator and try again.")
		os.Exit(1)
	}

	fmt.Println(successStyle.Render("✓ Successfully updated JAVA_HOME!"))
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Faint(true).Render("Note: You may need to restart your terminal or applications for changes to take effect."))
}

func handleCurrent() {
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		fmt.Println("JAVA_HOME is not set.")
		return
	}

	detector := java.NewDetector()
	version := detector.GetVersion(javaHome)

	fmt.Printf("Current Java version: %s\n", version)
	fmt.Printf("JAVA_HOME: %s\n", javaHome)
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
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var pathToRemove string

	// Interactive mode if no path specified
	if len(os.Args) < 3 {
		if len(cfg.CustomPaths) == 0 {
			fmt.Println("No custom Java installations to remove.")
			fmt.Println("Use 'jv add <path>' to add one.")
			return
		}

		// Build options
		detector := java.NewDetector()
		options := make([]huh.Option[string], len(cfg.CustomPaths))
		for i, path := range cfg.CustomPaths {
			version := detector.GetVersion(path)
			label := fmt.Sprintf("%s - %s", version, path)
			options[i] = huh.NewOption(theme.CommandStyle.Render(label), path)
		}

		err := huh.NewSelect[string]().
			Title("Select Java Installation to Remove").
			Description("Use arrow keys to navigate, Enter to select").
			Options(options...).
			Value(&pathToRemove).
			Run()

		if err != nil {
			fmt.Printf("Selection cancelled: %v\n", err)
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
		fmt.Println("Operation cancelled.")
		return
	}

	cfg.RemoveCustomPath(pathToRemove)
	cfg.RemoveInstalledJDK(pathToRemove) // Also remove from installed JDKs if present

	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
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

	fmt.Printf("✓ Added search path: %s\n", path)
	fmt.Println("Run 'jv list' to see detected versions.")
}

func handleRemovePath() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var pathToRemove string

	// Interactive mode if no path specified
	if len(os.Args) < 3 {
		if len(cfg.SearchPaths) == 0 {
			fmt.Println("No custom search paths to remove.")
			fmt.Println("Use 'jv add-path <directory>' to add one.")
			return
		}

		// Build options
		options := make([]huh.Option[string], len(cfg.SearchPaths))
		for i, path := range cfg.SearchPaths {
			options[i] = huh.NewOption(theme.PathStyle.Render(path), path)
		}

		err := huh.NewSelect[string]().
			Title("Select Search Path to Remove").
			Description("Use arrow keys to navigate, Enter to select").
			Options(options...).
			Value(&pathToRemove).
			Run()

		if err != nil {
			fmt.Printf("Selection cancelled: %v\n", err)
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
		fmt.Println("Operation cancelled.")
		return
	}

	cfg.RemoveSearchPath(pathToRemove)
	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
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
		fmt.Printf("Error finding Java versions: %v\n", err)
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println("No Java installations found.")
		fmt.Println("Run 'jv install' to install Java.")
		os.Exit(1)
	}

	// Show interactive selector
	target, err := selectJavaVersion(versions)
	if err != nil {
		fmt.Printf("Selection cancelled: %v\n", err)
		os.Exit(1)
	}

	// Confirm switch
	confirmed, err := confirmAction(
		fmt.Sprintf("Switch to Java %s?", target.Version),
		fmt.Sprintf("Path: %s", target.Path),
	)
	if err != nil || !confirmed {
		fmt.Println("Operation cancelled.")
		os.Exit(0)
	}

	fmt.Printf("Switching to Java %s...\n", target.Version)

	if err := env.SetJavaHome(target.Path); err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("\nNote: This command requires administrator privileges.")
		fmt.Println("Please run your terminal as Administrator and try again.")
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

	// 1. Check JAVA_HOME
	fmt.Println("Checking JAVA_HOME...")
	currentJavaHome := os.Getenv("JAVA_HOME")
	detector := java.NewDetector()

	if currentJavaHome == "" {
		fmt.Println("  ✗ JAVA_HOME is not set")
		issues = append(issues, "JAVA_HOME is not set")
	} else {
		if detector.IsValidJavaPath(currentJavaHome) {
			fmt.Printf("  ✓ JAVA_HOME is set and valid: %s\n", currentJavaHome)
		} else {
			fmt.Printf("  ✗ JAVA_HOME is set but invalid: %s\n", currentJavaHome)
			issues = append(issues, fmt.Sprintf("JAVA_HOME points to invalid location: %s", currentJavaHome))
		}
	}
	fmt.Println()

	// 2. Check PATH
	fmt.Println("Checking PATH...")
	pathEnv := os.Getenv("Path")
	hasJavaHomeInPath := strings.Contains(pathEnv, "%JAVA_HOME%\\bin") || strings.Contains(pathEnv, "%JAVA_HOME%/bin")
	hasJavaInPath := strings.Contains(strings.ToLower(pathEnv), "java")

	if hasJavaHomeInPath {
		fmt.Println("  ✓ %JAVA_HOME%\\bin is in PATH")
	} else if hasJavaInPath {
		fmt.Println("  ⚠ PATH contains Java, but not via %JAVA_HOME%\\bin")
		warnings = append(warnings, "PATH contains Java paths, but %JAVA_HOME%\\bin is missing")
	} else {
		fmt.Println("  ✗ No Java found in PATH")
		issues = append(issues, "%JAVA_HOME%\\bin is not in PATH")
	}
	fmt.Println()

	// 3. Check Java installations
	fmt.Println("Checking Java installations...")
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Printf("  ✗ Error finding Java versions: %v\n", err)
		issues = append(issues, fmt.Sprintf("Error detecting Java installations: %v", err))
	} else if len(versions) == 0 {
		fmt.Println("  ⚠ No Java installations found")
		warnings = append(warnings, "No Java installations detected. Run 'jv install' to install Java.")
	} else {
		fmt.Printf("  ✓ Found %d Java installation(s)\n", len(versions))
		for _, v := range versions {
			source := "auto"
			if v.IsCustom {
				source = "custom"
			}
			marker := "  "
			if strings.EqualFold(v.Path, currentJavaHome) {
				marker = "  * "
			}
			fmt.Printf("%s  %s - %s (%s)\n", marker, v.Version, v.Path, source)
		}
	}
	fmt.Println()

	// 4. Check configuration file
	fmt.Println("Checking configuration...")
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
			fmt.Println("  ⚠ Configuration file does not exist (will be created when needed)")
		} else {
			fmt.Println("  ✓ Configuration file exists and is valid")
		}
		if len(cfg.CustomPaths) > 0 {
			fmt.Printf("  ✓ Custom paths configured: %d\n", len(cfg.CustomPaths))
		}
		if len(cfg.SearchPaths) > 0 {
			fmt.Printf("  ✓ Search paths configured: %d\n", len(cfg.SearchPaths))
		}
		if len(cfg.InstalledJDKs) > 0 {
			fmt.Printf("  ✓ Tracked JDKs: %d\n", len(cfg.InstalledJDKs))
		}
	}
	fmt.Println()

	// 5. Check administrator privileges
	fmt.Println("Checking privileges...")
	isAdmin := env.IsAdmin()
	if isAdmin {
		fmt.Println("  ✓ Running with administrator privileges")
	} else {
		fmt.Println("  ⚠ Not running as administrator (some operations require admin)")
		warnings = append(warnings, "Administrator privileges may be required for 'jv use' and 'jv repair'")
	}
	fmt.Println()

	// 6. Check if jv.exe is accessible
	fmt.Println("Checking jv tool...")
	if _, err := os.Executable(); err != nil {
		fmt.Println("  ⚠ Could not determine jv executable path")
	} else {
		fmt.Println("  ✓ jv tool is accessible")
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
		summaryContent += "\n" + theme.InfoMessage("Run 'jv repair' to fix issues")
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
	fmt.Println("Java Version Switcher - Auto Repair")
	fmt.Println("===================================")
	fmt.Println()

	isAdmin := env.IsAdmin()
	if !isAdmin {
		fmt.Println("⚠  Not running as Administrator")
		fmt.Println("   Some repairs require administrator privileges.")
		fmt.Println()
	}

	detector := java.NewDetector()

	// Step 1: Detect all issues
	fmt.Println("Scanning for issues...")
	fmt.Println()

	// Find Java installations first
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Printf("✗ Error finding Java versions: %v\n", err)
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println("✗ No Java installations found.")
		fmt.Println("Please install Java first: jv install")
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

	// Issue 2: PATH doesn't contain %JAVA_HOME%\bin
	pathEnv := os.Getenv("Path")
	hasJavaHomeInPath := strings.Contains(pathEnv, "%JAVA_HOME%\\bin") || strings.Contains(pathEnv, "%JAVA_HOME%/bin")
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
	fmt.Printf("Found %d issue(s):\n", len(issues))
	for i, issue := range issues {
		adminMarker := ""
		if issue.RequiresAdmin {
			adminMarker = " [Requires Admin]"
		}
		if !issue.CanFix {
			adminMarker += " [Cannot Fix]"
		}
		fmt.Printf("  %d. %s%s\n", i+1, issue.Description, adminMarker)
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
		fmt.Println("✗ No fixable issues (some require administrator privileges)")
		fmt.Println("  Run as administrator to fix all issues")
		os.Exit(1)
	}

	// Interactive selection of issues to fix
	options := make([]huh.Option[string], len(fixableIssues))
	for i, issue := range fixableIssues {
		options[i] = huh.NewOption(theme.WarningStyle.Render(issue.Description), issue.ID)
	}

	var selectedIssues []string
	err = huh.NewMultiSelect[string]().
		Title("Select Issues to Fix").
		Description("Use Space to select, Enter to confirm").
		Options(options...).
		Value(&selectedIssues).
		Run()

	if err != nil || len(selectedIssues) == 0 {
		fmt.Println("No issues selected. Repair cancelled.")
		return
	}

	// Perform repairs
	fmt.Println()
	fmt.Println("Performing repairs...")
	fmt.Println()

	repaired := []string{}
	for _, issueID := range selectedIssues {
		switch issueID {
		case "java_home_not_set", "java_home_invalid":
			// Let user select which Java to use
			target, err := selectJavaVersion(versions)
			if err != nil {
				fmt.Printf("✗ Skipped JAVA_HOME repair: %v\n", err)
				continue
			}

			if err := env.SetJavaHome(target.Path); err != nil {
				fmt.Printf("✗ Failed to set JAVA_HOME: %v\n", err)
				continue
			}

			repaired = append(repaired, fmt.Sprintf("Set JAVA_HOME to %s", target.Path))
			fmt.Printf("✓ JAVA_HOME set to Java %s\n", target.Version)

		case "path_missing_java_home":
			// This is handled by SetJavaHome above
			repaired = append(repaired, "Added %JAVA_HOME%\\bin to PATH")
			fmt.Println(successStyle.Render("✓ PATH updated"))

		case "config_error":
			cfg, err := config.Load()
			if err == nil {
				if err := cfg.Save(); err == nil {
					repaired = append(repaired, "Repaired configuration file")
					fmt.Println(successStyle.Render("✓ Configuration file repaired"))
				}
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("===================================")
	fmt.Println("Repair Complete")
	fmt.Println("===================================")
	fmt.Println()

	if len(repaired) == 0 {
		fmt.Println("✗ No repairs were successful")
		return
	}

	fmt.Println("Repairs performed:")
	for _, repair := range repaired {
		fmt.Printf("  ✓ %s\n", repair)
	}
	fmt.Println()
	fmt.Println("Note: You may need to restart your terminal for changes to take effect.")
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
	interactiveStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Italic(true)

	fmt.Println(categoryStyle.Render("INSTALLATION & SETUP"))
	fmt.Printf("  %s  %s %s\n",
		commandStyle.Render("install"),
		descStyle.Render("Install Java from open-source distributors"),
		interactiveStyle.Render("(interactive)"))
	fmt.Printf("  %s   %s %s\n",
		commandStyle.Render("doctor"),
		descStyle.Render("Run diagnostics on your Java environment"),
		interactiveStyle.Render("(styled)"))
	fmt.Printf("  %s   %s %s\n",
		commandStyle.Render("repair"),
		descStyle.Render("Automatically fix configuration issues"),
		interactiveStyle.Render("(interactive)"))
	fmt.Println()

	fmt.Println(categoryStyle.Render("VERSION MANAGEMENT"))
	fmt.Printf("  %s     %s %s\n",
		commandStyle.Render("list"),
		descStyle.Render("List all available Java versions"),
		interactiveStyle.Render("(styled)"))
	fmt.Printf("  %s [version]  %s %s\n",
		commandStyle.Render("use"),
		descStyle.Render("Switch to Java version"),
		interactiveStyle.Render("(interactive if no version)"))
	fmt.Printf("  %s   %s %s\n",
		commandStyle.Render("switch"),
		descStyle.Render("Quick interactive version switcher"),
		interactiveStyle.Render("(always interactive)"))
	fmt.Printf("  %s  %s\n",
		commandStyle.Render("current"),
		descStyle.Render("Show current Java version"))
	fmt.Println()

	fmt.Println(categoryStyle.Render("CUSTOM INSTALLATIONS"))
	fmt.Printf("  %s <path>     %s\n",
		commandStyle.Render("add"),
		descStyle.Render("Add a specific Java installation"))
	fmt.Printf("  %s [path]  %s %s\n",
		commandStyle.Render("remove"),
		descStyle.Render("Remove a custom installation"),
		interactiveStyle.Render("(interactive if no path)"))
	fmt.Println()

	fmt.Println(categoryStyle.Render("SEARCH PATHS"))
	fmt.Printf("  %s <dir>  %s\n",
		commandStyle.Render("add-path"),
		descStyle.Render("Add directory to scan for Java installations"))
	fmt.Printf("  %s [dir]  %s %s\n",
		commandStyle.Render("remove-path"),
		descStyle.Render("Remove directory from search paths"),
		interactiveStyle.Render("(interactive if no dir)"))
	fmt.Printf("  %s   %s\n",
		commandStyle.Render("list-paths"),
		descStyle.Render("Show all search paths (standard + custom)"))
	fmt.Println()

	fmt.Println(categoryStyle.Render("OTHER"))
	fmt.Printf("  %s  %s\n",
		commandStyle.Render("version"),
		descStyle.Render("Show version information"))
	fmt.Printf("  %s     %s\n",
		commandStyle.Render("help"),
		descStyle.Render("Show this help message"))
	fmt.Println()

	// Examples section
	fmt.Println(theme.Title.Render("EXAMPLES"))
	fmt.Println("  " + theme.Code.Render("jv list") + "                         # List Java versions")
	fmt.Println("  " + theme.Code.Render("jv switch") + "                       # Interactive switcher")
	fmt.Println("  " + theme.Code.Render("jv use 17") + "                      # Switch to Java 17")
	fmt.Println("  " + theme.Code.Render("jv install") + "                     # Install Java interactively")
	fmt.Println("  " + theme.Code.Render("jv add C:\\custom\\jdk-21") + "        # Add custom installation")
	fmt.Println("  " + theme.Code.Render("jv doctor") + "                      # Check system health")
	fmt.Println()

	// Autocomplete note
	fmt.Println(theme.InfoStyle.Italic(true).Render("💡 Tip: PowerShell autocomplete is installed automatically by install.ps1"))

	fmt.Println()

	// Note section with theme
	note := theme.WarningBox.Render("⚠ Administrator privileges required for: use, switch, install, repair")
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

	// Build options
	options := make([]huh.Option[int], len(versions))
	for i, v := range versions {
		label := fmt.Sprintf("%-15s %s", v.Version, v.Path)

		// Add scope info
		if scope, found := scopeMap[v.Path]; found {
			switch scope {
			case "system":
				label += " (system-wide)"
			case "user":
				label += " (user-only)"
			}
		} else if v.IsCustom {
			label += " (custom)"
		} else {
			label += " (auto)"
		}

		options[i] = huh.NewOption(theme.CommandStyle.Render(label), i)
	}

	var selectedIdx int

	err := huh.NewSelect[int]().
		Title("Select Java Version").
		Description("Use arrow keys to navigate, Enter to select").
		Options(options...).
		Value(&selectedIdx).
		Run()

	if err != nil {
		return nil, err
	}

	return &versions[selectedIdx], nil
}

// confirmAction shows a confirmation prompt
func confirmAction(title, description string) (bool, error) {
	var confirmed bool

	err := huh.NewConfirm().
		Title(title).
		Description(description).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed).
		Run()

	return confirmed, err
}
