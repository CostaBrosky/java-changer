<#
.SYNOPSIS
Installer for Java Version Switcher (jv)

.DESCRIPTION
Downloads and installs jv.exe with optional Java JDK download from Eclipse Adoptium.

.PARAMETER Version
Version of jv to install (default: latest)

.PARAMETER JavaVersion
Java version to download if not found. Options: 8, 11, 17, 21 (default: 21)

.PARAMETER InstallDir
Installation directory (default: $HOME\.local\bin)

.PARAMETER NoJava
Skip Java detection and download prompt

.PARAMETER NoModifyPath
Don't add to PATH environment variable

.PARAMETER Silent
Non-interactive mode, uses all defaults

.EXAMPLE
irm https://raw.githubusercontent.com/USER/java-changer/main/install.ps1 | iex

.EXAMPLE
irm https://raw.githubusercontent.com/USER/java-changer/main/install.ps1 | iex -Args "-JavaVersion", "17"

.EXAMPLE
.\install.ps1 -Silent -JavaVersion 21
#>

param(
    [Parameter(HelpMessage = "Version of jv to install")]
    [string]$Version = "latest",

    [Parameter(HelpMessage = "Java version to download if not found")]
    [ValidateSet("8", "11", "17", "21")]
    [string]$JavaVersion = "21",

    [Parameter(HelpMessage = "Installation directory")]
    [string]$InstallDir,

    [Parameter(HelpMessage = "Skip Java download prompt")]
    [switch]$NoJava,

    [Parameter(HelpMessage = "Don't modify PATH")]
    [switch]$NoModifyPath,

    [Parameter(HelpMessage = "Non-interactive mode")]
    [switch]$Silent
)

$ErrorActionPreference = "Stop"

# Constants
$GITHUB_REPO = "CostaBrosky/jv"
$ADOPTIUM_API = "https://api.adoptium.net/v3"

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

# Download Java JDK from Eclipse Adoptium
function DownloadJava($version, $arch) {
    Write-Info "Downloading Java $version from Eclipse Adoptium..."

    try {
        # Query Adoptium API
        $apiUrl = "$ADOPTIUM_API/assets/latest/$version/hotspot"
        $params = "?architecture=$arch&image_type=jdk&os=windows&vendor=eclipse"

        Write-Info "Querying Adoptium API..."
        $jdkInfo = Invoke-RestMethod -Uri "$apiUrl$params" -ErrorAction Stop

        if ($jdkInfo.Count -eq 0) {
            throw "No JDK found for Java $version on $arch"
        }

        $binary = $jdkInfo[0].binary
        $downloadUrl = $binary.package.link
        $checksum = $binary.package.checksum
        $size = [math]::Round($binary.package.size / 1MB, 2)
        $jdkVersion = $jdkInfo[0].version.openjdk_version

        Write-Info "Found JDK $jdkVersion"
        Write-Info "Size: $size MB"

        $tempDir = Join-Path $env:TEMP "jv-install"
        $zipPath = Join-Path $tempDir "jdk-$version.zip"

        Write-Info "Downloading JDK... (this may take a few minutes)"
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -ErrorAction Stop

        # Verify checksum
        Write-Info "Verifying checksum..."
        $actualHash = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()
        if ($actualHash -ne $checksum.ToLower()) {
            Remove-Item $zipPath -ErrorAction SilentlyContinue
            throw "Checksum verification failed! Expected: $checksum, Got: $actualHash"
        }
        Write-Success "Checksum verified"

        # Determine installation directory
        # Try Program Files first (requires admin), fallback to user directory
        $isAdmin = Test-Administrator
        if ($isAdmin) {
            $javaInstallBase = "C:\Program Files\Eclipse Adoptium"
        } else {
            Write-Warn "Not running as administrator, installing to user directory"
            $javaInstallBase = Join-Path $HOME ".jv"
        }

        if (-not (Test-Path $javaInstallBase)) {
            New-Item -ItemType Directory -Path $javaInstallBase -Force | Out-Null
        }

        # Extract to temp location first
        Write-Info "Extracting JDK..."
        $tempExtractDir = Join-Path $tempDir "extract"
        if (Test-Path $tempExtractDir) {
            Remove-Item -Path $tempExtractDir -Recurse -Force
        }
        Expand-Archive -Path $zipPath -DestinationPath $tempExtractDir -Force

        # Find the extracted directory
        $extractedDirs = Get-ChildItem -Path $tempExtractDir -Directory | Where-Object { $_.Name -match "jdk" }
        if ($extractedDirs.Count -eq 0) {
            throw "Failed to find extracted JDK directory"
        }

        $extractedJdkPath = $extractedDirs[0].FullName

        # Verify bin\java.exe exists
        $javaExe = Join-Path $extractedJdkPath "bin\java.exe"
        if (-not (Test-Path $javaExe)) {
            throw "Invalid JDK structure: bin\java.exe not found in $extractedJdkPath"
        }

        # Create a clean directory name: jdk-<version>
        $cleanDirName = "jdk-$version"
        $finalJdkPath = Join-Path $javaInstallBase $cleanDirName

        # Remove old installation if exists
        if (Test-Path $finalJdkPath) {
            Write-Info "Removing existing installation at $finalJdkPath"
            Remove-Item -Path $finalJdkPath -Recurse -Force
        }

        # Move to final location
        Move-Item -Path $extractedJdkPath -Destination $finalJdkPath -Force

        # Cleanup
        Remove-Item $zipPath -ErrorAction SilentlyContinue
        Remove-Item $tempExtractDir -Recurse -Force -ErrorAction SilentlyContinue

        Write-Success "Java $version installed to: $finalJdkPath"
        return $finalJdkPath
    }
    catch {
        throw "Failed to download Java: $_"
    }
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

# Set Java environment variables (JAVA_HOME and PATH)
function Set-JavaEnvironment($javaPath) {
    Write-Info "Setting up Java environment variables..."

    if (-not (Test-Administrator)) {
        Write-Warn "Administrator privileges required to set system environment variables"
        Write-Warn "JAVA_HOME will not be set automatically. You can set it manually or run 'jv use <version>' as administrator"
        return $false
    }

    try {
        # Verify Java installation
        $javaExe = Join-Path $javaPath "bin\java.exe"
        if (-not (Test-Path $javaExe)) {
            Write-Warn "Java executable not found at $javaExe, skipping environment setup"
            return $false
        }

        # Set JAVA_HOME at system level
        $regPath = "HKLM:\System\CurrentControlSet\Control\Session Manager\Environment"

        Write-Info "Setting JAVA_HOME to: $javaPath"
        Set-ItemProperty -Path $regPath -Name "JAVA_HOME" -Value $javaPath -ErrorAction Stop

        # Update system PATH
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
        $paths = $currentPath -split ";" | Where-Object { $_ -ne "" }

        # Remove old Java paths
        $filteredPaths = $paths | Where-Object {
            $_ -notmatch "java" -and
            $_ -notmatch "jdk" -and
            $_ -notmatch "jre" -and
            $_ -ne "%JAVA_HOME%\bin"
        }

        # Add %JAVA_HOME%\bin at the beginning
        $newPaths = @("%JAVA_HOME%\bin") + $filteredPaths
        $newPath = $newPaths -join ";"

        Write-Info "Adding %JAVA_HOME%\bin to system PATH"
        Set-ItemProperty -Path $regPath -Name "Path" -Value $newPath -ErrorAction Stop

        # Broadcast environment change
        BroadcastEnvironmentChange

        Write-Success "Java environment variables set successfully"
        Write-Success "JAVA_HOME = $javaPath"
        Write-Success "Added %JAVA_HOME%\bin to system PATH"
        return $true
    }
    catch {
        Write-Err "Failed to set Java environment variables: $_"
        return $false
    }
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
    Broadcast-EnvironmentChange

    Write-Success "Added to user PATH"
    return $true
}

# Broadcast environment variable changes
function Broadcast-EnvironmentChange {
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
    }

    $configJson = $config | ConvertTo-Json -Depth 10
    $configJson | Set-Content -Path $configPath -Encoding UTF8

    Write-Success "Configuration created at: $configPath"
    return $configPath
}

# Main execution
try {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "   Java Version Switcher Installer" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""

    # Check for admin privileges and offer to elevate
    $isAdmin = Test-Administrator
    if (-not $isAdmin -and -not $Silent) {
        Write-Host ""
        Write-Warn "Not running as Administrator"
        Write-Host ""
        Write-Host "Administrator privileges are recommended for:" -ForegroundColor Yellow
        Write-Host "  - Installing Java to Program Files" -ForegroundColor Yellow
        Write-Host "  - Setting JAVA_HOME system variable" -ForegroundColor Yellow
        Write-Host "  - Adding Java to system PATH" -ForegroundColor Yellow
        Write-Host ""
        $response = Read-Host "Would you like to restart with administrator privileges? (Y/n)"

        if ($response -eq "" -or $response -eq "y" -or $response -eq "Y") {
            # Build argument list
            $scriptArgs = @()
            if ($Version -ne "latest") { $scriptArgs += "-Version", $Version }
            if ($JavaVersion -ne "21") { $scriptArgs += "-JavaVersion", $JavaVersion }
            if ($InstallDir) { $scriptArgs += "-InstallDir", $InstallDir }
            if ($NoJava) { $scriptArgs += "-NoJava" }
            if ($NoModifyPath) { $scriptArgs += "-NoModifyPath" }

            Invoke-ElevatedScript -Arguments $scriptArgs
        }

        Write-Info "Continuing with limited installation (user-level only)..."
    }

    Initialize-Environment

    $arch = Get-WindowsArchitecture
    Write-Info "Detected architecture: $arch"

    $finalInstallDir = Get-InstallDirectory
    Write-Info "Install directory: $finalInstallDir"

    # Download jv.exe
    $jvPath = DownloadJv -version $Version -arch $arch

    # Check for Java installations
    $javaInstalls = Find-JavaInstallations
    $downloadedJava = $null

    if ($javaInstalls.Count -eq 0 -and -not $NoJava) {
        $downloadJava = $true

        if (-not $Silent) {
            Write-Host ""
            $response = Read-Host "No Java installation found. Download Java $JavaVersion? (Y/n)"
            $downloadJava = ($response -eq "" -or $response -eq "y" -or $response -eq "Y")
        }

        if ($downloadJava) {
            $downloadedJava = DownloadJava -version $JavaVersion -arch $arch
            $javaInstalls += $downloadedJava
        }
    }

    # Install jv.exe to ~/.local/bin
    $binDir = Install-JV -binPath $jvPath -binDir $finalInstallDir

    # Add ~/.local/bin to PATH (user level)
    $pathModified = $false
    if (-not $NoModifyPath) {
        $pathModified = Add-ToPath -directory $binDir
    }

    # Set Java environment variables if Java was downloaded
    $javaEnvSet = $false
    if ($downloadedJava) {
        $javaEnvSet = Set-JavaEnvironment -javaPath $downloadedJava
    }

    # Create config in ~/.config/jv/
    $configPath = Initialize-Config -javaInstallations $javaInstalls

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

    if ($downloadedJava) {
        Write-Success "Java $JavaVersion installed to: $downloadedJava"

        if ($javaEnvSet) {
            Write-Success "JAVA_HOME and system PATH configured"
        } else {
            Write-Warn "JAVA_HOME not set (requires administrator privileges)"
        }
    }

    Write-Host ""
    Write-Host "Installation Summary:" -ForegroundColor Yellow
    Write-Host "  Executable: $binDir\jv.exe" -ForegroundColor Cyan
    Write-Host "  Config:     $configPath" -ForegroundColor Cyan
    if ($javaEnvSet) {
        Write-Host "  JAVA_HOME:  $downloadedJava" -ForegroundColor Cyan
        Write-Host "  PATH:       Includes %JAVA_HOME%\bin" -ForegroundColor Cyan
    }

    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  1. Restart your terminal to reload environment variables"
    Write-Host "     (or run: `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','User'))"
    Write-Host "  2. Run: jv list"

    if ($javaEnvSet) {
        Write-Host "  3. Java is ready to use! (JAVA_HOME already configured)" -ForegroundColor Green
    } else {
        Write-Host "  3. Switch Java version: jv use $JavaVersion  (requires administrator)" -ForegroundColor Cyan
    }

    Write-Host ""
    Write-Host "For help: jv help" -ForegroundColor Gray
    Write-Host ""
}
catch {
    Write-Err "Installation failed: $_"
    Write-Host ""
    Write-Host "For help, visit: https://github.com/$GITHUB_REPO/issues" -ForegroundColor Gray
    exit 1
}
