# notify.ps1 — send a notification to the AWTRIX3 pixel display
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
#
# The binary is resolved from PATH; set AWTRIX_HOST or --host to target device.

param(
    [Parameter(Position = 0, Mandatory = $true)]
    [string]$EventType,

    [Parameter(Position = 1, Mandatory = $true)]
    [string]$Message
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Verify the binary is available
if (-not (Get-Command awtrix3-client -ErrorAction SilentlyContinue)) {
    Write-Error "awtrix3-client not found on PATH.`nInstall with: go install github.com/terzi/awtrix3-client@latest"
    exit 1
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

# Echo the resolved command so the agent can log what was sent
Write-Host ("+ awtrix3-client " + ($invokeArgs -join ' ')) -ForegroundColor DarkGray

& awtrix3-client @invokeArgs
exit $LASTEXITCODE
