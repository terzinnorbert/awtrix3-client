package api

import "fmt"

// --- Status ---

// GetStats returns general device statistics.
func (c *Client) GetStats() (*Stats, error) {
	var s Stats
	return &s, c.get("/stats", &s)
}

// GetEffects returns the list of available background effects.
func (c *Client) GetEffects() ([]string, error) {
	var v []string
	return v, c.get("/effects", &v)
}

// GetTransitions returns the list of available transition effects.
func (c *Client) GetTransitions() ([]string, error) {
	var v []string
	return v, c.get("/transitions", &v)
}

// GetLoop returns the current app loop list.
func (c *Client) GetLoop() ([]LoopItem, error) {
	var v []LoopItem
	return v, c.get("/loop", &v)
}

// GetScreen returns the raw matrix pixel data as a flat array of 24-bit RGB colors (32×8 = 256 values).
func (c *Client) GetScreen() ([]int, error) {
	var v []int
	return v, c.get("/screen", &v)
}

// --- Power ---

// SetPower turns the matrix on (true) or off (false).
func (c *Client) SetPower(on bool) error {
	return c.post("/power", map[string]bool{"power": on})
}

// Sleep puts the device into deep sleep for the given number of seconds.
func (c *Client) Sleep(seconds int) error {
	return c.post("/sleep", map[string]int{"sleep": seconds})
}

// --- Sound ---

// PlaySound plays a melody by filename (without extension) from the MELODIES folder.
func (c *Client) PlaySound(name string) error {
	return c.post("/sound", map[string]string{"sound": name})
}

// PlayRTTTL plays a sound from a raw RTTTL string.
func (c *Client) PlayRTTTL(rtttl string) error {
	return c.postRaw("/rtttl", rtttl)
}

// --- Mood Lighting ---

// SetMoodLight enables mood lighting with the given settings.
func (c *Client) SetMoodLight(ml MoodLight) error {
	return c.post("/moodlight", ml)
}

// DisableMoodLight disables mood lighting by sending an empty payload.
func (c *Client) DisableMoodLight() error {
	return c.post("/moodlight", nil)
}

// --- Indicators ---

// SetIndicator sets a colored indicator. num must be 1, 2, or 3.
func (c *Client) SetIndicator(num int, ind Indicator) error {
	if num < 1 || num > 3 {
		return fmt.Errorf("indicator number must be 1, 2, or 3")
	}
	return c.post(fmt.Sprintf("/indicator%d", num), ind)
}

// ClearIndicator hides an indicator by sending an empty payload. num must be 1, 2, or 3.
func (c *Client) ClearIndicator(num int) error {
	if num < 1 || num > 3 {
		return fmt.Errorf("indicator number must be 1, 2, or 3")
	}
	return c.post(fmt.Sprintf("/indicator%d", num), nil)
}

// ClearAllIndicators hides all three indicators.
func (c *Client) ClearAllIndicators() error {
	for i := 1; i <= 3; i++ {
		if err := c.ClearIndicator(i); err != nil {
			return err
		}
	}
	return nil
}

// --- Custom Apps ---

// PushCustomApp creates or updates a custom app by name.
func (c *Client) PushCustomApp(name string, app CustomApp) error {
	return c.post("/custom?name="+name, app)
}

// DeleteCustomApp removes a custom app by sending an empty payload.
func (c *Client) DeleteCustomApp(name string) error {
	return c.post("/custom?name="+name, nil)
}

// --- Notifications ---

// SendNotification sends a one-time notification.
func (c *Client) SendNotification(n CustomApp) error {
	return c.post("/notify", n)
}

// DismissNotification dismisses the current held notification.
func (c *Client) DismissNotification() error {
	return c.post("/notify/dismiss", nil)
}

// --- App Navigation ---

// NextApp switches to the next app in the loop.
func (c *Client) NextApp() error {
	return c.post("/nextapp", nil)
}

// PrevApp switches to the previous app in the loop.
func (c *Client) PrevApp() error {
	return c.post("/previousapp", nil)
}

// SwitchApp switches directly to the named app.
func (c *Client) SwitchApp(name string) error {
	return c.post("/switch", map[string]string{"name": name})
}

// --- Settings ---

// GetSettings retrieves the current device settings.
func (c *Client) GetSettings() (*Settings, error) {
	var s Settings
	return &s, c.get("/settings", &s)
}

// SetSettings applies device settings (only non-nil fields are sent).
func (c *Client) SetSettings(s Settings) error {
	return c.post("/settings", s)
}

// --- System ---

// Reboot reboots the device.
func (c *Client) Reboot() error {
	return c.post("/reboot", nil)
}

// DoUpdate triggers an OTA firmware update.
func (c *Client) DoUpdate() error {
	return c.post("/doupdate", nil)
}

// Erase performs a factory reset: formats flash + EEPROM but keeps WiFi settings.
func (c *Client) Erase() error {
	return c.post("/erase", nil)
}

// ResetSettings resets all API settings but keeps WiFi and flash files.
func (c *Client) ResetSettings() error {
	return c.post("/resetSettings", nil)
}
