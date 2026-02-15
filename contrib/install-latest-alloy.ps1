$ErrorActionPreference = "Stop"
$installDir = "C:\Program Files\Alloy"
$githubRepo = "grafana/alloy"

Write-Host "`nFetching latest Alloy release info from GitHub..."
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$githubRepo/releases/latest" `
    -Headers @{ "User-Agent" = "PowerShell" }

$asset = $release.assets | Where-Object { $_.name -eq "alloy-windows-amd64.exe.zip" } | Select-Object -First 1

if (-not $asset) {
    Write-Error "Could not find alloy-windows-amd64.exe.zip in the latest release!"
    exit 1
}

$tempZip = "$env:TEMP\alloy-latest.zip"
Write-Host "Downloading: $($asset.browser_download_url)"
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tempZip

Write-Host "Extracting to $installDir ..."
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}
Expand-Archive -Path $tempZip -DestinationPath $installDir -Force

$sysPath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
if ($sysPath -notmatch [regex]::Escape($installDir)) {
    [System.Environment]::SetEnvironmentVariable("Path", "$sysPath;$installDir", "Machine")
    Write-Host "Added $installDir to system PATH."
} else {
    Write-Host "$installDir is already in PATH."
}

Write-Host "`nAlloy binary is now installed at $installDir."
Write-Host "You can run it with: alloy.exe"
Write-Host "To start as service or autostart, please configure according to your environment."
