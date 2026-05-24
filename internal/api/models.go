package api

// Stats represents the response from GET /api/stats.
type Stats struct {
	Battery    int     `json:"bat"`
	BatteryRaw int     `json:"bat_raw"`
	Type       int     `json:"type"`
	Humidity   int     `json:"hum"`
	Lux        int     `json:"lux"`
	Matrix     bool    `json:"matrix"`
	IP         string  `json:"ip_address"`
	UID        string  `json:"uid"`
	Uptime     int     `json:"uptime"`
	Wifi       int     `json:"wifi_signal"`
	Messages   int     `json:"messages"`
	Version    string  `json:"version"`
	RAM        int     `json:"ram"`
	Indicator1 bool    `json:"indicator1"`
	Indicator2 bool    `json:"indicator2"`
	Indicator3 bool    `json:"indicator3"`
	App        string  `json:"app"`
	Temp       float64 `json:"temp"`
}

// LoopItem represents an entry in the app loop returned by GET /api/loop.
type LoopItem struct {
	Name string `json:"name"`
}

// CustomApp represents the JSON payload for custom apps and notifications.
// All fields are optional; only set what you need.
type CustomApp struct {
	Text           interface{}            `json:"text,omitempty"`
	TextCase       *int                   `json:"textCase,omitempty"`
	TopText        *bool                  `json:"topText,omitempty"`
	TextOffset     *int                   `json:"textOffset,omitempty"`
	Center         *bool                  `json:"center,omitempty"`
	Color          interface{}            `json:"color,omitempty"`
	Gradient       interface{}            `json:"gradient,omitempty"`
	BlinkText      *int                   `json:"blinkText,omitempty"`
	FadeText       *int                   `json:"fadeText,omitempty"`
	Background     interface{}            `json:"background,omitempty"`
	Rainbow        *bool                  `json:"rainbow,omitempty"`
	Icon           string                 `json:"icon,omitempty"`
	PushIcon       *int                   `json:"pushIcon,omitempty"`
	Repeat         *int                   `json:"repeat,omitempty"`
	Duration       *int                   `json:"duration,omitempty"`
	Hold           *bool                  `json:"hold,omitempty"`
	Sound          string                 `json:"sound,omitempty"`
	Rtttl          string                 `json:"rtttl,omitempty"`
	LoopSound      *bool                  `json:"loopSound,omitempty"`
	Bar            []int                  `json:"bar,omitempty"`
	Line           []int                  `json:"line,omitempty"`
	Autoscale      *bool                  `json:"autoscale,omitempty"`
	Progress       *int                   `json:"progress,omitempty"`
	ProgressC      interface{}            `json:"progressC,omitempty"`
	ProgressBC     interface{}            `json:"progressBC,omitempty"`
	Pos            *int                   `json:"pos,omitempty"`
	Lifetime       *int                   `json:"lifetime,omitempty"`
	LifetimeMode   *int                   `json:"lifetimeMode,omitempty"`
	Stack          *bool                  `json:"stack,omitempty"`
	Wakeup         *bool                  `json:"wakeup,omitempty"`
	NoScroll       *bool                  `json:"noScroll,omitempty"`
	Clients        []string               `json:"clients,omitempty"`
	ScrollSpeed    *int                   `json:"scrollSpeed,omitempty"`
	Effect         string                 `json:"effect,omitempty"`
	EffectSettings map[string]interface{} `json:"effectSettings,omitempty"`
	Save           *bool                  `json:"save,omitempty"`
	Overlay        string                 `json:"overlay,omitempty"`
}

// Indicator represents the payload for POST /api/indicator1..3.
type Indicator struct {
	Color interface{} `json:"color,omitempty"`
	Blink *int        `json:"blink,omitempty"`
	Fade  *int        `json:"fade,omitempty"`
}

// MoodLight represents the payload for POST /api/moodlight.
type MoodLight struct {
	Brightness int         `json:"brightness"`
	Color      interface{} `json:"color,omitempty"`
	Kelvin     *int        `json:"kelvin,omitempty"`
}

// Settings represents the AWTRIX device settings (GET/POST /api/settings).
// Use pointer fields so only explicitly set values are included in the JSON.
type Settings struct {
	ATIME       *int        `json:"ATIME,omitempty"`
	TEFF        *int        `json:"TEFF,omitempty"`
	TSPEED      *int        `json:"TSPEED,omitempty"`
	TCOL        interface{} `json:"TCOL,omitempty"`
	TMODE       *int        `json:"TMODE,omitempty"`
	CHCOL       interface{} `json:"CHCOL,omitempty"`
	CBCOL       interface{} `json:"CBCOL,omitempty"`
	CTCOL       interface{} `json:"CTCOL,omitempty"`
	WD          *bool       `json:"WD,omitempty"`
	WDCA        interface{} `json:"WDCA,omitempty"`
	WDCI        interface{} `json:"WDCI,omitempty"`
	BRI         *int        `json:"BRI,omitempty"`
	ABRI        *bool       `json:"ABRI,omitempty"`
	ATRANS      *bool       `json:"ATRANS,omitempty"`
	CCORRECTION []int       `json:"CCORRECTION,omitempty"`
	CTEMP       []int       `json:"CTEMP,omitempty"`
	TFORMAT     string      `json:"TFORMAT,omitempty"`
	DFORMAT     string      `json:"DFORMAT,omitempty"`
	SOM         *bool       `json:"SOM,omitempty"`
	CEL         *bool       `json:"CEL,omitempty"`
	BLOCKN      *bool       `json:"BLOCKN,omitempty"`
	UPPERCASE   *bool       `json:"UPPERCASE,omitempty"`
	TIME_COL    interface{} `json:"TIME_COL,omitempty"`
	DATE_COL    interface{} `json:"DATE_COL,omitempty"`
	TEMP_COL    interface{} `json:"TEMP_COL,omitempty"`
	HUM_COL     interface{} `json:"HUM_COL,omitempty"`
	BAT_COL     interface{} `json:"BAT_COL,omitempty"`
	SSPEED      *int        `json:"SSPEED,omitempty"`
	TIM         *bool       `json:"TIM,omitempty"`
	DAT         *bool       `json:"DAT,omitempty"`
	HUM         *bool       `json:"HUM,omitempty"`
	TEMP        *bool       `json:"TEMP,omitempty"`
	BAT         *bool       `json:"BAT,omitempty"`
	MATP        *bool       `json:"MATP,omitempty"`
	VOL         *int        `json:"VOL,omitempty"`
	OVERLAY     string      `json:"OVERLAY,omitempty"`
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool { return &b }

// IntPtr returns a pointer to the given int value.
func IntPtr(i int) *int { return &i }
