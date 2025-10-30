# Java Version Switcher (jv)

> A simple and fast CLI tool to switch Java versions on Windows with a single command.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Windows](https://img.shields.io/badge/Platform-Windows-0078D6?logo=windows)](https://www.microsoft.com/windows)

## Features

- 🎯 **Beautiful Interactive UI** - Navigate with arrow keys, powered by Huh! and Lip Gloss
- 📊 **Animated Progress Bars** - Real-time download progress with speed indicator
- 📦 **Batch Installation** - Install multiple Java versions at once
- 🔍 **Auto-detection** of Java installations from multiple vendors
- 🔄 **Permanent switching** - Modifies system environment variables (JAVA_HOME and PATH)
- 🎨 **Colored Output** - Styled terminal output with visual hierarchy
- ⌨️ **Shell Autocomplete** - Tab completion for PowerShell, Bash, and Fish
- ✅ **Confirmation Prompts** - Prevents accidental operations
- 💾 **Persistent configuration** - Tracks installations and preferences
- 🚀 **Zero dependencies** - Standalone executable
- 🌐 **All distributions supported** - Oracle JDK, OpenJDK, Adoptium, Zulu, Corretto, Microsoft

## Table of Contents

- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Commands](#-commands)
- [Examples](#-examples)
- [How It Works](#-how-it-works)
- [Configuration](#-configuration)
- [FAQ](#-faq)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)
- [License](#-license)

## Prerequisites

- **Operating System**: Windows 10 or Windows 11
- **Go**: 1.21+ (only for building from source)
- **Privileges**: Administrator (to modify system environment variables)

## Installation

### Method 1: Download Executable (Recommended)

1. Download the latest `jv.exe` from the [Releases](https://github.com/CostaBrosky/jv/releases) page
2. Create a dedicated directory for your tools (e.g., `C:\tools\`)
3. Copy `jv.exe` to that directory
4. Add the directory to your PATH (see instructions below)
5. Open a new terminal and verify: `jv version`

### Method 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/CostaBrosky/jv.git
cd jv

# Download dependencies
go mod download

# Build
go build -ldflags="-s -w" -o jv.exe .

# Copy to your tools directory
copy jv.exe C:\tools\
```

### Adding to PATH

1. Press `Win + X` → Select "System"
2. Click "Advanced system settings"
3. Click "Environment Variables"
4. Under "User variables" (or "System variables"), select "Path" → "Edit"
5. Click "New" and add the path to your tools directory (e.g., `C:\tools`)
6. Click "OK" on all windows
7. **Restart your terminal**

## Windows SmartScreen & Security Warning

### Why does Windows block this executable?

When you download `jv.exe`, **Windows SmartScreen will show a security warning** because the executable is **not digitally signed**. This is normal and expected for open-source tools without a code signing certificate (which costs $200-500/year).

**This does NOT mean the file is malicious** - it simply means Windows doesn't recognize the publisher.

### How to download and use jv.exe safely

#### Step 1: Download the file

1. Download `jv.exe` from the [Releases](https://github.com/CostaBrosky/jv/releases) page
2. Windows SmartScreen may show: **"Windows protected your PC"**
3. Click **"More info"**
4. Click **"Run anyway"**

#### Step 2: If Windows Defender quarantines the file

Windows Defender may automatically move `jv.exe` to quarantine. To restore it:

1. Open **Windows Security** (search in Start menu)
2. Go to **"Virus & threat protection"**
3. Click **"Protection history"**
4. Find the `jv.exe` entry
5. Click **"Actions"** → **"Restore"**
6. Confirm the action

#### Step 3: Add a permanent exception (recommended)

To prevent Windows Defender from blocking `jv.exe` in the future:

1. Open **Windows Security**
2. Go to **"Virus & threat protection"**
3. Click **"Manage settings"** under "Virus & threat protection settings"
4. Scroll down to **"Exclusions"**
5. Click **"Add or remove exclusions"**
6. Click **"Add an exclusion"** → **"File"**
7. Navigate to and select `jv.exe` (e.g., `C:\tools\jv.exe`)
8. Click **"Open"**

Now Windows Defender will not scan or quarantine this file.

#### Step 4: Verify file authenticity (optional but recommended)

To ensure the file hasn't been tampered with, verify the SHA256 checksum:

```powershell
# In PowerShell, navigate to where you downloaded jv.exe
cd C:\tools

# Calculate the checksum
Get-FileHash jv.exe -Algorithm SHA256

# Compare the output with checksums.txt from the GitHub release
```

The checksum should match exactly with the one in `checksums.txt` from the release.

### Why don't you sign the executable?

Code signing certificates for Windows cost $200-500 per year and require identity verification. For a free, open-source tool, this cost is not justified. Instead:

- All source code is **publicly available** on GitHub for audit
- **SHA256 checksums** are provided to verify file integrity
- The build process is **transparent** (you can build from source yourself)
- The project is **open source** under MIT license

If you're concerned about security, you can always **build from source** (see Method 2 in Installation).

## Quick Start

```bash
# 1. List all available Java versions (beautiful colored output!)
jv list

# 2. If no Java found, install Java (interactive with arrows!)
jv install

# 3. Switch to a version interactively (RUN AS ADMINISTRATOR!)
jv switch
# or
jv use

# 4. Or switch directly:
jv use 17

# 5. Verify current version
jv current
java -version

# 6. Check system health
jv doctor

# 7. Fix issues interactively
jv repair
```

**IMPORTANT**: The `jv use`, `jv switch`, `jv install`, and `jv repair` commands require administrator privileges for system-wide configuration. Right-click on CMD/PowerShell → "Run as administrator"

## Commands

### Installation & Setup

| Command | Description | Example |
|---------|-------------|---------|
| `jv install` | Install Java from open-source distributors (Adoptium, etc.) | `jv install` |
| `jv doctor` | Run diagnostics on your Java environment | `jv doctor` |
| `jv repair` | Automatically fix configuration issues | `jv repair` |

### Version Management

| Command | Description | Example |
|---------|-------------|---------|
| `jv list` | List all available Java versions (with colors!) | `jv list` |
| `jv use [version]` | Switch to specified version (interactive if no version) | `jv use 17` or `jv use` |
| `jv switch` | Quick interactive version switcher | `jv switch` |
| `jv current` | Show current Java version | `jv current` |

### Custom Installations

| Command | Description | Example |
|---------|-------------|---------|
| `jv add <path>` | Add a specific Java installation (with confirmation) | `jv add C:\custom\jdk-21` |
| `jv remove [path]` | Remove a custom installation (interactive if no path) | `jv remove` or `jv remove C:\custom\jdk-21` |

### Search Paths

| Command | Description | Example |
|---------|-------------|---------|
| `jv add-path <dir>` | Add directory to scan for Java installations (with confirmation) | `jv add-path C:\DevTools\Java` |
| `jv remove-path [dir]` | Remove a search path (interactive if no dir) | `jv remove-path` or `jv remove-path C:\DevTools\Java` |
| `jv list-paths` | Show all search paths | `jv list-paths` |

### Utilities

| Command | Description | Example |
|---------|-------------|---------|
| `jv version` | Show jv version | `jv version` |
| `jv help` | Show help message | `jv help` |

## Examples

All commands now feature beautiful interactive UIs powered by Huh! and Lip Gloss! 🎨

### Scenario 0: Installing Java (Interactive!)

jv provides a beautiful, interactive installation experience powered by Huh!:

```bash
# Interactive installation with arrow key navigation
jv install
```

**Features:**
- 🎯 **Navigate with arrow keys** - No more typing numbers!
- ✨ **Visual selection** - See what you're choosing
- 📦 **Batch installation** - Install multiple Java versions at once
- ✅ **Smart detection** - Shows which versions are already installed
- 🎨 **Beautiful UI** - Modern terminal experience

**Installation Flow:**

1. **Select Distributor** (currently Eclipse Adoptium)
2. **Choose Mode:**
   - Install single version
   - Install multiple versions (batch)
3. **Select Version(s):**
   - Single: Use ↑↓ arrows, press Enter
   - Multi: Use Space to select, Enter to confirm
4. **Choose Scope:**
   - System-wide (admin) - `C:\Program Files\...`
   - User-only - `%USERPROFILE%\.jv\...`
5. **Auto-download and install**

**Example - Single Version:**
```bash
jv install
→ Select "Install single version"
→ Navigate to "Java 21 [LTS]"
→ Press Enter
→ Select scope
→ Done!
```

**Example - Multiple Versions:**
```bash
jv install
→ Select "Install multiple versions"
→ Press Space on "Java 21 [LTS]"
→ Press Space on "Java 17 [LTS]"
→ Press Enter to confirm
→ Select scope
→ Both versions install automatically!
```

**Note**: Currently supports Eclipse Adoptium (Temurin). More distributors coming soon!

### Interactive UI Examples

**Quick Version Switcher:**
```bash
$ jv switch
```
```
┌─────────────────────────────────────────────────────────────────┐
│ Select Java Version                                             │
│ Use arrow keys to navigate, Enter to select                     │
│                                                                 │
│   ○ 21.0.8         C:\Program Files\...\jdk-21 (system-wide)    │
│ > ● 17.0.12        C:\Program Files\...\jdk-17 (auto)           │
│   ○ 11.0.24        C:\Users\...\jdk-11 (user-only)              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Interactive Removal:**
```bash
$ jv remove
```
```
┌─────────────────────────────────────────────────────────────────┐
│ Select Java Installation to Remove                              │
│ Use arrow keys to navigate, Enter to select                     │
│                                                                 │
│ > ● 17.0.12 - C:\custom\jdk-17                                  │
│   ○ 11.0.24 - C:\custom\jdk-11                                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Remove Java 17.0.12?                                            │
│ Path: C:\custom\jdk-17                                          │
│                                                                 │
│ > ● Yes   ○ No                                                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Health Check with Styled Output:**
```bash
$ jv doctor
```
```
Java Version Switcher - System Diagnostics
═══════════════════════════════════════════

Checking JAVA_HOME...
  ✓ JAVA_HOME is set and valid: C:\Program Files\Eclipse Adoptium\jdk-21

Checking PATH...
  ✓ %JAVA_HOME%\bin is in PATH

...

╭─────────────────────────────────────────────────────────────────╮
│ ✓ All checks passed!                                            │
│                                                                 │
│ Your Java environment is properly configured.                   │
╰─────────────────────────────────────────────────────────────────╯
```

**Download with Animated Progress Bar:**
```bash
$ jv install
# After selecting version...
```
```
Downloading JDK...

[=========================================>           ] 75%

204.5 MB / 273.2 MB (75%) - 12.34 MB/s
```

### Shell Autocomplete

**PowerShell autocomplete is installed automatically!** 🎉

When you install jv using `install.ps1`, the autocomplete is configured automatically in your PowerShell profile. Just restart PowerShell and it works!

```powershell
# After installing jv, restart PowerShell and try:
jv <tab>          # Shows all commands
jv use <tab>      # Shows all Java versions
jv add C:\<tab>   # Path completion
```

**Skip autocomplete during install:**
```powershell
# If you don't want autocomplete
.\install.ps1 -NoCompletion
```

**Note:** Autocomplete is currently Windows/PowerShell only. For WSL/Git Bash users, the completion script is embedded in `install.ps1` and can be extracted if needed.

**What Gets Autocompleted:**
- ✅ All jv commands with descriptions
- ✅ Java versions for `jv use <tab>`
- ✅ Directory paths for `jv add <tab>` and `jv add-path <tab>`
- ✅ Custom paths for `jv remove <tab>`
- ✅ Search paths for `jv remove-path <tab>`

**How it works:** The `install.ps1` script embeds a PowerShell completion script directly in your `$PROFILE`. No need to run any commands - just restart PowerShell!

### Scenario 1: Switching between versions (New Interactive Mode!)

```bash
# List available versions (with beautiful colors!)
jv list

# Output:
# Available Java Versions:
#
# → 21.0.8         C:\Program Files\Eclipse Adoptium\jdk-21 (system-wide)
#   17.0.1         C:\Program Files\Java\jdk-17 (auto)
#   11.0.12        C:\Users\username\.jv\jdk-11 (user-only)

# Switch interactively (navigate with arrows!)
jv switch
# or
jv use

# Or switch directly (classic mode)
jv use 11

# Verify
jv current
java -version
```

### Scenario 2: Adding a custom directory

If you have Java in a non-standard directory (e.g., `C:\DevTools\Java\` with multiple versions):

```bash
# Add the base directory
jv add-path C:\DevTools\Java

# The detector will automatically find all versions in that directory
jv list

# Output:
# Available Java versions:
#
#   17.0.1          C:\Program Files\Java\jdk-17 (auto)
#   21.0.0          C:\DevTools\Java\jdk-21 (auto)
#   19.0.2          C:\DevTools\Java\jdk-19 (auto)
```

### Scenario 3: Adding a specific installation

```bash
# Add ONE specific installation
jv add D:\Projects\special-jdk-17

# Use that version
jv use special
```

### Difference between `add` and `add-path`

**`jv add <path>`**: Add ONE specific Java installation
```bash
jv add C:\custom\jdk-17
# Adds ONLY this installation
```

**`jv add-path <directory>`**: Scan a directory for ALL Java installations
```bash
jv add-path C:\DevTools\Java
# If it contains jdk-17, jdk-19, jdk-21, finds all three
```

## How It Works

### 1. Auto-detection

The tool automatically scans these standard directories:

```
C:\Program Files\Java
C:\Program Files (x86)\Java
C:\Program Files\Eclipse Adoptium
C:\Program Files\Eclipse Foundation
C:\Program Files\Zulu
C:\Program Files\Amazon Corretto
C:\Program Files\Microsoft
C:\DevTools\Java
```

Plus any custom search paths added with `jv add-path`.

### 2. Persistent Configuration

Configuration is saved in `%USERPROFILE%\jv.json` (JSON file):

```json
{
  "custom_paths": [
    "C:\\custom\\jdk-special"
  ],
  "search_paths": [
    "C:\\DevTools\\Java",
    "D:\\Work\\java-installations"
  ]
}
```

### 3. Environment Variable Modification

When you run `jv use <version>`, the tool:

1. Modifies `JAVA_HOME` in the system Registry
2. Updates `PATH`:
   - Removes old Java references (e.g., old `%JAVA_HOME%\bin`)
   - Adds `%JAVA_HOME%\bin` at the beginning of PATH
3. Broadcasts `WM_SETTINGCHANGE` to notify Windows of the changes

**Technical details**:
- Uses Windows Registry API (`HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\Environment`)
- Requires administrator privileges to write to system Registry
- Changes are permanent and survive reboots

### 4. Version Detection

The tool identifies Java versions in two ways:

1. **Runs `java -version`** and parses the output
2. **Fallback**: extracts from directory name (e.g., `jdk-17`, `jdk1.8.0_322`)

## Configuration

### Configuration File

Location: `%USERPROFILE%\jv.json`

Example:
```json
{
  "custom_paths": [
    "C:\\MyJava\\jdk-17-custom",
    "D:\\Projects\\special-jdk"
  ],
  "search_paths": [
    "C:\\DevTools\\Java",
    "D:\\JavaInstalls"
  ]
}
```

### Manual Editing (Advanced)

You can manually edit the `jv.json` file with a text editor, then run `jv list` to see the changes.

## FAQ

<details>
<summary><b>Do I always need to run as administrator?</b></summary>

No, only the `jv use` command requires administrator privileges because it modifies system environment variables. Other commands (`list`, `current`, `add-path`, etc.) work normally.
</details>

<details>
<summary><b>Are the changes permanent?</b></summary>

Yes! `jv use` permanently modifies system environment variables. Changes survive reboots and are visible to all applications.
</details>

<details>
<summary><b>Does it work with all Java distributions?</b></summary>

Yes! It works with:
- Oracle JDK
- OpenJDK
- Eclipse Adoptium (Temurin)
- Azul Zulu
- Amazon Corretto
- Microsoft OpenJDK
- Liberica JDK
- Any other distribution with the standard `bin/java.exe` structure
</details>

<details>
<summary><b>Can I use it with Java 8, 11, 17, 21?</b></summary>

Yes, all Java versions are supported (from Java 1.6 onwards).
</details>

<details>
<summary><b>What happens to PATH when I switch versions?</b></summary>

The tool:
1. Automatically removes old Java paths from PATH
2. Adds `%JAVA_HOME%\bin` at the beginning of PATH
3. This ensures the correct version is always used
</details>

## 🔧 Troubleshooting

### "jv is not recognized as a command"

**Cause**: `jv.exe` is not in PATH

**Solution**:
```bash
# Check where jv.exe is located
where jv

# If not found, add it to PATH (see Installation section)
```

### "failed to open registry key (run as administrator)"

**Cause**: You're running `jv use` without administrator privileges

**Solution**:
1. Right-click "CMD" or "PowerShell"
2. Select "Run as administrator"
3. Re-run the command

### "No Java installations found"

**Cause**: Java is not in a standard directory or not installed

**Solution**:
```bash
# Add the directory where you installed Java
jv add-path C:\path\to\java\directory

# Or add the specific installation
jv add C:\path\to\jdk
```

### Changes don't apply immediately

**Cause**: The terminal or applications haven't reloaded environment variables

**Solution**:
1. Close and reopen the terminal
2. Restart applications (IDEs, etc.)
3. In extreme cases, restart Windows

### Windows Defender blocks the executable

**Cause**: Windows may block executables downloaded from the internet

**Solution**:
1. Verify the source (official GitHub releases)
2. Build from source (Method 2)
3. Add an exception in Windows Defender

### "Invalid Java installation path"

**Cause**: The specified path doesn't contain `bin\java.exe`

**Solution**:
```bash
# Make sure to specify the JDK ROOT directory
# Correct:
jv add C:\Program Files\Java\jdk-17

# Wrong:
jv add C:\Program Files\Java\jdk-17\bin
```

## Contributing

Contributions are welcome! If you have ideas, bug reports, or feature requests:

1. Open an [Issue](https://github.com/CostaBrosky/jv/issues)
2. Fork the project
3. Create a branch (`git checkout -b feature/amazing-feature`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## License

This project is released under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Extras

### Project Structure

```
jv/            # CLI entry point
├── internal/
│   ├── java/            # Java version detection
│   ├── config/          # Configuration management
│   └── env/             # Windows environment variables
├── main.go              # Main Go file
├── go.mod
└── README.md
```

### Useful Links

- [Quick Start Guide (QUICKSTART.md)](QUICKSTART.md)
- [Installation Guide (INSTALL.md)](INSTALL.md)
- [Project Structure (PROJECT_STRUCTURE.md)](PROJECT_STRUCTURE.md)
- [Changelog (CHANGELOG.md)](CHANGELOG.md)

---

<div align="center">
  
**Made with ❤️ to simplify Java development on Windows**

</div>
