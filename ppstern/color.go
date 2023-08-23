package ppstern

import (
	"hash/fnv"
	"strings"

	"github.com/fatih/color"
)

var colorList = [][2]*color.Color{
	{color.New(color.FgHiCyan), color.New(color.FgCyan)},
	{color.New(color.FgHiGreen), color.New(color.FgGreen)},
	{color.New(color.FgHiMagenta), color.New(color.FgMagenta)},
	{color.New(color.FgHiYellow), color.New(color.FgYellow)},
	{color.New(color.FgHiBlue), color.New(color.FgBlue)},
	{color.New(color.FgHiRed), color.New(color.FgRed)},
}

func determineColor(podName string) (podColor, containerColor *color.Color) {
	hash := fnv.New32()
	_, _ = hash.Write([]byte(podName))
	idx := hash.Sum32() % uint32(len(colorList))

	colors := colorList[idx]
	return colors[0], colors[1]
}

func levelColor(level string) string {
	var levelColor *color.Color
	switch strings.ToLower(level) {
	case "debug":
		levelColor = color.New(color.FgMagenta)
	case "info":
		levelColor = color.New(color.FgBlue)
	case "warn", "warning":
		levelColor = color.New(color.FgYellow)
	case "error", "dpanic", "panic":
		levelColor = color.New(color.FgRed)
	case "fatal", "critical":
		levelColor = color.New(color.FgCyan)
	default:
		return level
	}
	return levelColor.Sprint(level)
}
