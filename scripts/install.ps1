# Self-installer for the Evolution DevServices CLI (eds) on Windows.
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/cloud-ru/evolution-devservices-cli/main/scripts/install.ps1 | iex
#   $env:EDS_CLI_VERSION = "v0.2.0"; iwr -useb .../install.ps1 | iex
#
# The script:
#   1. Detects OS / architecture.
#   2. Downloads the matching static binary from a GitHub Release (default)
#      or a custom mirror if EDS_CLI_BASE_URL is set.
#   3. Installs it into %LOCALAPPDATA%\Programs\eds (or EDS_CLI_DIR).
#   4. Adds the directory to the user PATH if missing.
#   5. Verifies the install with `eds version`.
#
# Environment variables (overridable):
#   EDS_CLI_VERSION   - tag/version to download (default: latest)
#   EDS_CLI_REPO      - GitHub "org/repo" to fetch releases from
#                       (default: cloud-ru/evolution-devservices-cli)
#   EDS_CLI_BASE_URL  - base URL of a custom artifact mirror
#   EDS_CLI_DIR       - install directory (default: $env:LOCALAPPDATA\Programs\eds)
#   EDS_CLI_BIN       - binary name (default: eds)
#   EDS_CLI_QUIET     - if set to 1, suppress progress output

param(
    [string]$Version = $env:EDS_CLI_VERSION,
    [string]$Repo    = $env:EDS_CLI_REPO,
    [string]$BaseUrl = $env:EDS_CLI_BASE_URL,
    [string]$Dir     = $env:EDS_CLI_DIR,
    [string]$Bin     = $env:EDS_CLI_BIN,
    [switch]$Quiet   = ($env:EDS_CLI_QUIET -eq "1")
)

if (-not $Version) { $Version = "latest" }
if (-not $Repo)    { $Repo    = "cloud-ru/evolution-devservices-cli" }
if (-not $Dir)     { $Dir     = Join-Path $env:LOCALAPPDATA "Programs\eds" }
if (-not $Bin)     { $Bin     = "eds" }

function Write-Log($msg) {
    if (-not $Quiet) { Write-Host "[eds-cli installer] $msg" }
}

# -- detect platform ----------------------------------------------------------

$archInfo = [System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture
switch ($archInfo) {
    X64  { $ARCH = "amd64" }
    Arm64 { $ARCH = "arm64" }
    default {
        Write-Error "Unsupported architecture: $archInfo"
        exit 1
    }
}

$OS = "windows"
$EXT = ".exe"
$ASSET = "eds-${OS}-${ARCH}${EXT}"

# -- resolve download URL -----------------------------------------------------

if ($BaseUrl) {
    $BaseUrl = $BaseUrl.TrimEnd('/')
    if ($Version -eq "latest") {
        try {
            $Version = (Invoke-WebRequest -Uri "${BaseUrl}/latest" -UseBasicParsing -ErrorAction Stop).Content.Trim()
        } catch {
            Write-Error "Could not resolve 'latest' version from ${BaseUrl}/latest"
            exit 1
        }
    }
    $URL = "${BaseUrl}/${Version}/${ASSET}"
} else {
    if ($Version -eq "latest") {
        $URL = "https://github.com/${Repo}/releases/latest/download/${ASSET}"
    } else {
        $URL = "https://github.com/${Repo}/releases/download/${Version}/${ASSET}"
    }
}

Write-Log "Downloading $URL"

# -- install ------------------------------------------------------------------

if (-not (Test-Path $Dir)) {
    New-Item -ItemType Directory -Path $Dir -Force | Out-Null
}

$Dest = Join-Path $Dir "${Bin}${EXT}"
$Tmp  = Join-Path $env:TEMP ([System.IO.Path]::GetRandomFileName())

try {
    Invoke-WebRequest -Uri $URL -OutFile $Tmp -UseBasicParsing -ErrorAction Stop
} catch {
    Write-Error "Download failed: $_"
    Remove-Item -Path $Tmp -ErrorAction SilentlyContinue
    exit 1
}

Move-Item -Path $Tmp -Destination $Dest -Force

# -- PATH ---------------------------------------------------------------------

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathParts = $userPath -split ";" | Where-Object { $_ -ne "" }
$inPath = $false
foreach ($part in $pathParts) {
    if ((Resolve-Path $part -ErrorAction SilentlyContinue).Path -eq (Resolve-Path $Dir -ErrorAction SilentlyContinue).Path) {
        $inPath = $true
        break
    }
}
if (-not $inPath) {
    $newPath = ($pathParts + $Dir) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Log "Added to user PATH: $Dir"
    # Also update the current session so `eds` is available immediately
    $env:Path = "$Dir;$env:Path"
}

# -- verify -------------------------------------------------------------------

Write-Log "Installed to $Dest"
try {
    $ver = & $Dest version 2>$null
    if ($LASTEXITCODE -ne 0 -or -not $ver) {
        throw "version check failed"
    }
    Write-Log "Verified: $ver"
} catch {
    Write-Error "Installed binary failed to report version"
    exit 1
}

Write-Output $Dest
