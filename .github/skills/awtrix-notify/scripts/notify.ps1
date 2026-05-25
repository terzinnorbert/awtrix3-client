# notify.ps1 — send a notification to the AWTRIX3 pixel display
#
# No installation required — uses "go run" to fetch and execute the client.
# Requires: Go 1.21+  (https://go.dev/dl/)
#
# Usage (positional):  .\notify.ps1 start "Planning migration"
# Usage (named):       .\notify.ps1 -EventType success -Message "Build complete"
#
# Event types:
#   start      — starting a long/complex task        (yellow)
#   success    — task completed successfully          (green)
#   error      — task failed or error encountered     (red, held)
#   attention  — user input or attention required     (orange, held)
#   build      — build result                         (blue)
#   test       — test run result                      (purple)

param(
    [Parameter(Position = 0, Mandatory = $true)]
    [string]$EventType,

    [Parameter(Position = 1, Mandatory = $true)]
    [string]$Message
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Verify Go is available
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go not found on PATH.`nInstall Go 1.21+ from https://go.dev/dl/"
    exit 1
}

# Resolve AWTRIX_HOST — prompt interactively and persist if not set
$hostValue = [System.Environment]::GetEnvironmentVariable("AWTRIX_HOST")
if (-not $hostValue) {
    if ([System.Environment]::UserInteractive -and -not [Console]::IsInputRedirected) {
        $hostValue = Read-Host "AWTRIX_HOST not set. Enter device IP address"
        if (-not $hostValue) {
            Write-Error "No IP address provided."
            exit 1
        }
        $env:AWTRIX_HOST = $hostValue
        [System.Environment]::SetEnvironmentVariable("AWTRIX_HOST", $hostValue, "User")
        Write-Host "Saved AWTRIX_HOST=$hostValue to user environment variables." -ForegroundColor DarkGray
    } else {
        # Non-interactive (e.g. called from an AI agent): exit with a distinct code
        # so the caller knows to ask the user for the IP first.
        Write-Host "Error: AWTRIX_HOST is not set." -ForegroundColor Red
        Write-Host "Set it with:" -ForegroundColor Yellow
        Write-Host '  $env:AWTRIX_HOST = "192.168.x.x"'
        Write-Host '  [Environment]::SetEnvironmentVariable("AWTRIX_HOST", "192.168.x.x", "User")'
        exit 2
    }
}

# Truncate message to 30 characters — keeps the scrolling pixel display readable
if ($Message.Length -gt 30) {
    $Message = $Message.Substring(0, 30)
}

# Map event type → color and flags
$Color = '#FFFFFF'
$Hold  = $false

switch ($EventType.ToLower()) {
    'start'                              { $Color = '#FFAA00' }
    'success'                            { $Color = '#00FF00' }
    { $_ -in 'error','fail','failure' }  { $Color = '#FF0000'; $Hold = $true }
    { $_ -in 'attention','input' }       { $Color = '#FF8800'; $Hold = $true }
    'build'                              { $Color = '#00AAFF' }
    'test'                               { $Color = '#AA44FF' }
    default                              { Write-Warning "Unknown event type '$EventType', using white." }
}

# Build argument list
$invokeArgs = @('notify', '--text', $Message, '--color', $Color, '--wakeup')
if ($Hold) { $invokeArgs += '--hold' }

# Echo the resolved command (properly quoted) so the agent can log what was sent
$quoted = $invokeArgs | ForEach-Object { if ($_ -match '\s') { "'$_'" } else { $_ } }
Write-Host ("+ go run github.com/terzinnorbert/awtrix3-client@latest " + ($quoted -join ' ')) -ForegroundColor DarkGray

& go run github.com/terzinnorbert/awtrix3-client@latest @invokeArgs
exit $LASTEXITCODE
