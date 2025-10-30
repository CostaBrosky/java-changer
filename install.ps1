<#
.SYNOPSIS
Installer for Java Version Switcher (jv)

.DESCRIPTION
Downloads and installs jv.exe. Java installations can be managed using 'jv install' command.

.PARAMETER Version
Version of jv to install (default: latest)

.PARAMETER InstallDir
Installation directory (default: $HOME\.local\bin)

.PARAMETER NoModifyPath
Don't add to PATH environment variable

.PARAMETER NoCompletion
Skip PowerShell completion installation

.PARAMETER Silent
Non-interactive mode, uses all defaults

.EXAMPLE
irm https://raw.githubusercontent.com/CostaBrosky/jv/main/install.ps1 | iex

.EXAMPLE
.\install.ps1
#>

param(
    [Parameter(HelpMessage = "Version of jv to install")]
    [string]$Version = "latest",

    [Parameter(HelpMessage = "Installation directory")]
    [string]$InstallDir,

    [Parameter(HelpMessage = "Don't modify PATH")]
    [switch]$NoModifyPath,

    [Parameter(HelpMessage = "Skip PowerShell completion")]
    [switch]$NoCompletion,

    [Parameter(HelpMessage = "Non-interactive mode")]
    [switch]$Silent
)

$ErrorActionPreference = "Stop"

# Constants
$GITHUB_REPO = "CostaBrosky/jv"

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

# Check if running as administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Restart script with elevated privileges
function Invoke-ElevatedScript {
    param([string[]]$Arguments)

    Write-Info "Administrator privileges required for system-wide installation"
    Write-Info "Restarting script with elevated privileges..."

    $scriptPath = $MyInvocation.PSCommandPath
    if (-not $scriptPath) {
        $scriptPath = $PSCommandPath
    }

    try {
        $argString = ($Arguments | ForEach-Object {
            if ($_ -match '\s') {
                "`"$_`""
            } else {
                $_
            }
        }) -join ' '

        Start-Process powershell.exe -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`" $argString" -Verb RunAs -Wait
        exit 0
    }
    catch {
        Write-Err "Failed to elevate privileges: $_"
        Write-Host ""
        Write-Host "Please run this script as Administrator manually:" -ForegroundColor Yellow
        Write-Host "  Right-click PowerShell -> Run as Administrator" -ForegroundColor Yellow
        Write-Host "  Then run: .\install.ps1" -ForegroundColor Yellow
        exit 1
    }
}

# Initialize environment and validate prerequisites
function Initialize-Environment {
    Write-Info "Validating environment..."

    # Check PowerShell version
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        throw "PowerShell 5.0 or higher is required. Current version: $($PSVersionTable.PSVersion)"
    }

    # Check execution policy
    $allowedPolicies = @('Unrestricted', 'RemoteSigned', 'Bypass')
    $currentPolicy = (Get-ExecutionPolicy).ToString()
    if ($currentPolicy -notin $allowedPolicies) {
        Write-Err "PowerShell execution policy is too restrictive: $currentPolicy"
        Write-Host ""
        Write-Host "To fix this, run PowerShell as Administrator and execute:"
        Write-Host "  Set-ExecutionPolicy RemoteSigned -Scope CurrentUser"
        Write-Host ""
        throw "Execution policy check failed"
    }

    # Check TLS 1.2 support
    if ([System.Enum]::GetNames([System.Net.SecurityProtocolType]) -notcontains 'Tls12') {
        throw "TLS 1.2 support is required. Please install .NET Framework 4.5 or higher"
    }

    # Ensure TLS 1.2 is enabled
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

    Write-Success "Environment validation passed"
}

# Detect Windows architecture
function Get-WindowsArchitecture {
    # Try environment variables first (most reliable and always available)
    $processorArch = $env:PROCESSOR_ARCHITECTURE
    $processorArchW6432 = $env:PROCESSOR_ARCHITEW6432

    # Handle WOW64 scenarios (32-bit PowerShell on 64-bit Windows)
    # PROCESSOR_ARCHITEW6432 exists only when running 32-bit process on 64-bit Windows
    if ($processorArchW6432) {
        $processorArch = $processorArchW6432
    }

    # Map Windows architecture names to expected format
    switch -Regex ($processorArch) {
        "AMD64|x64" { return "x64" }
        "x86|i386|i686" { return "x86" }
        "ARM64|aarch64" { return "arm64" }
        "ARM" { return "arm" }
        default {
            # Final fallback using .NET
            if ([Environment]::Is64BitOperatingSystem) {
                Write-Warn "Could not determine exact architecture from '$processorArch', defaulting to x64"
                return "x64"
            }
            else {
                Write-Warn "Could not determine exact architecture from '$processorArch', defaulting to x86"
                return "x86"
            }
        }
    }
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

# Download jv.exe from GitHub releases
function DownloadJv($version, $arch) {
    Write-Info "Downloading jv $version for $arch..."

    try {
        if ($version -eq "latest") {
            $releaseUrl = "https://api.github.com/repos/$GITHUB_REPO/releases/latest"
            $release = Invoke-RestMethod -Uri $releaseUrl -ErrorAction Stop
            $version = $release.tag_name
        }

        # Map architecture to your naming convention
        $archName = switch ($arch) {
            "x64" { "amd64" }
            "arm64" { "arm64" }
            default { "amd64" }
        }

        # Construct download URL for ZIP file
        # Format: jv_v1.0.0_windows_amd64.zip
        $zipName = "jv_${version}_windows_${archName}.zip"
        $downloadUrl = "https://github.com/$GITHUB_REPO/releases/download/$version/$zipName"

        $tempDir = Join-Path $env:TEMP "jv-install"
        if (-not (Test-Path $tempDir)) {
            New-Item -ItemType Directory -Path $tempDir | Out-Null
        }

        $zipPath = Join-Path $tempDir $zipName

        Write-Info "Downloading from: $downloadUrl"
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -ErrorAction Stop

        Write-Success "Downloaded $zipName"

        # Extract ZIP
        Write-Info "Extracting..."
        $extractDir = Join-Path $tempDir "extracted"
        if (Test-Path $extractDir) {
            Remove-Item -Path $extractDir -Recurse -Force
        }
        Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

        # Find the exe inside (format: jv-windows-amd64.exe or jv-windows-arm64.exe)
        $exeName = "jv-windows-${archName}.exe"
        $exePath = Join-Path $extractDir $exeName

        if (-not (Test-Path $exePath)) {
            # Try alternative: maybe it's just jv.exe
            $exePath = Join-Path $extractDir "jv.exe"
            if (-not (Test-Path $exePath)) {
                throw "Could not find executable in ZIP. Expected: $exeName"
            }
        }

        Write-Success "Extracted jv executable"
        return $exePath
    }
    catch {
        throw "Failed to download jv: $_"
    }
}

# Find existing Java installations
function Find-JavaInstallations {
    Write-Info "Scanning for Java installations..."

    $searchPaths = @(
        "C:\Program Files\Java",
        "C:\Program Files (x86)\Java",
        "C:\Program Files\Eclipse Adoptium",
        "C:\Program Files\Eclipse Foundation",
        "C:\Program Files\Zulu",
        "C:\Program Files\Amazon Corretto",
        "C:\Program Files\Microsoft"
    )

    $found = @()

    foreach ($basePath in $searchPaths) {
        if (Test-Path $basePath) {
            $dirs = Get-ChildItem -Path $basePath -Directory -ErrorAction SilentlyContinue
            foreach ($dir in $dirs) {
                $javaExe = Join-Path $dir.FullName "bin\java.exe"
                if (Test-Path $javaExe) {
                    $found += $dir.FullName
                }
            }
        }
    }

    if ($found.Count -gt 0) {
        Write-Success "Found $($found.Count) Java installation(s)"
    }
    else {
        Write-Warn "No Java installations found"
    }

    return $found
}

# Install jv.exe (XDG-compliant)
function Install-JV($binPath, $binDir) {
    Write-Info "Installing jv to $binDir..."

    # Create ~/.local/bin directory if it doesn't exist
    if (-not (Test-Path $binDir)) {
        New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    }

    # Install jv.exe directly in ~/.local/bin/
    $targetPath = Join-Path $binDir "jv.exe"
    Copy-Item -Path $binPath -Destination $targetPath -Force

    Write-Success "Installed jv.exe to: $targetPath"
    return $binDir
}

# Add directory to user PATH
function Add-ToPath($directory) {
    Write-Info "Adding $directory to user PATH..."

    $regPath = "HKCU:\Environment"
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    # Check if already in PATH
    $paths = $currentPath -split ";" | Where-Object { $_ -ne "" }
    $normalizedDir = $directory.TrimEnd('\')

    foreach ($p in $paths) {
        if ($p.TrimEnd('\') -eq $normalizedDir) {
            Write-Info "Directory already in user PATH"
            return $false
        }
    }

    # Add to PATH
    $newPath = "$normalizedDir;$currentPath"
    Set-ItemProperty -Path $regPath -Name "Path" -Value $newPath

    # Broadcast environment change
    BroadcastEnvironmentChange

    Write-Success "Added to user PATH"
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

# Create initial config file (XDG-compliant)
function Initialize-Config($javaInstallations) {
    Write-Info "Creating configuration..."

    # Save config following XDG Base Directory: $HOME/.config/jv/jv.json
    $configDir = Join-Path $HOME ".config\jv"
    if (-not (Test-Path $configDir)) {
        New-Item -ItemType Directory -Path $configDir -Force | Out-Null
    }

    $configPath = Join-Path $configDir "jv.json"

    $config = @{
        custom_paths = @($javaInstallations)
        search_paths = @()
        installed_jdks = @()
    }

    $configJson = $config | ConvertTo-Json -Depth 10
    # Use UTF8 without BOM to avoid parsing issues
    $utf8NoBom = New-Object System.Text.UTF8Encoding $false
    [System.IO.File]::WriteAllText($configPath, $configJson, $utf8NoBom)

    Write-Success "Configuration created at: $configPath"
    return $configPath
}

# Install PowerShell completion
function Install-Completion {
    Write-Info "Setting up PowerShell autocomplete..."

    try {
        # Check if profile exists
        if (-not (Test-Path $PROFILE)) {
            $profileDir = Split-Path $PROFILE -Parent
            if (-not (Test-Path $profileDir)) {
                New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
            }
            New-Item -ItemType File -Path $PROFILE -Force | Out-Null
        }

        # Read current profile
        $currentProfile = Get-Content $PROFILE -Raw -ErrorAction SilentlyContinue

        # Check if jv completion is already installed
        if ($currentProfile -and $currentProfile.Contains("# jv completion")) {
            Write-Info "PowerShell completion already configured"
            return $false
        }

        # PowerShell completion script (inline, no dependency on jv.exe)
        $completionScript = @'

# jv completion - Auto-generated by installer
$jvCommands = @{
    'list'        = 'List all available Java versions'
    'use'         = 'Switch to specified Java version'
    'switch'      = 'Quick interactive version switcher'
    'current'     = 'Show current Java version'
    'add'         = 'Add a specific Java installation'
    'remove'      = 'Remove a custom installation'
    'add-path'    = 'Add directory to search for Java installations'
    'remove-path' = 'Remove directory from search paths'
    'list-paths'  = 'Show all search paths'
    'install'     = 'Install Java from open-source distributors'
    'doctor'      = 'Run diagnostics on your Java environment'
    'repair'      = 'Automatically fix configuration issues'
    'completion'  = 'Generate shell completion scripts'
    'version'     = 'Show version information'
    'help'        = 'Show help message'
}

function Get-JvVersions {
    try {
        $output = jv list 2>$null
        if ($output) {
            $versions = $output | Select-String -Pattern '^\s*[*→\s]\s*(\S+)' | ForEach-Object {
                $_.Matches.Groups[1].Value
            }
            return $versions
        }
    }
    catch { }
    return @()
}

function Get-JvCustomPaths {
    try {
        $configPath = "$env:USERPROFILE\.config\jv\jv.json"
        if (Test-Path $configPath) {
            $config = Get-Content $configPath -Raw | ConvertFrom-Json
            return $config.custom_paths
        }
    }
    catch { }
    return @()
}

function Get-JvSearchPaths {
    try {
        $configPath = "$env:USERPROFILE\.config\jv\jv.json"
        if (Test-Path $configPath) {
            $config = Get-Content $configPath -Raw | ConvertFrom-Json
            return $config.search_paths
        }
    }
    catch { }
    return @()
}

Register-ArgumentCompleter -CommandName jv -ScriptBlock {
    param($commandName, $wordToComplete, $commandAst, $fakeBoundParameters)

    $line = $commandAst.ToString()
    $words = $line -split '\s+'

    if ($words.Count -le 2) {
        $jvCommands.Keys | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $jvCommands[$_])
        }
        return
    }

    $command = $words[1]

    switch ($command) {
        'use' {
            $versions = Get-JvVersions
            $versions | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', "Java $_")
            }
        }
        'add' {
            Get-ChildItem -Path "$wordToComplete*" -Directory -ErrorAction SilentlyContinue | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_.FullName, $_.Name, 'ParameterValue', $_.FullName)
            }
        }
        'remove' {
            $paths = Get-JvCustomPaths
            $paths | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'add-path' {
            Get-ChildItem -Path "$wordToComplete*" -Directory -ErrorAction SilentlyContinue | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_.FullName, $_.Name, 'ParameterValue', $_.FullName)
            }
        }
        'remove-path' {
            $paths = Get-JvSearchPaths
            $paths | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'completion' {
            @('powershell', 'bash', 'fish') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', 'Shell completion')
            }
        }
    }
}

'@

        # Append completion script to profile
        $utf8NoBom = New-Object System.Text.UTF8Encoding $false
        $newContent = $currentProfile + "`n" + $completionScript + "`n"
        [System.IO.File]::WriteAllText($PROFILE, $newContent, $utf8NoBom)

        Write-Success "PowerShell completion installed to: $PROFILE"
        return $true
    }
    catch {
        Write-Warn "Failed to install completion: $_"
        return $false
    }
}

# Main execution
try {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "   Java Version Switcher Installer" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""

    # Note: Admin privileges are not required for jv tool installation
    # Only Java environment configuration (via 'jv install' or 'jv repair') requires admin

    Initialize-Environment

    $arch = Get-WindowsArchitecture
    Write-Info "Detected architecture: $arch"

    $finalInstallDir = Get-InstallDirectory
    Write-Info "Install directory: $finalInstallDir"

    # Download jv.exe
    $jvPath = DownloadJv -version $Version -arch $arch

    # Detect existing Java installations for initial config
    $javaInstalls = Find-JavaInstallations

    # Install jv.exe to ~/.local/bin
    $binDir = Install-JV -binPath $jvPath -binDir $finalInstallDir

    # Add ~/.local/bin to PATH (user level)
    $pathModified = $false
    if (-not $NoModifyPath) {
        $pathModified = Add-ToPath -directory $binDir
    }

    # Create config in ~/.config/jv/ with detected Java installations
    $configPath = Initialize-Config -javaInstallations $javaInstalls

    # Install PowerShell completion (no dependency on jv.exe being in PATH)
    $completionInstalled = $false
    if (-not $NoCompletion) {
        $completionInstalled = Install-Completion
    }

    # Cleanup
    $tempDir = Join-Path $env:TEMP "jv-install"
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue

    # Success message
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "   Installation Complete!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Success "jv installed to: $binDir\jv.exe"
    Write-Success "Configuration: $configPath"

    if ($pathModified) {
        Write-Success "Added $binDir to user PATH"
    } else {
        Write-Info "$binDir already in user PATH"
    }

    if ($javaInstalls.Count -gt 0) {
        Write-Success "Found $($javaInstalls.Count) existing Java installation(s)"
    }

    if ($completionInstalled) {
        Write-Success "PowerShell autocomplete configured"
    }

    Write-Host ""
    Write-Host "Installation Summary:" -ForegroundColor Yellow
    Write-Host "  Executable:   $binDir\jv.exe" -ForegroundColor Cyan
    Write-Host "  Config:       $configPath" -ForegroundColor Cyan
    if ($completionInstalled) {
        Write-Host "  Autocomplete: $PROFILE" -ForegroundColor Cyan
    }

    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  1. Restart your terminal to reload environment variables"
    
    if ($completionInstalled) {
        Write-Host "     (This will also enable autocomplete - try: jv <tab>)" -ForegroundColor Gray
    } else {
        Write-Host "     (or run: `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','User'))"
    }
    
    Write-Host "  2. Run: jv list          (see detected Java installations)"
    
    if ($completionInstalled) {
        Write-Host "  3. Try autocomplete: jv <tab>" -ForegroundColor Green
    }
    
    Write-Host ""
    
    if ($javaInstalls.Count -eq 0) {
        Write-Host "No Java found. To install Java:" -ForegroundColor Yellow
        Write-Host "  jv install" -ForegroundColor Cyan
    } else {
        Write-Host "To switch Java version:" -ForegroundColor Yellow
        Write-Host "  jv use <version>" -ForegroundColor Cyan
        Write-Host "  or: jv switch (interactive)" -ForegroundColor Cyan
    }

    Write-Host ""
    Write-Host "For more information: jv help" -ForegroundColor Gray
    
    if ($completionInstalled) {
        Write-Host ""
        Write-Host "✨ Autocomplete enabled! After restarting, try:" -ForegroundColor Green
        Write-Host "   jv <tab>      - List all commands" -ForegroundColor Gray
        Write-Host "   jv use <tab>  - List Java versions" -ForegroundColor Gray
    }
    
    Write-Host ""
}
catch {
    Write-Err "Installation failed: $_"
    Write-Host ""
    Write-Host "For help, visit: https://github.com/$GITHUB_REPO/issues" -ForegroundColor Gray
    exit 1
}
