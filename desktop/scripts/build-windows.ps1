param(
    [string]$ApiBaseUrl = "",
    [switch]$AllowLocalApi
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$defaultApiBaseUrl = "https://dashpoint.my.id/api/v1"
$requiredWailsVersion = "v2.13.0"
$wailsPackage = "github.com/wailsapp/wails/v2/cmd/wails@$requiredWailsVersion"
$forbiddenPatterns = @(
    "-----BEGIN [A-Z ]*PRIVATE KEY-----",
    "JWT_SECRET=",
    "POSTGRES_PASSWORD=",
    "VPS_SSH_KEY=",
    "DATABASE_URL=postgres://"
)

function Assert-Command {
    param(
        [string]$Name,
        [string]$InstallHint
    )

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "$Name was not found on PATH. $InstallHint"
    }
}

function Invoke-Checked {
    param(
        [string]$Command,
        [string[]]$Arguments
    )

    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$Command exited with status $LASTEXITCODE"
    }
}

function Get-ToolVersion {
    param(
        [string]$Command,
        [string[]]$Arguments,
        [string]$Pattern
    )

    $output = (& $Command @Arguments 2>&1) -join [Environment]::NewLine
    if ($LASTEXITCODE -ne 0 -or $output -notmatch $Pattern) {
        throw "Unable to determine the version of $Command. Output: $output"
    }

    return [version]$Matches[1]
}

function Get-WailsVersion {
    param(
        [string]$Command
    )

    $output = (& $Command "version" 2>&1) -join [Environment]::NewLine
    if ($LASTEXITCODE -ne 0) {
        return $null
    }

    if ($output -notmatch "\bv([0-9]+\.[0-9]+\.[0-9]+)\b") {
        return $null
    }

    return "v$($Matches[1])"
}

function Get-GoBinPath {
    $goBin = (& "go" "env" "GOBIN").Trim()
    if (-not [string]::IsNullOrWhiteSpace($goBin)) {
        return $goBin
    }

    $goPath = (& "go" "env" "GOPATH").Trim()
    if ([string]::IsNullOrWhiteSpace($goPath)) {
        throw "Go returned an empty GOPATH; cannot locate the installed Wails CLI."
    }

    return Join-Path (($goPath -split ";")[0]) "bin"
}

function Resolve-ApiBaseUrl {
    param(
        [string]$RequestedUrl,
        [switch]$AllowLocal
    )

    $resolvedUrl = if ([string]::IsNullOrWhiteSpace($RequestedUrl)) {
        $defaultApiBaseUrl
    } else {
        $RequestedUrl.Trim()
    }

    try {
        $uri = [Uri]$resolvedUrl
    } catch {
        throw "API URL is not valid: $resolvedUrl"
    }

    $isLocal = $uri.Host -in @("localhost", "127.0.0.1", "::1")
    if ($isLocal) {
        if (-not $AllowLocal) {
            throw "Local API URLs require -AllowLocalApi."
        }
        if ($uri.Scheme -ne "http" -and $uri.Scheme -ne "https") {
            throw "Local API URLs must use HTTP or HTTPS."
        }
    } elseif ($uri.Scheme -ne "https") {
        throw "Production API URLs must use HTTPS."
    }

    return $resolvedUrl
}

function Assert-FileDoesNotContainPrivateConfig {
    param(
        [string]$Path,
        [string]$Description
    )

    if ((Get-Item -LiteralPath $Path).Length -gt 100MB) {
        throw "$Description is unexpectedly large to scan safely: $Path"
    }

    $content = [Text.Encoding]::ASCII.GetString([IO.File]::ReadAllBytes($Path))
    foreach ($pattern in $forbiddenPatterns) {
        if ($content -match $pattern) {
            throw "$Description contains a forbidden private configuration pattern: $pattern"
        }
    }
}

function Assert-DesktopAssetsSafe {
    $distPath = Join-Path $repoRoot "frontend\dist"
    if (-not (Test-Path -LiteralPath $distPath -PathType Container)) {
        throw "Desktop frontend output was not created: $distPath"
    }

    $envFiles = Get-ChildItem -LiteralPath $distPath -Recurse -Force -File |
        Where-Object { $_.Name -like ".env*" }
    if ($envFiles) {
        throw "Desktop assets contain environment files: $($envFiles.FullName -join ', ')"
    }

    foreach ($file in (Get-ChildItem -LiteralPath $distPath -Recurse -Force -File)) {
        Assert-FileDoesNotContainPrivateConfig $file.FullName "Desktop asset"
    }
}

Push-Location $repoRoot
try {
    $resolvedApiUrl = Resolve-ApiBaseUrl $ApiBaseUrl -AllowLocal:$AllowLocalApi

    Assert-Command "go" "Install Go 1.25 or newer."
    Assert-Command "node" "Install Node.js 22 or newer."
    Assert-Command "npm" "Install npm compatible with your Node.js version."

    $goVersion = Get-ToolVersion "go" @("version") "go([0-9]+\.[0-9]+(?:\.[0-9]+)?)"
    if ($goVersion -lt [version]"1.25.0") {
        throw "Go $goVersion is too old; Go 1.25 or newer is required."
    }

    $nodeVersion = Get-ToolVersion "node" @("--version") "v([0-9]+\.[0-9]+\.[0-9]+)"
    if ($nodeVersion.Major -lt 22) {
        throw "Node.js $nodeVersion is too old; Node.js 22 or newer is required."
    }

    $npmVersion = Get-ToolVersion "npm" @("--version") "([0-9]+\.[0-9]+\.[0-9]+)"
    Write-Host "Using Go $goVersion, Node.js $nodeVersion, npm $npmVersion"

    $wailsCommandInfo = Get-Command "wails" -ErrorAction SilentlyContinue
    $wailsCommand = if ($null -eq $wailsCommandInfo) { $null } else { $wailsCommandInfo.Path }
    $installedWailsVersion = if ($null -eq $wailsCommand) { $null } else { Get-WailsVersion $wailsCommand }

    if ($installedWailsVersion -ne $requiredWailsVersion) {
        Write-Host "Installing Wails CLI $requiredWailsVersion for this build..."
        Invoke-Checked "go" @("install", $wailsPackage)

        $wailsCommand = Join-Path (Get-GoBinPath) "wails.exe"
        if (-not (Test-Path -LiteralPath $wailsCommand -PathType Leaf)) {
            throw "Wails installation completed but $wailsCommand was not produced."
        }

        $installedWailsVersion = Get-WailsVersion $wailsCommand
        if ($installedWailsVersion -ne $requiredWailsVersion) {
            throw "Installed Wails CLI version $installedWailsVersion is not $requiredWailsVersion."
        }
    }

    $env:NEXT_PUBLIC_API_URL = $resolvedApiUrl

    Write-Host "Building desktop app with API $resolvedApiUrl"
    Invoke-Checked "go" @("mod", "download")
    Invoke-Checked $wailsCommand @("build", "-clean", "-f")

    Assert-DesktopAssetsSafe

    $outputPath = Join-Path $repoRoot "build\bin\DashPoint.exe"
    if (-not (Test-Path -LiteralPath $outputPath -PathType Leaf)) {
        throw "Wails completed without producing $outputPath"
    }
    Assert-FileDoesNotContainPrivateConfig $outputPath "Generated executable"

    Write-Host "Built $outputPath"
}
finally {
    Pop-Location
}
