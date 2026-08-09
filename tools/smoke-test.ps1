param(
    [switch]$RunDiagnostic
)

$ErrorActionPreference = 'Stop'
$env:GOTELEMETRY = 'off'
$cachePath = Join-Path $env:TEMP ("synctray-gocache-" + [guid]::NewGuid())
$env:GOCACHE = $cachePath

function Invoke-Go {
    param([string[]]$Arguments)

    & go @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "go $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

$tempExe = Join-Path $env:TEMP 'synctray-smoke.exe'
try {
    Invoke-Go -Arguments @('test', '-race', './...')
    Invoke-Go -Arguments @('vet', './...')
    Invoke-Go -Arguments @('build', '-o', $tempExe, '.')
    if ($RunDiagnostic) {
        & $tempExe check
        if ($LASTEXITCODE -ne 0) {
            throw "Diagnostic command failed with exit code $LASTEXITCODE"
        }
    }
}
finally {
    Remove-Item -LiteralPath $tempExe -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $cachePath -Force -Recurse -ErrorAction SilentlyContinue
}
