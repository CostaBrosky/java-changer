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
    try {
        $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
        switch ($arch) {
            "X64" { return "x64" }
            "X86" { return "x86" }
            "Arm64" { return "arm64" }
            default {
                Write-Warn "Unknown architecture: $arch, defaulting to x64"
                return "x64"
            }
        }
    }
    catch {
        # Fallback for older PowerShell versions
        if ([Environment]::Is64BitOperatingSystem) {
            return "x64"
        }
        else {
            return "x86"
        }
    }
}

# Get install directory
function Get-InstallDirectory {
    if ($InstallDir) {
        return $InstallDir
    }

    # Follow XDG pattern, fallback to C:\tools for Windows users
    $xdgBinHome = $env:XDG_BIN_HOME
    if ($xdgBinHome) {
        return $xdgBinHome
    }

    $localBin = Join-Path $HOME ".local\bin"
    return $localBin
}

# Download jv.exe from GitHub releases
function Download-JV($version, $arch) {
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
function Download-Java($version, $arch) {
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

        Write-Info "Found JDK $($jdkInfo[0].version.openjdk_version)"
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

        # Extract
        $javaInstallDir = Join-Path $HOME ".jv"
        if (-not (Test-Path $javaInstallDir)) {
            New-Item -ItemType Directory -Path $javaInstallDir | Out-Null
        }

        Write-Info "Extracting JDK..."
        Expand-Archive -Path $zipPath -DestinationPath $javaInstallDir -Force

        # Find the extracted directory
        $extractedDirs = Get-ChildItem -Path $javaInstallDir -Directory | Where-Object { $_.Name -match "jdk" }
        if ($extractedDirs.Count -eq 0) {
            throw "Failed to find extracted JDK directory"
        }

        $jdkPath = $extractedDirs[0].FullName

        # Cleanup
        Remove-Item $zipPath -ErrorAction SilentlyContinue

        Write-Success "Java $version installed to: $jdkPath"
        return $jdkPath
    }
    catch {
        throw "Failed to download Java: $_"
    }
}

# Install jv.exe
function Install-JV($binPath, $installDir) {
    Write-Info "Installing jv to $installDir..."

    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    $targetPath = Join-Path $installDir "jv.exe"
    Copy-Item -Path $binPath -Destination $targetPath -Force

    Write-Success "Installed jv.exe to: $targetPath"
    return $installDir
}

# Add directory to user PATH
function Add-ToPath($directory) {
    Write-Info "Updating PATH..."

    $regPath = "HKCU:\Environment"
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    # Check if already in PATH
    $paths = $currentPath -split ";" | Where-Object { $_ -ne "" }
    $normalizedDir = $directory.TrimEnd('\')

    foreach ($p in $paths) {
        if ($p.TrimEnd('\') -eq $normalizedDir) {
            Write-Info "Directory already in PATH"
            return $false
        }
    }

    # Add to PATH
    $newPath = "$normalizedDir;$currentPath"
    Set-ItemProperty -Path $regPath -Name "Path" -Value $newPath

    # Broadcast environment change
    try {
        $HWND_BROADCAST = [IntPtr]0xffff
        $WM_SETTINGCHANGE = 0x1a
        $result = [UIntPtr]::Zero

        Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
public class Win32 {
    [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
    public static extern IntPtr SendMessageTimeout(
        IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam,
        uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
}
"@

        [Win32]::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [UIntPtr]::Zero, "Environment", 2, 5000, [ref]$result) | Out-Null
    }
    catch {
        Write-Warn "Failed to broadcast environment change: $_"
    }

    Write-Success "Added to PATH"
    return $true
}

# Create initial config file
function Initialize-Config($javaInstallations) {
    Write-Info "Creating configuration..."

    $configPath = Join-Path $HOME ".javarc"

    $config = @{
        custom_paths = @($javaInstallations)
        search_paths = @()
    }

    $configJson = $config | ConvertTo-Json -Depth 10
    $configJson | Set-Content -Path $configPath -Encoding UTF8

    Write-Success "Configuration created at: $configPath"
}

# Main execution
try {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "   Java Version Switcher Installer" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""

    Initialize-Environment

    $arch = Get-WindowsArchitecture
    Write-Info "Detected architecture: $arch"

    $finalInstallDir = Get-InstallDirectory
    Write-Info "Install directory: $finalInstallDir"

    # Download jv.exe
    $jvPath = Download-JV -version $Version -arch $arch

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
            $downloadedJava = Download-Java -version $JavaVersion -arch $arch
            $javaInstalls += $downloadedJava
        }
    }

    # Install jv.exe
    $installedDir = Install-JV -binPath $jvPath -installDir $finalInstallDir

    # Update PATH
    if (-not $NoModifyPath) {
        $pathModified = Add-ToPath -directory $installedDir
    }

    # Create config
    Initialize-Config -javaInstallations $javaInstalls

    # Cleanup
    $tempDir = Join-Path $env:TEMP "jv-install"
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue

    # Success message
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "   Installation Complete!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Success "jv installed to: $installedDir"

    if ($downloadedJava) {
        Write-Success "Java $JavaVersion installed to: $downloadedJava"
    }

    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  1. Restart your terminal (or run: `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','User'))"
    Write-Host "  2. Run: jv list"
    Write-Host "  3. Switch Java version: jv use 17  (requires administrator)" -ForegroundColor Cyan
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
