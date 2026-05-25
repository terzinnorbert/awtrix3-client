package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/terzinnorbert/awtrix3-client/internal/api"
	"github.com/terzinnorbert/awtrix3-client/internal/tui"
)

var (
	flagHosts      []string
	flagMQTTBroker string
	flagMQTTPrefix string

	// notify subcommand flags
	notifyText      string
	notifyColor     string
	notifyIcon      string
	notifyDuration  int
	notifySound     string
	notifyRTTTL     string
	notifyHold      bool
	notifyWakeup    bool
	notifyStack     bool
	notifyLoopSound bool
	notifyClients   []string
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Send a notification to the AWTRIX device",
	Long: `Send a one-time notification directly without launching the TUI.

RTTTL format: name:d=<duration>,o=<octave>,b=<bpm>:notes
The name prefix is required — strings without it are silently ignored.

Examples:
  awtrix3-client notify --text "Motion detected" --color "#FF0000" --hold
  awtrix3-client notify --text "Hello" --icon 1234 --duration 5 --stack
  awtrix3-client notify --text "Alert" --sound alarm --wakeup
  awtrix3-client notify --text "Mario"     --rtttl "Mario:d=4,o=5,b=200:16e6,16e6,32p,8e6,16c6,8e6,8g6,8p,8g5,8p,8c6,16p,8g5,16p,8e5,16p,8a5,8b5,16a#5,8a5,8g5,16e6,16g6,8a6,16f6,8g6,8e6,16c6,16d6,8b5"
  awtrix3-client notify --text "Tetris"    --rtttl "Tetris:d=4,o=5,b=160:e6,8b5,8c6,d6,8c6,8b5,a5,8a5,8c6,e6,8d6,8c6,b5,8b5,8c6,d6,e6,c6,a5,a5"
  awtrix3-client notify --text "Star Wars" --rtttl "Imperial:d=4,o=5,b=112:8a4,8a4,8a4,2f4,2c5,8a4,2f4,2c5,1a4"
  awtrix3-client notify --text "Zelda"     --rtttl "Zelda:d=4,o=5,b=200:8g5,8f#5,8d#5,8a4,8g#4,8e5,8g#5,8c6"`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return runNotify()
	},
}

var dismissCmd = &cobra.Command{
	Use:   "dismiss",
	Short: "Dismiss the current held notification",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDismiss()
	},
}

var rootCmd = &cobra.Command{
	Use:   "awtrix3-client",
	Short: "Terminal UI client for AWTRIX 3 pixel clock",
	Long: `A feature-rich terminal UI for managing your AWTRIX 3 device.
Supports custom apps, notifications, indicators, mood lighting, sound, and more.

Configuration (in priority order):
  1. --host flag (comma-separated or repeated for multiple devices)
  2. AWTRIX_HOST environment variable (comma-separated for multiple)
  3. .env file in the current directory`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&flagHosts, "host", nil, "AWTRIX 3 device IP or hostname; comma-separated or repeated for multiple devices")
	rootCmd.PersistentFlags().StringVar(&flagMQTTBroker, "mqtt-broker", "", "MQTT broker URL (e.g. tcp://192.168.1.10:1883)")
	rootCmd.PersistentFlags().StringVar(&flagMQTTPrefix, "mqtt-prefix", "awtrix", "MQTT topic prefix")

	// notify send flags
	notifyCmd.Flags().StringVar(&notifyText, "text", "", "Notification text (required)")
	notifyCmd.Flags().StringVar(&notifyColor, "color", "", "Text color as hex (e.g. #FF0000)")
	notifyCmd.Flags().StringVar(&notifyIcon, "icon", "", "Icon ID or name")
	notifyCmd.Flags().IntVar(&notifyDuration, "duration", 0, "Display duration in seconds (0 = device default)")
	notifyCmd.Flags().StringVar(&notifySound, "sound", "", "Melody filename to play (without extension)")
	notifyCmd.Flags().StringVar(&notifyRTTTL, "rtttl", "", "RTTTL string to play")
	notifyCmd.Flags().BoolVar(&notifyHold, "hold", false, "Hold notification until dismissed with button")
	notifyCmd.Flags().BoolVar(&notifyWakeup, "wakeup", false, "Wake device if matrix is off")
	notifyCmd.Flags().BoolVar(&notifyStack, "stack", false, "Stack notification with others")
	notifyCmd.Flags().BoolVar(&notifyLoopSound, "loop-sound", false, "Loop the sound while notification is shown")
	notifyCmd.Flags().StringSliceVar(&notifyClients, "clients", nil, "Forward to additional device IPs (comma-separated or repeated flag)")
	_ = notifyCmd.MarkFlagRequired("text")

	notifyCmd.AddCommand(dismissCmd)
	rootCmd.AddCommand(notifyCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTUI() error {
	// Load .env file if present (errors are silently ignored so the app still works
	// when no .env file exists).
	_ = godotenv.Load()

	hosts := resolveHosts()
	if len(hosts) == 0 {
		fmt.Fprintln(os.Stderr, `Error: AWTRIX host not configured.

Set it via one of:
  --host 192.168.1.100
  AWTRIX_HOST=192.168.1.100 in environment
  AWTRIX_HOST=192.168.1.100 in .env file`)
		return fmt.Errorf("no host configured")
	}

	client := api.NewClient(hosts[0])

	// Optionally wire up MQTT client (non-fatal if connection fails).
	mqttBroker := resolveMQTTBroker()
	if mqttBroker != "" {
		prefix := resolveMQTTPrefix()
		_, err := api.NewMQTTClient(mqttBroker, prefix, "awtrix3-tui")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: MQTT connection failed: %v\n", err)
		}
		// Note: the MQTT client is stored for future MQTT-based publish support.
		// Currently the TUI uses HTTP; MQTT can be wired per-tab as a future enhancement.
	}

	model := tui.NewAppModel(client)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

func runNotify() error {
	_ = godotenv.Load()

	hosts := resolveHosts()
	if len(hosts) == 0 {
		return fmt.Errorf("no host configured: use --host, AWTRIX_HOST env var, or .env file")
	}

	if notifySound != "" && notifyRTTTL != "" {
		return fmt.Errorf("--sound and --rtttl are mutually exclusive")
	}

	client := api.NewMultiClient(hosts)

	n := api.CustomApp{
		Text:  notifyText,
		Icon:  notifyIcon,
		Sound: notifySound,
		Rtttl: notifyRTTTL,
	}
	if notifyColor != "" {
		n.Color = notifyColor
	}
	if notifyDuration > 0 {
		n.Duration = api.IntPtr(notifyDuration)
	}
	if notifyHold {
		n.Hold = api.BoolPtr(true)
	}
	if notifyWakeup {
		n.Wakeup = api.BoolPtr(true)
	}
	if notifyStack {
		n.Stack = api.BoolPtr(true)
	}
	if notifyLoopSound {
		n.LoopSound = api.BoolPtr(true)
	}
	if len(notifyClients) > 0 {
		n.Clients = notifyClients
	}

	if err := client.SendNotification(n); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	fmt.Println("Notification sent.")
	return nil
}

func runDismiss() error {
	_ = godotenv.Load()

	hosts := resolveHosts()
	if len(hosts) == 0 {
		return fmt.Errorf("no host configured: use --host, AWTRIX_HOST env var, or .env file")
	}

	if err := api.NewMultiClient(hosts).DismissNotification(); err != nil {
		return fmt.Errorf("failed to dismiss notification: %w", err)
	}

	fmt.Println("Notification dismissed.")
	return nil
}

// resolveHosts returns all configured AWTRIX hosts from flag > env > (already loaded .env).
// The --host flag and AWTRIX_HOST env var both accept comma-separated lists.
func resolveHosts() []string {
	if len(flagHosts) > 0 {
		return flagHosts
	}
	if v := os.Getenv("AWTRIX_HOST"); v != "" {
		var hosts []string
		for _, h := range strings.Split(v, ",") {
			if h = strings.TrimSpace(h); h != "" {
				hosts = append(hosts, h)
			}
		}
		return hosts
	}
	return nil
}

// resolveMQTTBroker returns the MQTT broker URL from flag > env.
func resolveMQTTBroker() string {
	if flagMQTTBroker != "" {
		return flagMQTTBroker
	}
	return os.Getenv("AWTRIX_MQTT_BROKER")
}

// resolveMQTTPrefix returns the MQTT prefix from flag > env, defaulting to "awtrix".
func resolveMQTTPrefix() string {
	if flagMQTTPrefix != "" {
		return flagMQTTPrefix
	}
	if v := os.Getenv("AWTRIX_MQTT_PREFIX"); v != "" {
		return v
	}
	return "awtrix"
}
