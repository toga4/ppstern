package ppstern

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	omitKeys = []string{
		"logging.googleapis.com/trace",
		"logging.googleapis.com/spanId",
		"logging.googleapis.com/trace_sampled",
		"caller",
		"stacktrace",
		"level",
		"severity",
		"msg",
		"message",
		"ts",
		"time",
		"timestamp",
	}
)

func ParseAndFormat(in []byte) (string, error) {
	var l Log
	if err := decodeJson(in, &l); err != nil {
		return "", err
	}
	l.PodColor, l.ContainerColor = determineColor(l.PodName)

	var m map[string]any
	if err := decodeJson([]byte(l.Message), &m); err != nil {
		return l.FormatRaw(), nil
	}

	timestamp := extractTimestamp(m)
	level := extractLevel(m)
	message := extractMessage(m)

	for _, key := range omitKeys {
		delete(m, key)
	}

	var b bytes.Buffer
	if timestamp != "" {
		b.WriteString(fmt.Sprintf("%s ", timestamp))
	}
	if level != "" {
		b.WriteString(fmt.Sprintf("%s ", levelColor(level)))
	}
	b.WriteString(fmt.Sprintf("%s %s %s", l.PodColor.Sprint(l.PodName), l.ContainerColor.Sprint(l.ContainerName), message))
	if len(m) > 0 {
		rests, err := json.Marshal(m)
		if err != nil {
			rests = []byte(err.Error())
		}
		b.WriteString(fmt.Sprintf(" %s", rests))
	}
	return b.String(), nil
}

func decodeJson(in []byte, data any) error {
	decoder := json.NewDecoder(bytes.NewReader(in))
	decoder.UseNumber()
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode log: %w", err)
	}
	return nil
}

type Log struct {
	NodeName      string `json:"nodeName"`
	Namespace     string `json:"namespace"`
	PodName       string `json:"podName"`
	ContainerName string `json:"containerName"`
	Message       string `json:"message"`

	PodColor       *color.Color `json:"-"`
	ContainerColor *color.Color `json:"-"`
}

func (l *Log) FormatRaw() string {
	return fmt.Sprintf("%s %s %s", l.PodColor.Sprint(l.PodName), l.ContainerColor.Sprint(l.ContainerName), l.Message)
}

func extractMessage(m map[string]any) string {
	if l, ok := extractAny(m, "msg", "message").(string); ok {
		return l
	}
	return ""
}

func extractLevel(m map[string]any) string {
	if l, ok := extractAny(m, "level", "severity").(string); ok {
		return l
	}
	return ""
}

func extractTimestamp(m map[string]any) string {
	t := extractAny(m, "ts", "time", "timestamp")

	layout := "2006-01-02T15:04:05.000Z07:00"

	var err error
	switch timestamp := t.(type) {
	case string:
		var t time.Time
		if t, err = time.Parse(time.RFC3339Nano, timestamp); err == nil {
			return t.Format(layout)
		}
	case json.Number:
		if strings.Contains(timestamp.String(), ".") {
			var f float64
			if f, err = timestamp.Float64(); err == nil {
				return time.Unix(int64(f), int64((f-float64(int64(f)))*1e9)).Format(layout)
			}
		} else {
			var i int64
			if i, err = timestamp.Int64(); err == nil {
				return time.Unix(i, 0).Format(layout)
			}
		}
	}
	return err.Error()
}

func extractAny(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v
		}
	}
	return nil
}
