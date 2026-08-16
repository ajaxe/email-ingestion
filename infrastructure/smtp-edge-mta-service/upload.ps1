[CmdletBinding()]
param(
    [string]$IdentityFile = "$HOME\.ssh\email_ingestion_mta_ec2",
    [string]$TerraformDir = "..\smtp-edge-mta",
    [string]$RemoteUser = "ec2-user",
    [string]$RemoteDest = "~/"
)

$ErrorActionPreference = "Stop"

# Determine script base directory robustly to be agnostic of current working directory
$ScriptDir = $PSScriptRoot
if ([string]::IsNullOrWhiteSpace($ScriptDir)) {
    $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " MTA Service Artifact Upload Script" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Resolve and expand paths relative to script base directory
if ($IdentityFile.StartsWith("~")) {
    $IdentityFile = $IdentityFile -replace "^~", $HOME
}
if (-not [System.IO.Path]::IsPathRooted($IdentityFile)) {
    $IdentityFile = [System.IO.Path]::GetFullPath((Join-Path $ScriptDir $IdentityFile))
} else {
    $IdentityFile = [System.IO.Path]::GetFullPath($IdentityFile)
}

if (-not [System.IO.Path]::IsPathRooted($TerraformDir)) {
    $TerraformPath = [System.IO.Path]::GetFullPath((Join-Path $ScriptDir $TerraformDir))
} else {
    $TerraformPath = [System.IO.Path]::GetFullPath($TerraformDir)
}

$BinaryPath    = [System.IO.Path]::GetFullPath((Join-Path $ScriptDir "output\email-ingest-arm64"))
$SetupPath     = [System.IO.Path]::GetFullPath((Join-Path $ScriptDir "setup.sh"))
$UninstallPath = [System.IO.Path]::GetFullPath((Join-Path $ScriptDir "uninstall.sh"))
$ServicePath   = [System.IO.Path]::GetFullPath((Join-Path $ScriptDir "email-ingest.service"))
$ConfigPath    = [System.IO.Path]::GetFullPath((Join-Path $ScriptDir "config.yaml"))

# 2. Validate prerequisites
Write-Host "[1/4] Checking local prerequisites..." -ForegroundColor Yellow

if (-not (Test-Path -Path $IdentityFile -PathType Leaf)) {
    Write-Error "SSH identity key not found at: $IdentityFile"
    exit 1
}

$Artifacts = @($BinaryPath, $SetupPath, $UninstallPath, $ServicePath, $ConfigPath)
foreach ($file in $Artifacts) {
    if (-not (Test-Path -Path $file -PathType Leaf)) {
        Write-Error "Required artifact file not found at: $file"
        exit 1
    }
}
Write-Host "  - All required local artifacts and SSH key verified." -ForegroundColor Green

# 3. Query Terraform Public IP
Write-Host "[2/4] Querying remote server public IP from Terraform..." -ForegroundColor Yellow
if (-not (Test-Path -Path $TerraformPath -PathType Container)) {
    Write-Error "Terraform directory not found at: $TerraformPath"
    exit 1
}

if (-not (Get-Command terraform -ErrorAction SilentlyContinue)) {
    Write-Error "'terraform' CLI tool was not found in PATH."
    exit 1
}

try {
    $PublicIp = (terraform -chdir="$TerraformPath" output -raw public_ip 2>$null).Trim()
} catch {
    $PublicIp = ""
}

if ([string]::IsNullOrWhiteSpace($PublicIp)) {
    Write-Error "Failed to retrieve public IP from terraform output in '$TerraformPath'."
    exit 1
}

Write-Host "  - Remote Server IP: $PublicIp" -ForegroundColor Green

# 4. Upload Files via SCP
Write-Host "[3/4] Uploading artifacts via SCP to ${RemoteUser}@${PublicIp}:${RemoteDest}..." -ForegroundColor Yellow

$scpArgs = @(
    "-i", $IdentityFile,
    $BinaryPath,
    $SetupPath,
    $UninstallPath,
    $ServicePath,
    $ConfigPath,
    "${RemoteUser}@${PublicIp}:${RemoteDest}"
)

Write-Host "  - Executing: scp -i $IdentityFile output\email-ingest-arm64 setup.sh uninstall.sh email-ingest.service config.yaml ${RemoteUser}@${PublicIp}:${RemoteDest}" -ForegroundColor Gray

& scp.exe @scpArgs

if ($LASTEXITCODE -ne 0) {
    Write-Error "SCP file transfer failed with exit code $LASTEXITCODE."
    exit $LASTEXITCODE
}

Write-Host "[4/4] Upload completed successfully!" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " Next Steps (Run on remote server):" -ForegroundColor Cyan
Write-Host "  1. SSH into instance:" -ForegroundColor White
Write-Host "     ssh -i $IdentityFile ${RemoteUser}@$PublicIp" -ForegroundColor Gray
Write-Host "  2. Make setup script executable and run setup:" -ForegroundColor White
Write-Host "     chmod +x setup.sh" -ForegroundColor Gray
Write-Host "     sudo ./setup.sh" -ForegroundColor Gray
Write-Host "  3. (Optional) To revert changes and uninstall service:" -ForegroundColor White
Write-Host "     chmod +x uninstall.sh" -ForegroundColor Gray
Write-Host "     sudo ./uninstall.sh" -ForegroundColor Gray
Write-Host "==========================================================" -ForegroundColor Cyan
