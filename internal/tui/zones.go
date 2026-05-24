package tui

// contentStartRow is the absolute terminal row where tab content begins.
// Rows 0-1: header + divider; rows 2-3: tab bar + divider.
const contentStartRow = 4

// clickZone describes a rectangular region of the terminal screen that maps
// to a named action.
type clickZone struct {
	YMin, YMax int
	XMin, XMax int
	ID         string
}

// hitZone returns the ID of the last zone whose rectangle contains (x, y),
// or "" if none matches.  Zones appended later take precedence.
func hitZone(zones []clickZone, x, y int) string {
	for i := len(zones) - 1; i >= 0; i-- {
		z := zones[i]
		if y >= z.YMin && y <= z.YMax && x >= z.XMin && x <= z.XMax {
			return z.ID
		}
	}
	return ""
}

// zoneLine computes the absolute Y of the next line to be written given the
// content built so far (counts "\n" characters already in the buffer).
func zoneLine(contentBuf string) int {
	n := 0
	for _, c := range contentBuf {
		if c == '\n' {
			n++
		}
	}
	return contentStartRow + n
}
