[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$pluginId = 'com.timetta.link-preview'
$version = '0.3.0'
$workspace = (Resolve-Path -LiteralPath $PSScriptRoot).Path
$buildRoot = Join-Path $workspace 'build'
$bundleRoot = Join-Path $buildRoot $pluginId
$serverDist = Join-Path $bundleRoot 'server\dist'
$distRoot = Join-Path $workspace 'dist'

foreach ($path in @($buildRoot, $distRoot)) {
    if (Test-Path -LiteralPath $path) {
        $resolved = (Resolve-Path -LiteralPath $path).Path
        if (-not $resolved.StartsWith($workspace, [System.StringComparison]::OrdinalIgnoreCase)) {
            throw "Refusing to remove a path outside the workspace: $resolved"
        }
        Remove-Item -LiteralPath $resolved -Recurse -Force
    }
}

New-Item -ItemType Directory -Path $serverDist -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $bundleRoot 'assets') -Force | Out-Null
New-Item -ItemType Directory -Path $distRoot -Force | Out-Null

Copy-Item -LiteralPath (Join-Path $workspace 'plugin.json') -Destination $bundleRoot
Copy-Item -LiteralPath (Join-Path $workspace 'assets\icon.svg') -Destination (Join-Path $bundleRoot 'assets\icon.svg')

$targets = @(
    @{ OS = 'linux';  Arch = 'amd64'; File = 'plugin-linux-amd64' },
    @{ OS = 'linux';  Arch = 'arm64'; File = 'plugin-linux-arm64' },
    @{ OS = 'darwin'; Arch = 'amd64'; File = 'plugin-darwin-amd64' },
    @{ OS = 'darwin'; Arch = 'arm64'; File = 'plugin-darwin-arm64' },
    @{ OS = 'windows'; Arch = 'amd64'; File = 'plugin-windows-amd64.exe' }
)

$oldGoos = $env:GOOS
$oldGoarch = $env:GOARCH
$oldCgo = $env:CGO_ENABLED
try {
    $env:CGO_ENABLED = '0'
    foreach ($target in $targets) {
        $env:GOOS = $target.OS
        $env:GOARCH = $target.Arch
        $output = Join-Path $serverDist $target.File
        Write-Host "Building $($target.OS)/$($target.Arch)..."
        & go build -trimpath -ldflags '-s -w' -o $output ./server
        if ($LASTEXITCODE -ne 0) {
            throw "Go build failed for $($target.OS)/$($target.Arch)."
        }
    }
}
finally {
    $env:GOOS = $oldGoos
    $env:GOARCH = $oldGoarch
    $env:CGO_ENABLED = $oldCgo
}

$archive = Join-Path $distRoot "$pluginId-$version.tar.gz"
& go run ./tools/bundle -source $bundleRoot -output $archive -root $pluginId
if ($LASTEXITCODE -ne 0) {
    throw 'Failed to create the plugin archive.'
}

Write-Host "Created $archive"
