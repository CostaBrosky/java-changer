# Quick Start Guide

## Prerequisites

1. **Install Go** (if you don't have it):
   - Download from: https://go.dev/dl/
   - Install and verify: `go version`

## Build & Installation (3 steps)

### 1. Build the project

```bash
# Double-click on build.bat
# OR from terminal:
build.bat
```

### 2. Install the tool

Create a dedicated directory and add it to PATH:

```cmd
# Create a directory for your tools
mkdir C:\tools
copy jv.exe C:\tools\

# Add C:\tools to PATH:
# 1. Search for "environment variables" in Start menu
# 2. Click "Edit environment variables for your account"
# 3. Select "Path" and click "Edit"
# 4. Click "New" and add: C:\tools
# 5. Click OK on all windows
# 6. Restart your terminal
```

### 3. Use the tool

```bash
# List available Java versions
jv list

# Switch to Java 17 (RUN AS ADMINISTRATOR!)
jv use 17

# Verify current version
jv current
java -version
```

## Main Commands

```bash
# Installation & Setup
jv install                      # Install Java from open-source distributors
jv init                         # Initialize/repair Java environment variables

# Version management
jv list                         # List all available versions
jv use 17                       # Switch to Java 17
jv current                      # Show current version

# Custom installations (specific)
jv add C:\my\jdk-17             # Add ONE specific installation

# Search paths (directory scanning)
jv add-path C:\DevTools\Java    # Add directory to scan
jv list-paths                   # Show all search paths
jv remove-path C:\DevTools\Java # Remove search path

# Help
jv help                         # Show complete help
```

**Difference between `add` and `add-path`:**
- `add`: For a single installation (e.g., `C:\custom\jdk-17`)
- `add-path`: For a directory containing multiple versions (e.g., `C:\DevTools\Java` containing jdk-17, jdk-21, etc.)

## IMPORTANT: Administrator Privileges

The `jv use` command requires administrator privileges because it modifies system environment variables.

**How to run as administrator:**
- Right-click "CMD" or "PowerShell" in Start menu
- Select "Run as administrator"
- Run `jv use <version>`

## Verify Installation

After installing jv.exe:

```bash
# Open a NEW terminal and try:
jv help
```

If you see the help message, installation succeeded!

## Common Issues

**"jv is not recognized..."**
- Make sure you added jv.exe to PATH
- Restart the terminal after modifying PATH

**"failed to open registry key"**
- Run the terminal as administrator

**"No Java installations found"**
- Install Java: `jv install`
- Or manually add: `jv add C:\path\to\jdk`

---

For complete details, see [README.md](README.md)
