package main

import "encoding/json"

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func compactFromAny(raw string) string {
	// If JSON, compact it; otherwise return as-is.
	var any interface{}
	if err := json.Unmarshal([]byte(raw), &any); err == nil {
		switch v := any.(type) {
		case string:
			return v
		default:
			b, _ := json.Marshal(v) // compact by default
			return string(b)
		}
	}
	return raw
}

func ensureProps(m map[string]any) map[string]any {
	props, ok := m["properties"].(map[string]any)
	if !ok || props == nil {
		props = map[string]any{}
		m["properties"] = props
	}
	return props
}
