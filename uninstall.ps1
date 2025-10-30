<#
.SYNOPSIS
Uninstaller for Java Version Switcher (jv)

.DESCRIPTION
Removes jv.exe, its configuration, and PowerShell completion setup.

.PARAMETER InstallDir
Installation directory where jv.exe was located (default: $HOME\.local\bin)

.PARAMETER NoModifyPath
Don't remove from PATH environment variable

.PARAMETER NoCompletion
Skip PowerShell completion removal

.PARAMETER Silent
Non-interactive mode, uses all defaults

.EXAMPLE
.\uninstall.ps1
#>

param(
    [Parameter(HelpMessage = "Installation directory where jv.exe was located")]
    [string]$InstallDir,

    [Parameter(HelpMessage = "Don't modify PATH")]
    [switch]$NoModifyPath,

    [Parameter(HelpMessage = "Skip PowerShell completion removal")]
    [switch]$NoCompletion,

    [Parameter(HelpMessage = "Non-interactive mode")]
    [switch]$Silent
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Info($message) {
    Write-Host "[INFO] $message" -ForegroundColor Cyan
}

function Write-Success($message) {
    Write-Host "[OK] $message" -ForegroundColor Green
}

function Write-Warn($message) {
    Write-Host "[WARN] $message" -ForegroundColor Yellow
}

function Write-Err($message) {
    Write-Host "[ERROR] $message" -ForegroundColor Red
}

# Get install directory (XDG-compliant)
function Get-InstallDirectory {
    if ($InstallDir) {
        return $InstallDir
    }

    # Follow XDG Base Directory specification
    # Executable: $HOME/.local/bin/jv.exe
    $localBin = Join-Path $HOME ".local\bin"

    return $localBin
}

# Remove jv.exe from the installation directory
function Remove-JV($binDir) {
    Write-Info "Removing jv from $binDir..."

    $targetPath = Join-Path $binDir "jv.exe"
    if (Test-Path $targetPath) {
        Remove-Item -Path $targetPath -Force
        Write-Success "Removed jv.exe from: $targetPath"
        return $true
    } else {
        Write-Warn "jv.exe not found at: $targetPath"
        return $false
    }
}

# Remove directory from user PATH
function Remove-FromPath($directory) {
    Write-Info "Removing $directory from user PATH..."

    $regPath = "HKCU:\Environment"
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if (-not $currentPath) {
        Write-Warn "User PATH environment variable is empty or not found."
        return $false
    }

    # Check if the directory is in PATH
    $paths = $currentPath -split ";" | Where-Object { $_ -ne "" }
    $normalizedDir = $directory.TrimEnd('\')
    $newPaths = $paths | Where-Object { $_.TrimEnd('\') -ne $normalizedDir }

    if ($paths.Count -eq $newPaths.Count) {
        Write-Info "Directory $normalizedDir was not found in user PATH."
        return $false
    }

    # Rebuild PATH
    $newPath = ($newPaths -join ";")
    if (-not $newPath) {
        # If PATH becomes empty after removal, set it to a minimal value or clear it
        # Setting it to $null or "" might not be ideal, so we could keep it as is or warn.
        # For this script, we'll set it to empty string which effectively removes it from the registry value.
        # However, the environment variable in running processes might still hold the old value.
        Write-Warn "Removing the last path entry will result in an empty PATH. This might affect other tools. Proceeding cautiously."
        $newPath = ""
    }

    Set-ItemProperty -Path $regPath -Name "Path" -Value $newPath

    # Broadcast environment change
    BroadcastEnvironmentChange

    Write-Success "Removed $normalizedDir from user PATH"
    return $true
}

# Broadcast environment variable changes
function BroadcastEnvironmentChange {
    try {
        $HWND_BROADCAST = [IntPtr]0xffff
        $WM_SETTINGCHANGE = 0x1a
        $result = [UIntPtr]::Zero

        if (-not ([System.Management.Automation.PSTypeName]'Win32.Environment').Type) {
            Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
namespace Win32 {
    public class Environment {
        [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
        public static extern IntPtr SendMessageTimeout(
            IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam,
            uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
    }
}
"@
        }

        [Win32.Environment]::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [UIntPtr]::Zero, "Environment", 2, 5000, [ref]$result) | Out-Null
    }
    catch {
        Write-Warn "Failed to broadcast environment change: $_"
    }
}

# Remove initial config file (XDG-compliant)
function Remove-Config {
    Write-Info "Removing configuration..."

    # Locate config following XDG Base Directory: $HOME/.config/jv/jv.json
    $configDir = Join-Path $HOME ".config\jv"
    $configPath = Join-Path $configDir "jv.json"

    $removedConfig = $false
    $removedDir = $false

    if (Test-Path $configPath) {
        Remove-Item -Path $configPath -Force
        Write-Success "Removed configuration file: $configPath"
        $removedConfig = $true
    } else {
        Write-Warn "Configuration file not found: $configPath"
    }

    if (Test-Path $configDir) {
        # Check if the directory is now empty
        if ((Get-ChildItem -Path $configDir -Force | Measure-Object).Count -eq 0) {
            Remove-Item -Path $configDir -Recurse -Force
            Write-Success "Removed empty configuration directory: $configDir"
            $removedDir = $true
        } else {
            Write-Info "Configuration directory $configDir contains other files, keeping it."
        }
    }

    return $removedConfig, $removedDir
}

# Remove PowerShell completion
function Remove-Completion {
    Write-Info "Removing PowerShell autocomplete..."

    try {
        if (-not (Test-Path $PROFILE)) {
            Write-Warn "PowerShell profile not found at: $PROFILE"
            return $false
        }

        # Read current profile
        $currentProfile = Get-Content $PROFILE -Raw -ErrorAction SilentlyContinue

        # Check if jv completion is installed
        if ($currentProfile -and $currentProfile.Contains("# jv completion")) {
            # Remove the completion block
            $pattern = [regex]::Escape("# jv completion - begin") + ".*?" + [regex]::Escape("# jv completion - end")
            $newContent = [regex]::Replace($currentProfile, $pattern, "", [System.Text.RegularExpressions.RegexOptions]::Singleline)

            # Write the updated profile back (UTF8 without BOM)
            $utf8NoBom = New-Object System.Text.UTF8Encoding $false
            [System.IO.File]::WriteAllText($PROFILE, $newContent, $utf8NoBom)

            Write-Success "PowerShell completion removed from: $PROFILE"
            return $true
        } else {
            Write-Info "jv PowerShell completion not found in profile: $PROFILE"
            return $false
        }
    }
    catch {
        Write-Warn "Failed to remove completion: $_"
        return $false
    }
}

# Main execution
try {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "   Java Version Switcher Uninstaller" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Host ""

    $finalInstallDir = Get-InstallDirectory
    Write-Info "Install directory to check: $finalInstallDir"

    # Remove jv.exe
    $exeRemoved = Remove-JV -binDir $finalInstallDir

    # Remove config
    $configRemoved, $configDirRemoved = Remove-Config

    # Remove from PATH
    $pathModified = $false
    if (-not $NoModifyPath) {
        $pathModified = Remove-FromPath -directory $finalInstallDir
    }

    # Remove PowerShell completion
    $completionRemoved = $false
    if (-not $NoCompletion) {
        $completionRemoved = Remove-Completion
    }

    # Success message
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "   Uninstallation Complete!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""

    if ($exeRemoved) {
        Write-Success "jv.exe has been removed."
    } else {
        Write-Info "jv.exe was not found or already removed."
    }

    if ($configRemoved) {
        Write-Success "jv configuration file has been removed."
    } else {
        Write-Info "jv configuration file was not found or already removed."
    }

    if ($configDirRemoved) {
        Write-Success "jv configuration directory has been removed."
    }

    if ($pathModified) {
        Write-Success "jv directory removed from user PATH."
    } else {
        Write-Info "jv directory was not in user PATH or PATH was not modified."
    }

    if ($completionRemoved) {
        Write-Success "PowerShell autocomplete for jv has been removed."
    } else {
        Write-Info "jv PowerShell autocomplete was not found or not removed."
    }

    Write-Host ""
    Write-Host "Summary:" -ForegroundColor Yellow
    Write-Host "  - Removed executable (if present)" -ForegroundColor Cyan
    Write-Host "  - Removed configuration file (if present)" -ForegroundColor Cyan
    Write-Host "  - Removed from PATH (if present)" -ForegroundColor Cyan
    Write-Host "  - Removed PowerShell completion (if present)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Note: You may need to restart your terminal or PowerShell session" -ForegroundColor Yellow
    Write-Host "      for all changes (especially PATH) to take effect." -ForegroundColor Yellow
    Write-Host ""

}
catch {
    Write-Err "Uninstallation failed: $_"
    Write-Host ""
    Write-Host "For help, visit: https://github.com/CostaBrosky/jv/issues" -ForegroundColor Gray
    exit 1
}