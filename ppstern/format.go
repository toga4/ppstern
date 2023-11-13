package ppstern

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

	timestamp := formatTimestamp(removeAny[any](m, "time", "ts", "timestamp"))
	level := removeAny[string](m, "level", "severity")
	message := removeAny[string](m, "msg", "message")
	trace := removeAny[string](m, "trace", "logging.googleapis.com/trace")
	span := removeAny[string](m, "span", "logging.googleapis.com/spanId")

	omitKeys := []string{
		"logging.googleapis.com/trace",
		"logging.googleapis.com/spanId",
		"logging.googleapis.com/trace_sampled",
		"time",
		"ts",
		"caller",
		"stacktrace",
	}
	for _, key := range omitKeys {
		delete(m, key)
	}

	out := &Output{
		Timestamp:     timestamp,
		Level:         level,
		PodName:       in.PodName,
		ContainerName: in.ContainerName,
		Message:       message,
		Trace:         trace,
		Span:          span,
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
	podColor, containerColor := determineColorForPod(i.PodName)
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

func removeAny[T any](m map[string]any, keys ...string) (v T) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if v, ok := v.(T); ok {
				delete(m, k)
				return v
			}
		}
	}
	return
}

var supportedLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	time.DateTime,
	time.Layout,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
}

const (
	timestampLayout = "2006-01-02T15:04:05.000Z07:00"
)

func formatTimestamp(t any) string {
	var err error
	switch timestamp := t.(type) {
	case string:
		var t time.Time
		for _, l := range supportedLayouts {
			if t, err = time.Parse(l, timestamp); err == nil {
				return t.Format(timestampLayout)
			}
		}
		return timestamp
	case json.Number:
		if strings.Contains(timestamp.String(), ".") {
			var f float64
			if f, err = timestamp.Float64(); err == nil {
				return time.Unix(int64(f), int64((f-float64(int64(f)))*1e9)).Format(timestampLayout)
			}
		} else {
			var i int64
			if i, err = timestamp.Int64(); err == nil {
				return time.Unix(i, 0).Format(timestampLayout)
			}
		}
	default:
		return fmt.Sprintf("%v", t)
	}
	return err.Error()
}

type Output struct {
	Timestamp     string
	Level         string
	PodName       string
	ContainerName string
	Message       string
	Trace         string
	Span          string
	Rests         map[string]any
}

func (o *Output) Format() string {
	var b bytes.Buffer
	podColor, containerColor := determineColorForPod(o.PodName)

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

	if o.Trace != "" {
		traceColor := determineColor(o.Trace)
		b.WriteString(traceColor.Sprint(lastN(o.Trace, 6)))
		b.WriteString(" ")
	}
	if o.Span != "" {
		spanColor := determineColor(o.Span)
		b.WriteString(spanColor.Sprint(lastN(o.Span, 6)))
		b.WriteString(" ")
	}

	b.WriteString(o.Message)

	if len(o.Rests) > 0 {
		b.WriteString(" ")
		b.Write(jsonFormatter.Format(o.Rests))
	}

	return b.String()
}

func lastN(s string, n int) string {
	return s[max(0, len(s)-n):]
}
