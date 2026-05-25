# awtrix3-client

A terminal UI client for the [AWTRIX 3](https://blueforcer.github.io/awtrix3/) pixel clock (Ulanzi TC001 and compatible devices), written in Go with [Bubbletea](https://github.com/charmbracelet/bubbletea).

Supports both HTTP REST and MQTT. Distributed as a single static binary — no runtime required.

---

## Installation

### Pre-built binary

Download the latest release for your platform from the [GitHub Releases](https://github.com/terzinnorbert/awtrix3-client/releases) page.

### From source

```sh
go install github.com/terzinnorbert/awtrix3-client@latest
```

---

## Configuration

The device host is resolved in the following priority order:

1. `--host` CLI flag
2. `AWTRIX_HOST` environment variable
3. `AWTRIX_HOST` key in a `.env` file in the current directory

Create a `.env` file (see `.env.example`):

```env
AWTRIX_HOST=192.168.1.100
AWTRIX_MQTT_BROKER=tcp://192.168.1.10:1883
AWTRIX_MQTT_PREFIX=awtrix
```

---

## Usage

### Interactive TUI

Launch the full terminal UI:

```sh
awtrix3-client --host 192.168.1.100
```

#### Global flags

| Flag | Default | Description |
|---|---|---|
| `--host` | — | Device IP or hostname |
| `--mqtt-broker` | — | MQTT broker URL (e.g. `tcp://host:1883`) |
| `--mqtt-prefix` | `awtrix` | MQTT topic prefix |

---

### TUI Tabs

Navigate with number keys `1`–`8`, or click a tab with the mouse.

#### 1 · Dashboard

Live device stats (IP, RAM, battery, WiFi RSSI, uptime, firmware version, temperature, humidity) with auto-refresh every 5 seconds.

Quick controls: previous app, next app, switch to a specific app, power off, reboot.

#### 2 · Apps

Custom app builder: compose text, icon, color, background, duration, repeat, scroll speed, effects, overlays, gradient, and more. Preview the JSON payload before pushing.

Manage the app loop: view all active apps, delete entries, navigate the list with `j`/`k` or scroll wheel.

#### 3 · Notifications

Send one-time notifications with text, color, icon, sound, RTTTL melody, hold, wakeup, stack, and loop-sound options. Forward to additional device IPs. Dismiss the current held notification.

#### 4 · Indicators

Control the three hardware LEDs (upper-right, right-side, lower-right): set color, blink interval, or fade speed. Clear individual indicators or all at once.

#### 5 · Mood Lighting

Set ambient matrix lighting in **RGB color** mode (hex or r,g,b) or **color temperature** mode (1000–10000 K). Adjust brightness with a slider. Live color preview.

#### 6 · Sound

Play audio by **melody filename** (from the device's MELODIES folder) or by **RTTTL string**. Adjust volume before playback.

#### 7 · Settings

Configure display (brightness, auto-brightness, app duration, transition effect, scroll speed, global text color, uppercase), time & date formats, week start day, temperature unit, and global matrix overlay (snow, rain, drizzle, storm, thunder, frost).

#### 8 · System

Power on/off, matrix-only disable, deep sleep timer, reboot, OTA firmware update, reset settings, factory erase.

---

### Mouse support

- **Tab bar**: left-click to switch tabs instantly.
- **Scroll wheel**: navigate the app loop cursor on Dashboard and Apps tabs.
- **Buttons**: left-click any primary action button fires its action.

---

### Non-interactive commands

#### Send a notification

```sh
awtrix3-client notify --text "Message"
```

| Flag | Default | Description |
|---|---|---|
| `--text` | *(required)* | Notification text |
| `--color` | — | Text color as hex, e.g. `#FF0000` |
| `--icon` | — | Icon ID or name |
| `--duration` | device default | Display duration in seconds |
| `--sound` | — | Melody filename (without extension) |
| `--rtttl` | — | RTTTL string to play |
| `--hold` | `false` | Hold until dismissed with the button |
| `--wakeup` | `false` | Wake the device if the matrix is off |
| `--stack` | `false` | Stack with other pending notifications |
| `--loop-sound` | `false` | Loop the sound while shown |
| `--clients` | — | Forward to extra device IPs (comma-separated) |

**Examples:**

```sh
# Bold red alert, held until button press
awtrix3-client notify --text "Motion detected" --color "#FF0000" --hold --wakeup

# Notification with icon and sound, stacked
awtrix3-client notify --text "Build passed" --icon 1234 --sound success --stack

# Play a tune alongside the message (RTTTL format: name:defaults:notes)
awtrix3-client notify --text "Mario" --rtttl "Mario:d=4,o=5,b=200:16e6,16e6,32p,8e6,16c6,8e6,8g6,8p,8g5,8p,8c6,16p,8g5,16p,8e5,16p,8a5,8b5,16a#5,8a5,8g5,16e6,16g6,8a6,16f6,8g6,8e6,16c6,16d6,8b5"

# Forward to multiple devices
awtrix3-client notify --text "Dinner time" --clients 192.168.1.101,192.168.1.102
```

> **RTTTL format:** `name:d=<duration>,o=<octave>,b=<bpm>:note1,note2,...`
> The name prefix is required — strings without it will be silently ignored by the device.
>
> **Ready-to-use tones:**
>
> | Tune | RTTTL string |
> |---|---|
> | Super Mario Bros | `Mario:d=4,o=5,b=200:16e6,16e6,32p,8e6,16c6,8e6,8g6,8p,8g5,8p,8c6,16p,8g5,16p,8e5,16p,8a5,8b5,16a#5,8a5,8g5,16e6,16g6,8a6,16f6,8g6,8e6,16c6,16d6,8b5` |
> | Tetris (Korobeiniki) | `Tetris:d=4,o=5,b=160:e6,8b5,8c6,d6,8c6,8b5,a5,8a5,8c6,e6,8d6,8c6,b5,8b5,8c6,d6,e6,c6,a5,a5` |
> | Star Wars Imperial March | `Imperial:d=4,o=5,b=112:8a4,8a4,8a4,2f4,2c5,8a4,2f4,2c5,1a4` |
> | Nokia ringtone | `Nokia:d=4,o=5,b=225:8e6,8d6,f#5,g#5,8c#6,8b5,d5,e5,8b5,8a5,c#5,e5,2a5` |
> | Zelda secret | `Zelda:d=4,o=5,b=200:8g5,8f#5,8d#5,8a4,8g#4,8e5,8g#5,8c6` |

#### Dismiss the current notification

```sh
awtrix3-client notify dismiss
```

---

## API coverage

| Feature | Endpoint | Tab / command |
|---|---|---|
| Device stats | `GET /api/stats` | Dashboard |
| Effects list | `GET /api/effects` | Apps |
| Transitions list | `GET /api/transitions` | Settings |
| App loop | `GET /api/loop` | Apps, Dashboard |
| Power on/off | `POST /api/power` | System, Dashboard |
| Deep sleep | `POST /api/sleep` | System |
| Play melody | `POST /api/sound` | Sound |
| Play RTTTL | `POST /api/rtttl` | Sound |
| Mood lighting | `POST /api/moodlight` | Mood |
| Indicators 1–3 | `POST /api/indicator1..3` | Indicators |
| Push custom app | `POST /api/custom?name=X` | Apps |
| Delete custom app | `POST /api/custom?name=X` (empty) | Apps |
| Send notification | `POST /api/notify` | Notify tab, `notify` command |
| Dismiss notification | `POST /api/notify/dismiss` | Notify tab, `notify dismiss` |
| Next / previous app | `POST /api/nextapp`, `/previousapp` | Dashboard |
| Switch to app | `POST /api/switch` | Dashboard |
| Get / set settings | `GET/POST /api/settings` | Settings |
| Reboot | `POST /api/reboot` | System |
| OTA update | `POST /api/doupdate` | System |
| Erase flash | `POST /api/erase` | System |
| Reset settings | `POST /api/resetSettings` | System |

MQTT publish is supported for the same set of topics using the `--mqtt-broker` flag.

---

## Building from source

```sh
git clone https://github.com/terzinnorbert/awtrix3-client
cd awtrix3-client
go build -o awtrix3-client .
```

Local multi-platform builds (output in `dist/`):

```sh
make dist
```

Full release build via GoReleaser:

```sh
goreleaser build --snapshot --clean
```

---

## Agent Skill (GitHub Copilot)

This repository ships a [VS Code agent skill](https://code.visualstudio.com/docs/copilot/customization/agent-skills) that instructs GitHub Copilot to push color-coded notifications to your AWTRIX3 display during coding sessions — automatically, without any extra prompting.

### What it does

The agent sends a short message to the pixel clock whenever it:
- Starts a long or complex task
- Completes work successfully
- Encounters an error or build failure
- Needs your attention or input
- Finishes a build or test run

### Prerequisites

- VS Code with the **GitHub Copilot** extension
- `awtrix3-client` binary on your `PATH` (see [Installation](#installation))
- `AWTRIX_HOST` environment variable set to your device IP

### Usage

The skill is loaded automatically — once `AWTRIX_HOST` is set and the binary is on your `PATH`, the agent will start notifying the display on its own.

You can also invoke it explicitly in chat:

```
/awtrix-notify success "Deployment complete"
```

Available event types: `start` · `success` · `error` · `attention` · `build` · `test`

### Skill location

The skill lives at [`.github/skills/awtrix-notify/`](.github/skills/awtrix-notify/) and works on Linux, macOS, and Windows.
