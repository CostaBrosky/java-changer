# Installation Guide

Step-by-step guide to install Java Version Switcher (jv) on Windows.

## Method 1: Download from GitHub Releases (Recommended)

### Step 1: Download

1. Open your browser and go to: https://github.com/USERNAME/java-changer/releases
2. Click on the latest release (the top one)
3. Under "Assets", click **`jv.exe`** to download the executable

### Step 2: Verify (Optional but Recommended)

Verify that the downloaded file is authentic using SHA256:

```powershell
# In PowerShell, go to the Downloads folder
cd ~\Downloads

# Calculate the checksum
Get-FileHash jv.exe -Algorithm SHA256

# Compare the output with the checksum in checksums.txt from the release
```

### Step 3: Installation

**Create a dedicated directory for your tools:**

1. Create a directory for your tools:
   ```cmd
   mkdir C:\tools
   ```

2. Copy `jv.exe` to `C:\tools\`

3. Add `C:\tools` to PATH:
   - Press `Win + X` → Select "System"
   - Click "Advanced system settings"
   - Click "Environment Variables"
   - Under "User variables" (or "System variables"), select "Path" and click "Edit"
   - Click "New" and add: `C:\tools`
   - Click OK on all windows
   - **Restart the terminal** to apply changes

### Step 4: Verify Installation

Open a new terminal (CMD or PowerShell) and try:

```cmd
jv version
jv help
```

If you see the output, installation succeeded! 🎉

## Method 2: Build from Source

### Prerequisites

- Go 1.21 or higher installed ([Download](https://go.dev/dl/))
- Git installed ([Download](https://git-scm.com/download/win))

### Steps

1. **Clone the repository:**
   ```bash
   git clone https://github.com/USERNAME/java-changer.git
   cd java-changer
   ```

2. **Download dependencies:**
   ```bash
   go mod download
   ```

3. **Build:**
   ```bash
   go build -ldflags="-s -w" -o jv.exe .
   ```

4. **Verify:**
   ```bash
   .\jv.exe version
   ```

5. **Install** (copy to a directory in PATH, see Method 1, Step 3)

## Basic Usage

After installation, here are the basic commands:

```bash
# List available Java versions
jv list

# Switch to Java 17 (RUN TERMINAL AS ADMINISTRATOR!)
jv use 17

# Show current version
jv current

# Add custom directory to scan
jv add-path C:\DevTools\Java
```

**IMPORTANT:** The `jv use` command requires administrator privileges because it modifies system environment variables.

## How to Run as Administrator

### CMD/PowerShell
1. Search for "cmd" or "PowerShell" in Start menu
2. **Right-click** → "Run as administrator"
3. Run `jv` commands

### Windows Terminal
1. Open Windows Terminal
2. Click the dropdown arrow ▼ next to the tab
3. Hold `Ctrl` and click on the profile (CMD or PowerShell)
4. It will open with administrator privileges

## Troubleshooting

### "jv is not recognized as an internal or external command"

**Solution:**
- Make sure you added `jv.exe` to PATH
- Restart the terminal after modifying PATH
- Verify with: `where jv` (should show the path to jv.exe)

### "failed to open registry key (run as administrator)"

**Solution:**
- The `jv use` command requires administrator privileges
- Run the terminal as administrator (see above)

### "No Java installations found"

**Solution:**
- Verify that Java is installed on your system
- If Java is in a non-standard directory, add it:
  ```bash
  jv add C:\path\to\jdk
  # or
  jv add-path C:\path\to\java-directory
  ```

### Windows Defender blocks the executable

**Solution:**
- This can happen with executables downloaded from the internet
- Verify the SHA256 checksum to ensure it's authentic
- Add an exception in Windows Defender if necessary
- Alternatively, build from source (Method 2)

### Environment variable changes don't apply

**Solution:**
- After running `jv use`, restart:
  - The current terminal (close and reopen)
  - Applications that need to use Java (e.g., IDEs)
- In extreme cases, restart Windows

## Uninstallation

To remove jv:

1. Delete `jv.exe` from where you installed it:
   ```cmd
   # If installed in tools directory
   del C:\tools\jv.exe
   ```

2. (Optional) Remove the configuration:
   ```cmd
   del %USERPROFILE%\.javarc
   ```

3. (Optional) Remove `C:\tools` from PATH if you're not using it for other tools

## Next Steps

Once installed, read the [Quick Start Guide](QUICKSTART.md) to learn how to use jv effectively.

For complete documentation, see the [README](README.md).

---

**Happy switching! ☕**
