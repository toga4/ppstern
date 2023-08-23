package ppstern

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

var jsonFormatter = NewJsonFormatter()

func ParseAndFormat(raw []byte) (string, error) {
	var in Input
	if err := decodeJson(raw, &in); err != nil {
		return "", err
	}

	var m map[string]any
	if err := decodeJson([]byte(in.Message), &m); err != nil {
		return in.Format(), nil
	}

	timestamp := extractTimestamp(m)
	level := extractLevel(m)
	message := extractMessage(m)

	for _, key := range omitKeys {
		delete(m, key)
	}

	out := &Output{
		Timestamp:     timestamp,
		Level:         level,
		PodName:       in.PodName,
		ContainerName: in.ContainerName,
		Message:       message,
		Rests:         m,
	}

	return out.Format(), nil
}

type Input struct {
	NodeName      string `json:"nodeName"`
	Namespace     string `json:"namespace"`
	PodName       string `json:"podName"`
	ContainerName string `json:"containerName"`
	Message       string `json:"message"`
}

func (i *Input) Format() string {
	podColor, containerColor := determineColor(i.PodName)
	return fmt.Sprintf("%s %s %s", podColor.Sprint(i.PodName), containerColor.Sprint(i.ContainerName), i.Message)
}

func decodeJson(in []byte, data any) error {
	decoder := json.NewDecoder(bytes.NewReader(in))
	decoder.UseNumber()
	if err := decoder.Decode(data); err != nil {
		return err
	}
	return nil
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
	default:
		return fmt.Sprintf("%v", t)
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

type Output struct {
	Timestamp     string
	Level         string
	PodName       string
	ContainerName string
	Message       string
	Rests         map[string]any
}

func (o *Output) Format() string {
	var b bytes.Buffer
	podColor, containerColor := determineColor(o.PodName)

	if o.Timestamp != "" {
		b.WriteString(o.Timestamp)
		b.WriteString(" ")
	}

	if o.Level != "" {
		b.WriteString(levelColor(o.Level))
		b.WriteString(" ")
	}

	b.WriteString(podColor.Sprint(o.PodName))
	b.WriteString(" ")
	b.WriteString(containerColor.Sprint(o.ContainerName))
	b.WriteString(" ")
	b.WriteString(o.Message)

	if len(o.Rests) > 0 {
		b.WriteString(" ")
		b.Write(jsonFormatter.Format(o.Rests))
	}

	return b.String()
}
