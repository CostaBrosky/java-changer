# Installer Documentation

Technical documentation for the Java Version Switcher PowerShell installer script.

## Overview

The `install.ps1` script provides automated installation of `jv`. Java installations can be managed separately using the `jv install` command. The script follows modern PowerShell installer patterns inspired by tools like UV (Astral's Python tool) and Rustup.

## Features

- ✅ One-liner installation via `irm | iex`
- ✅ Downloads jv.exe from GitHub releases
- ✅ Auto-detects Windows architecture (x64, x86, ARM64)
- ✅ Scans for existing Java installations
- ✅ Automatic PATH configuration
- ✅ **Automatic PowerShell autocomplete installation** (NEW!)
- ✅ Creates initial configuration file
- ✅ Interactive and silent modes
- ✅ Lightweight and focused (Java management via `jv install` command)

## Usage

### Basic Installation

```powershell
# Standard installation with prompts
irm https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 | iex
```

### Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `-Version` | string | "latest" | Version of jv to install |
| `-InstallDir` | string | `$HOME\.local\bin` | Installation directory |
| `-NoModifyPath` | switch | false | Don't modify PATH |
| `-NoCompletion` | switch | false | Skip PowerShell autocomplete installation |
| `-Silent` | switch | false | Non-interactive mode |

### Examples

```powershell
# Basic installation (with autocomplete!)
irm https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 | iex

# Silent installation for CI/CD
irm https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 | iex -Args "-Silent"

# Custom installation directory
irm https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 | iex -Args "-InstallDir", "C:\tools"

# Skip autocomplete installation
irm https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 | iex -Args "-NoCompletion"

# Download script and run locally
Invoke-WebRequest -Uri https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 -OutFile install.ps1
.\install.ps1

# After installation, restart PowerShell:
# 1. Autocomplete is already enabled!
# 2. Try: jv <tab>
# 3. Try: jv use <tab>
# 4. Use jv to install Java: jv install
```

**Note:** The `-NoCompletion` flag skips autocomplete installation if you don't want it.

## Installation Flow

### 1. Environment Validation

The script validates:
- PowerShell version (requires 5.0+)
- Execution policy (must be Unrestricted, RemoteSigned, or Bypass)
- TLS 1.2 support (requires .NET Framework 4.5+)

### 2. Architecture Detection

Uses `System.Runtime.InteropServices.RuntimeInformation` to detect:
- x64 (64-bit Intel/AMD)
- x86 (32-bit)
- ARM64 (ARM 64-bit)

Fallback to `[Environment]::Is64BitOperatingSystem` for older PowerShell versions.

### 3. Download jv.exe

- Queries GitHub API for latest release (or uses specified version)
- Downloads `jv.exe` to temporary directory
- No checksum verification for jv.exe (relies on HTTPS)

### 4. Java Detection

Scans standard directories:
```
C:\Program Files\Java
C:\Program Files (x86)\Java
C:\Program Files\Eclipse Adoptium
C:\Program Files\Eclipse Foundation
C:\Program Files\Zulu
C:\Program Files\Amazon Corretto
C:\Program Files\Microsoft
```

Validates each directory for `bin\java.exe` existence.

### 5. Installation

- Creates install directory if not exists
- Copies `jv.exe` to install directory
- Destination: `{InstallDir}\jv.exe`

### 6. PATH Modification

If not skipped:
- Reads current user PATH from registry (`HKCU:\Environment`)
- Checks for duplicates
- Prepends install directory to PATH
- Updates registry
- Broadcasts `WM_SETTINGCHANGE` message to notify Windows

**Registry Path**: `HKEY_CURRENT_USER\Environment\Path`

### 7. Configuration

Creates `%USERPROFILE%\.config\jv\jv.json` with detected Java installations:
```json
{
  "custom_paths": [
    "C:\\Program Files\\Java\\jdk-17",
    "C:\\Program Files\\Eclipse Adoptium\\jdk-21"
  ],
  "search_paths": [],
  "installed_jdks": []
}
```

The `installed_jdks` array will be populated when using `jv install` command.

### 8. PowerShell Autocomplete Installation (NEW!)

Automatically sets up tab completion without requiring `jv.exe` to be in PATH:

**Process:**
1. Checks if `$PROFILE` exists (creates if needed)
2. Reads current profile content
3. Checks for duplicates (avoids re-installing)
4. Appends completion script inline (self-contained PowerShell code)
5. Writes to profile with UTF-8 no BOM

**Completion Script Includes:**
- `Register-ArgumentCompleter` for `jv` command
- `Get-JvVersions()` - Parses `jv list` output for version completion
- `Get-JvCustomPaths()` - Reads config for `jv remove` completion
- `Get-JvSearchPaths()` - Reads config for `jv remove-path` completion
- Path completion for `jv add` and `jv add-path`

**Result:** After restarting PowerShell:
- `jv <tab>` → Shows all commands with descriptions
- `jv use <tab>` → Shows Java versions (e.g., 21, 17, 11)
- `jv add <tab>` → Directory path completion
- `jv remove <tab>` → Custom paths from config
- `jv remove-path <tab>` → Search paths from config

**Skip with:** `-NoCompletion` parameter

**Why inline?** The completion script is embedded directly in PowerShell (no dependency on `jv.exe`), so it works immediately without requiring PATH refresh.

### 9. Cleanup

- Removes temporary download directory
- Displays success message and next steps

## Java Installation via `jv install`

After installing the `jv` tool, Java distributions can be installed using the `jv install` command.

### Features

- **Interactive distributor selection**: Choose from Eclipse Adoptium, Azul Zulu, Amazon Corretto, etc.
- **Version menu**: Shows LTS and feature releases
- **Installed version detection**: Marks already-installed versions
- **Automatic download & extraction**: Downloads, verifies checksum, and extracts JDK
- **Smart installation location**:
  - With admin: `C:\Program Files\{Distributor}\jdk-{version}`
  - Without admin: `%USERPROFILE%\.jv\jdk-{version}`
- **Auto-configuration**: Sets JAVA_HOME if not already set (requires admin)

### Usage

```bash
# Interactive installation with beautiful TUI
jv install
```

**Interactive Features:**
- **Arrow key navigation** (↑↓) - No typing required
- **Visual selection** - See what you're choosing in real-time
- **Multi-select support** - Install multiple versions with Space key
- **Smart detection** - Shows [Installed] markers
- **Keyboard shortcuts:**
  - `↑↓` or `j/k`: Navigate
  - `Space`: Toggle selection (multi-select)
  - `Enter`: Confirm
  - `Ctrl+C`: Cancel

**Installation Flow:**
1. Select distributor (currently: Eclipse Adoptium)
2. Choose installation mode:
   - Single version
   - Multiple versions (batch)
3. Select version(s) - navigate with arrows, select with Space
4. Select installation scope (system-wide or user-only, if admin)
5. Auto-download and install

### Installation Scopes

**With administrator privileges:**

You can choose between:
- **System-wide** (recommended): Installs to `C:\Program Files\Eclipse Adoptium\jdk-{version}`, available for all users
- **User-only**: Installs to `%USERPROFILE%\.jv\jdk-{version}`, only for current user

**Without administrator privileges:**

- Automatically installs to `%USERPROFILE%\.jv\jdk-{version}` (user-only)
- JAVA_HOME cannot be set automatically (requires running `jv use` as administrator)

### Scope Tracking

Installations via `jv install` are tracked in the config file with their scope:

```json
{
  "installed_jdks": [
    {
      "version": "21",
      "path": "C:\\Program Files\\Eclipse Adoptium\\jdk-21",
      "distributor": "Eclipse Adoptium",
      "installed_at": "2025-10-30T12:34:56Z",
      "scope": "system"
    }
  ]
}
```

This allows `jv list` to show the installation scope:

```
Available Java versions:

* 21.0.8          C:\Program Files\Eclipse Adoptium\jdk-21 (installed, system-wide)
  17.0.12         C:\Users\username\.jv\jdk-17 (installed, user-only)
  11.0.24         C:\Program Files\Java\jdk-11 (auto)
```

### Supported Distributors

| Distributor | Status | Notes |
|-------------|--------|-------|
| Eclipse Adoptium (Temurin) | ✅ Active | Default, well-supported |
| Azul Zulu | 🔜 Coming Soon | Planned |
| Amazon Corretto | 🔜 Coming Soon | Planned |
| Microsoft Build of OpenJDK | 🔜 Coming Soon | Planned |

## Technical Details

### Eclipse Adoptium API

**Endpoint**: `https://api.adoptium.net/v3/`

**Request Example**:
```http
GET /v3/assets/latest/21/hotspot?architecture=x64&image_type=jdk&os=windows&vendor=eclipse
```

**Response** (simplified):
```json
[
  {
    "binary": {
      "architecture": "x64",
      "image_type": "jdk",
      "os": "windows",
      "package": {
        "link": "https://github.com/adoptium/temurin21-binaries/releases/download/.../OpenJDK21U-jdk_x64_windows_hotspot_21.0.5_11.zip",
        "checksum": "abc123...",
        "size": 204850009
      }
    },
    "version": {
      "major": 21,
      "openjdk_version": "21.0.5+11-LTS"
    }
  }
]
```

### PATH Broadcasting

To notify Windows shell and applications of environment variable changes:

```powershell
$HWND_BROADCAST = [IntPtr]0xffff
$WM_SETTINGCHANGE = 0x1a

[Win32]::SendMessageTimeout(
    $HWND_BROADCAST,
    $WM_SETTINGCHANGE,
    [UIntPtr]::Zero,
    "Environment",
    2,
    5000,
    [ref]$result
)
```

This broadcasts `WM_SETTINGCHANGE` to all top-level windows, notifying them to reload environment variables.

## Security Considerations

### 1. Execution Policy

The script checks for allowed policies:
- `Unrestricted`
- `RemoteSigned` (recommended)
- `Bypass`

If not allowed, provides instructions to fix:
```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 2. TLS 1.2 Enforcement

```powershell
[Net.ServicePointManager]::SecurityProtocol =
    [Net.ServicePointManager]::SecurityProtocol -bor
    [Net.SecurityProtocolType]::Tls12
```

Ensures all downloads use secure connections.

### 3. Checksum Verification

**For Java JDK downloads**:
```powershell
$actualHash = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()
if ($actualHash -ne $checksum.ToLower()) {
    throw "Checksum verification failed!"
}
```

**Note**: jv.exe currently doesn't verify checksum (GitHub HTTPS is trusted).

### 4. No Admin Required

- Installer runs under user context
- Modifies user-level PATH only
- Downloads to user directories
- No system-level changes

This reduces security risks and permission requirements.

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| "PowerShell 5.0 or higher is required" | Old PowerShell version | Update to Windows PowerShell 5.1 or PowerShell 7+ |
| "Execution policy is too restrictive" | Strict execution policy | Run `Set-ExecutionPolicy RemoteSigned -Scope CurrentUser` |
| "TLS 1.2 support is required" | Old .NET Framework | Install .NET Framework 4.5+ |
| "Failed to download jv.exe" | Network/GitHub issue | Check internet connection, firewall |
| "Checksum verification failed" | Corrupted download | Re-run installer |
| "No JDK found for Java X" | Adoptium doesn't have that version/arch | Try different Java version |

### Rollback

On failure:
```powershell
catch {
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    Write-Error "Installation failed: $_"
    exit 1
}
```

Temporary files are cleaned up, but partial PATH modifications may remain.

## CI/CD Integration

### GitHub Actions

```yaml
- name: Install jv
  shell: powershell
  run: |
    irm https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 | iex -Args "-Silent", "-JavaVersion", "17"
    $env:Path = [System.Environment]::GetEnvironmentVariable('Path','User')
    jv version
```

### Azure Pipelines

```yaml
- task: PowerShell@2
  inputs:
    targetType: 'inline'
    script: |
      irm https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 | iex -Args "-Silent", "-NoJava"
```

## Troubleshooting

### Script doesn't run

**Issue**: `running scripts is disabled on this system`

**Solution**:
```powershell
# As Administrator
Set-ExecutionPolicy RemoteSigned -Scope LocalMachine

# Or for current user only
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Java download fails

**Issue**: Adoptium API returns no results

**Possible causes**:
- Unsupported architecture (e.g., x86 for Java 21)
- Network/firewall blocking API access
- Adoptium service down

**Solution**:
- Use `-NoJava` and install Java manually
- Try different Java version

### PATH not updated

**Issue**: `jv` command not found after installation

**Solution**:
1. Restart terminal
2. Or reload PATH: `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','User')`
3. Verify install location: `where.exe jv`

## Comparison with Other Installers

| Feature | jv installer | UV | Rustup |
|---------|--------------|-----|--------|
| One-liner install | ✅ | ✅ | ✅ |
| Auto-download tool | ✅ | ✅ | ✅ |
| Auto-download runtime | ✅ (Java) | ❌ | ❌ |
| Architecture detection | ✅ | ✅ | ✅ |
| Checksum verification | ✅ | ✅ | ✅ |
| PATH modification | ✅ | ✅ | ✅ |
| Silent mode | ✅ | ✅ | ✅ |
| Rollback on failure | ⚠️ Partial | ✅ | ✅ |
| Self-update | ❌ | ✅ | ✅ |

## Future Enhancements

Potential improvements:

1. **Self-update capability**: `jv update` command
2. **Uninstaller**: `uninstall.ps1` script
3. **Multiple Java distributions**: Support Corretto, Zulu, etc.
4. **Proxy configuration**: Better corporate firewall support
5. **Installation receipt**: JSON metadata file for tracking
6. **Progress bars**: Visual feedback for downloads
7. **Rollback mechanism**: Complete undo on failure
8. **Offline mode**: Install from local files
9. **GUI installer**: Optional graphical interface
10. **Chocolatey/Scoop packages**: Package manager integration

## Contributing

To modify the installer:

1. Edit `install.ps1`
2. Test locally: `.\install.ps1 -WhatIf` (if implemented)
3. Test with parameters: `.\install.ps1 -Silent -JavaVersion 17`
4. Test error cases (no internet, wrong version, etc.)
5. Update this documentation
6. Commit changes

## References

- [Eclipse Adoptium API Docs](https://api.adoptium.net/q/swagger-ui/)
- [UV Installer Source](https://github.com/astral-sh/uv/blob/main/docs/guides/install-python.md)
- [PowerShell Best Practices](https://learn.microsoft.com/en-us/powershell/scripting/developer/cmdlet/approved-verbs-for-windows-powershell-commands)
- [Windows Environment Variables](https://learn.microsoft.com/en-us/windows/win32/procthread/environment-variables)

---

**Last Updated**: 2025
**Maintainer**: Java Version Switcher Contributors
