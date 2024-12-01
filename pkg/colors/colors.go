package colors

import (
	"fmt"
)

type Color int

func (c Color) Code() string {
	return fmt.Sprintf("\x1b[%dm", c)
}

func (c Color) Surround(msg string) string {
	return fmt.Sprintf("%s%s%s", c.Code(), msg, Reset.Code())
}

const (
	Reset Color = iota
	Bold
	Dim
)

const (
	Red Color = iota + 31
	Green
	Yellow
	Blue
	Magenta
	Cyan

	minColor = int(Red)
	maxColor = int(Cyan)
)

// Deterministic random color
func RandD(seed string) Color {
	total := 0
	for _, c := range seed {
		total += int(c) + 'r'
	}

	// bounded by available colors
	value := total%(maxColor-minColor+1) + minColor

	return Color(value)
}
