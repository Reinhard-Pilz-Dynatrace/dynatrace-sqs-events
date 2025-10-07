package main

type dynatraceEvent struct {
	EndTime        int64          `json:"endTime"`
	EntitySelector string         `json:"entitySelector,omitempty"`
	EventType      string         `json:"eventType"`
	Properties     map[string]any `json:"properties"`
	StartTime      int64          `json:"startTime"`
	Timeout        int64          `json:"timeout"`
	Title          string         `json:"title"`
}

func looksLikeDynatraceEvent(m map[string]interface{}) bool {
	if m == nil {
		return false
	}
	_, hasType := m["eventType"]
	_, hasTitle := m["title"]
	if hasType && hasTitle {
		return true
	}
	keys := []string{"endTime", "entitySelector", "eventType", "properties", "startTime", "timeout", "title"}
	score := 0
	for _, k := range keys {
		if _, ok := m[k]; ok {
			score++
		}
	}
	return score >= 3
}
