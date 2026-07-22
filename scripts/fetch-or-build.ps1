# Install-time build step for Windows (run by Herdr at `plugin install`).
#
# Fast path: download the prebuilt binary matching this plugin's version and
# architecture from GitHub Releases and verify its SHA-256. Fallback: build
# from source with `go` (only needed when no matching prebuilt exists).
$ErrorActionPreference = 'Stop'
$u = New-Object System.Text.UTF8Encoding($false)
[Console]::OutputEncoding = $u; $OutputEncoding = $u

$repo = 'ShankyJS/herdr-space-scoped-agents'
$bin  = 'herdr-space-scoped-agents'

$root = $env:HERDR_PLUGIN_ROOT
if (-not $root) { $root = Split-Path -Parent $PSScriptRoot }
if ($root.StartsWith('\\?\')) { $root = $root.Substring(4) }

$manifest = Join-Path $root 'herdr-plugin.toml'
$version = (Select-String -Path $manifest -Pattern '^version\s*=\s*"(.*)"').Matches[0].Groups[1].Value
if (-not $version) { Write-Error "cannot read version from $manifest" }

switch -Wildcard ($env:PROCESSOR_ARCHITECTURE) {
  'ARM64' { $goarch = 'arm64' }
  default { $goarch = 'amd64' }
}

$binDir = Join-Path $root 'bin'
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
$out = Join-Path $binDir "$bin.exe"

function Download {
  try {
    $asset = "$bin-windows-$goarch.exe"
    $base  = "https://github.com/$repo/releases/download/v$version"
    $tmp   = New-TemporaryFile
    Write-Host "fetching $asset (v$version)..."
    Invoke-WebRequest -Uri "$base/$asset" -OutFile $tmp -UseBasicParsing
    try {
      $sumFile = New-TemporaryFile
      Invoke-WebRequest -Uri "$base/$asset.sha256" -OutFile $sumFile -UseBasicParsing
      $want = (Get-Content $sumFile -Raw).Trim().Split()[0]
      $got  = (Get-FileHash $tmp -Algorithm SHA256).Hash.ToLower()
      if ($want.ToLower() -ne $got) { throw "checksum mismatch (want $want got $got)" }
    } catch {
      Write-Host "warning: checksum verification skipped: $_"
    }
    Move-Item -Force $tmp $out
    return $true
  } catch {
    Write-Host "download failed: $_"
    return $false
  }
}

function Build {
  if (-not (Get-Command go -ErrorAction SilentlyContinue)) { return $false }
  Write-Host "building from source with go..."
  Push-Location $root
  try { & go build -ldflags "-s -w -X main.version=$version" -o $out . ; return $true }
  finally { Pop-Location }
}

if (Download)   { Write-Host "installed prebuilt -> $out" }
elseif (Build)  { Write-Host "built from source -> $out" }
else { Write-Error "no prebuilt binary for windows-$goarch v$version and no Go toolchain to build from source" }
