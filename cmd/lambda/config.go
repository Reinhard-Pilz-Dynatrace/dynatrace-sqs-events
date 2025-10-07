package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	dtURL                 = mustGetEnv("DT_URL") // e.g. https://abc123.live.dynatrace.com
	dtToken               = mustGetEnv("DT_TOKEN")
	dtEntitySelector      = getEnv("DT_ENTITY_SELECTOR", "")
	dtEventType           = getEnv("DT_EVENT_TYPE", "AVAILABILITY_EVENT")
	dtTimeoutMS           = mustParseInt(getEnv("DT_TIMEOUT_MS", "0"))
	dtTitleMax            = int(mustParseInt(getEnv("DT_TITLE_MAX", "500")))
	dtOriginalQueueProp   = getEnv("DT_ORIGINAL_QUEUE_PROP", "originalQueueArn")
	dlqPrefix             = getEnv("DT_DLQ_PREFIX", "[DQL]")
	dynatraceEventsIngest = strings.TrimRight(dtURL, "/") + "/api/v2/events/ingest"
)

func mustGetEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Errorf("missing required env var: %s", k))
	}
	return v
}
func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func mustParseInt(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Errorf("invalid int value: %s", s))
	}
	return n
}
