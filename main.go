package main

import (
	"fmt"
	"os"
	"strings"

	"jv/internal/config"
	"jv/internal/env"
	"jv/internal/installer"
	"jv/internal/java"
)

// Version is set during build time via ldflags
var Version = "dev"

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
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Printf("Error finding Java versions: %v\n", err)
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println("No Java installations found.")
		return
	}

	current := os.Getenv("JAVA_HOME")
	fmt.Println("Available Java versions:")
	fmt.Println()

	for _, v := range versions {
		marker := "  "
		if strings.EqualFold(v.Path, current) {
			marker = "* "
		}
		source := "auto"
		if v.IsCustom {
			source = "custom"
		}
		fmt.Printf("%s%-15s %s (%s)\n", marker, v.Version, v.Path, source)
	}
}

func handleUse() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: jv use <version>")
		fmt.Println("Example: jv use 17")
		os.Exit(1)
	}

	version := os.Args[2]

	detector := java.NewDetector()
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Printf("Error finding Java versions: %v\n", err)
		os.Exit(1)
	}

	var target *java.Version
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

	fmt.Printf("Switching to Java %s...\n", target.Version)
	fmt.Printf("Path: %s\n", target.Path)

	if err := env.SetJavaHome(target.Path); err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("\nNote: This command requires administrator privileges.")
		fmt.Println("Please run your terminal as Administrator and try again.")
		os.Exit(1)
	}

	fmt.Println("Successfully updated JAVA_HOME!")
	fmt.Println("\nNote: You may need to restart your terminal or applications for changes to take effect.")
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
		fmt.Println("Usage: jv add <path>")
		fmt.Println("Example: jv add C:\\custom\\jdk-21")
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
		fmt.Println("This path is already in the custom paths list.")
		return
	}

	cfg.AddCustomPath(path)
	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	version := detector.GetVersion(path)
	fmt.Printf("Added Java %s to custom paths.\n", version)
}

func handleRemove() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: jv remove <path>")
		fmt.Println("Example: jv remove C:\\custom\\jdk-21")
		os.Exit(1)
	}

	path := os.Args[2]

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if !cfg.HasCustomPath(path) {
		fmt.Println("This path is not in the custom paths list.")
		return
	}

	cfg.RemoveCustomPath(path)
	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Removed from custom paths.")
}

func handleAddPath() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: jv add-path <directory>")
		fmt.Println("Example: jv add-path C:\\DevTools\\Java")
		fmt.Println()
		fmt.Println("This adds a directory where the detector will search for Java installations.")
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
		fmt.Println("This search path is already configured.")
		return
	}

	cfg.AddSearchPath(path)
	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Added search path: %s\n", path)
	fmt.Println("The detector will now scan this directory for Java installations.")
	fmt.Println("Run 'jv list' to see detected versions.")
}

func handleRemovePath() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: jv remove-path <directory>")
		fmt.Println("Example: jv remove-path C:\\DevTools\\Java")
		os.Exit(1)
	}

	path := os.Args[2]

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if !cfg.HasSearchPath(path) {
		fmt.Println("This path is not in the search paths list.")
		return
	}

	cfg.RemoveSearchPath(path)
	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Removed search path.")
}

func handleListPaths() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	detector := java.NewDetector()

	fmt.Println("Java search paths:")
	fmt.Println()

	// Show standard paths
	fmt.Println("Standard paths (built-in):")
	standardPaths := []string{
		"C:\\Program Files\\Java",
		"C:\\Program Files (x86)\\Java",
		"C:\\Program Files\\Eclipse Adoptium",
		"C:\\Program Files\\Eclipse Foundation",
		"C:\\Program Files\\Zulu",
		"C:\\Program Files\\Amazon Corretto",
		"C:\\Program Files\\Microsoft",
	}

	for _, p := range standardPaths {
		exists := detector.IsValidSearchPath(p)
		marker := ""
		if exists {
			marker = " [exists]"
		}
		fmt.Printf("  %s%s\n", p, marker)
	}

	// Show custom paths
	if len(cfg.SearchPaths) > 0 {
		fmt.Println()
		fmt.Println("Custom search paths:")
		for _, p := range cfg.SearchPaths {
			exists := detector.IsValidSearchPath(p)
			marker := ""
			if exists {
				marker = " [exists]"
			} else {
				marker = " [not found]"
			}
			fmt.Printf("  %s%s\n", p, marker)
		}
	} else {
		fmt.Println()
		fmt.Println("No custom search paths configured.")
		fmt.Println("Use 'jv add-path <directory>' to add one.")
	}
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

func handleDoctor() {
	fmt.Println("Java Version Switcher - System Diagnostics")
	fmt.Println("==========================================")
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
	fmt.Println("==========================================")
	fmt.Println("Diagnostics Summary")
	fmt.Println("==========================================")
	fmt.Println()

	if len(issues) == 0 && len(warnings) == 0 {
		fmt.Println("✓ All checks passed! Your Java environment is properly configured.")
		return
	}

	if len(issues) > 0 {
		fmt.Println("Issues found:")
		for _, issue := range issues {
			fmt.Printf("  ✗ %s\n", issue)
		}
		fmt.Println()
	}

	if len(warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
		fmt.Println()
	}

	if len(issues) > 0 {
		fmt.Println("To fix issues, run: jv repair")
		fmt.Println("(Note: repair requires administrator privileges for system changes)")
	}
}

func handleRepair() {
	fmt.Println("Java Version Switcher - Auto Repair")
	fmt.Println("===================================")
	fmt.Println()

	isAdmin := env.IsAdmin()
	if !isAdmin {
		fmt.Println("⚠  Not running as Administrator")
		fmt.Println("   Some repairs require administrator privileges.")
		fmt.Println("   Continuing with user-level repairs...")
		fmt.Println()
	}

	detector := java.NewDetector()
	repairs := []string{}

	// 1. Find Java installations
	fmt.Println("Scanning for Java installations...")
	versions, err := detector.FindAll()
	if err != nil {
		fmt.Printf("✗ Error finding Java versions: %v\n", err)
		fmt.Println()
		fmt.Println("Cannot proceed with repair - unable to detect Java installations.")
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println("✗ No Java installations found.")
		fmt.Println()
		fmt.Println("Please install Java first:")
		fmt.Println("  jv install")
		os.Exit(1)
	}

	fmt.Printf("✓ Found %d Java installation(s)\n", len(versions))
	fmt.Println()

	// 2. Check and fix JAVA_HOME
	fmt.Println("Checking JAVA_HOME...")
	currentJavaHome := os.Getenv("JAVA_HOME")
	javaHomeNeedsFix := false

	if currentJavaHome == "" {
		fmt.Println("  ✗ JAVA_HOME is not set")
		javaHomeNeedsFix = true
	} else if !detector.IsValidJavaPath(currentJavaHome) {
		fmt.Printf("  ✗ JAVA_HOME is invalid: %s\n", currentJavaHome)
		javaHomeNeedsFix = true
	} else {
		fmt.Printf("  ✓ JAVA_HOME is valid: %s\n", currentJavaHome)
	}

	if javaHomeNeedsFix {
		// Use the first available Java installation
		targetJava := versions[0]
		fmt.Printf("  → Setting JAVA_HOME to: %s\n", targetJava.Path)

		if !isAdmin {
			fmt.Println()
			fmt.Println("✗ Cannot set JAVA_HOME - requires administrator privileges")
			fmt.Println("  Please run as administrator and try again:")
			fmt.Println("    1. Right-click CMD/PowerShell")
			fmt.Println("    2. Select 'Run as administrator'")
			fmt.Println("    3. Run: jv repair")
			fmt.Println()
			fmt.Println("Or manually set JAVA_HOME to:")
			fmt.Printf("    %s\n", targetJava.Path)
			os.Exit(1)
		}

		if err := env.SetJavaHome(targetJava.Path); err != nil {
			fmt.Printf("✗ Failed to set JAVA_HOME: %v\n", err)
			os.Exit(1)
		}

		repairs = append(repairs, fmt.Sprintf("Set JAVA_HOME to %s", targetJava.Path))
		repairs = append(repairs, "Added %JAVA_HOME%\\bin to PATH")
		fmt.Println("  ✓ JAVA_HOME set successfully")
	}
	fmt.Println()

	// 3. Check and fix PATH
	fmt.Println("Checking PATH...")
	pathEnv := os.Getenv("Path")
	hasJavaHomeInPath := strings.Contains(pathEnv, "%JAVA_HOME%\\bin") || strings.Contains(pathEnv, "%JAVA_HOME%/bin")

	if !hasJavaHomeInPath {
		if !isAdmin {
			fmt.Println("  ✗ %JAVA_HOME%\\bin is not in PATH")
			fmt.Println("     (Cannot fix - requires administrator privileges)")
		} else {
			// JAVA_HOME was just set, so PATH should already be updated by SetJavaHome
			fmt.Println("  ✓ PATH will be updated when JAVA_HOME is set")
		}
	} else {
		fmt.Println("  ✓ %JAVA_HOME%\\bin is in PATH")
	}
	fmt.Println()

	// 4. Verify/repair configuration
	fmt.Println("Checking configuration...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("  ⚠ Error loading config: %v\n", err)
		fmt.Println("  → Attempting to fix configuration...")
		// Load() returns a new empty config if file doesn't exist, so if it failed
		// the file might be corrupted. Try loading again - if it fails again,
		// the issue is with file permissions or disk, not config format
		cfg2, err2 := config.Load()
		if err2 != nil {
			fmt.Printf("  ✗ Cannot fix configuration: %v\n", err2)
		} else {
			// Try to save to create/fix the file
			if err := cfg2.Save(); err != nil {
				fmt.Printf("  ✗ Failed to save config: %v\n", err)
			} else {
				repairs = append(repairs, "Repaired configuration file")
				fmt.Println("  ✓ Configuration file repaired")
			}
		}
	} else {
		// Config loaded successfully, verify it's saved properly
		if err := cfg.Save(); err != nil {
			fmt.Printf("  ⚠ Configuration read but could not verify write: %v\n", err)
		} else {
			fmt.Println("  ✓ Configuration file is valid")
		}
	}
	fmt.Println()

	// Summary
	fmt.Println("===================================")
	fmt.Println("Repair Summary")
	fmt.Println("===================================")
	fmt.Println()

	if len(repairs) == 0 {
		fmt.Println("✓ No repairs needed - your environment is properly configured!")
		return
	}

	fmt.Println("Repairs performed:")
	for _, repair := range repairs {
		fmt.Printf("  ✓ %s\n", repair)
	}
	fmt.Println()
	fmt.Println("Note: You may need to restart your terminal for changes to take effect.")
}

func printVersion() {
	fmt.Printf("Java Version Switcher (jv) version %s\n", Version)
	fmt.Println("https://github.com/CostaBrosky/jv")
}

func printUsage() {
	fmt.Println("Java Version Switcher - Easy Java version management for Windows")
	fmt.Println()
	fmt.Println("Usage: jv <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  Version Management:")
	fmt.Println("    list              List all available Java versions")
	fmt.Println("    use <version>     Switch to specified Java version")
	fmt.Println("    current           Show current Java version")
	fmt.Println()
	fmt.Println("  Custom Installations:")
	fmt.Println("    add <path>        Add a specific Java installation")
	fmt.Println("    remove <path>     Remove a custom installation")
	fmt.Println()
	fmt.Println("  Search Paths:")
	fmt.Println("    add-path <dir>    Add directory to search for Java installations")
	fmt.Println("    remove-path <dir> Remove directory from search paths")
	fmt.Println("    list-paths        Show all search paths (standard + custom)")
	fmt.Println()
	fmt.Println("  Setup & Maintenance:")
	fmt.Println("    install           Install Java from open-source distributors")
	fmt.Println("    doctor            Run diagnostics on your Java environment")
	fmt.Println("    repair            Automatically fix configuration issues")
	fmt.Println("    init              [Deprecated] Use 'jv install' instead")
	fmt.Println()
	fmt.Println("  Other:")
	fmt.Println("    version           Show version information")
	fmt.Println("    help              Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  jv list")
	fmt.Println("  jv use 17")
	fmt.Println("  jv current")
	fmt.Println("  jv add C:\\custom\\jdk-21")
	fmt.Println("  jv add-path C:\\DevTools\\Java")
	fmt.Println("  jv list-paths")
	fmt.Println()
	fmt.Println("Note: Switching Java versions requires administrator privileges.")
}
