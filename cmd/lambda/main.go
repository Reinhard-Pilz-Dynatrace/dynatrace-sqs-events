package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	httpClient = &http.Client{Timeout: 10 * time.Second}
	sqsClient  *sqs.Client

	// cache per invocation for DLQ check
	dlqCache = map[string]bool{}
)

func baseEnrichment(rec events.SQSMessage) map[string]any {
	a := rec.Attributes
	return map[string]any{
		"sqsMessageId":            rec.MessageId,
		"queueArn":                rec.EventSourceARN,
		"approximateReceiveCount": a["ApproximateReceiveCount"],
		"sentTimestamp":           a["SentTimestamp"],
		"firstReceiveTimestamp":   a["ApproximateFirstReceiveTimestamp"],
	}
}

func messageAttr(rec events.SQSMessage, key string) (string, bool) {
	if ma, ok := rec.MessageAttributes[key]; ok {
		if ma.StringValue != nil {
			return *ma.StringValue, true
		}
	}
	return "", false
}

// arn: arn:aws:sqs:REGION:ACCOUNT:NAME
func parseQueueFromArn(arn string) (region, account, name string) {
	parts := strings.Split(arn, ":")
	if len(parts) >= 6 {
		return parts[3], parts[4], parts[5]
	}
	return "", "", ""
}

func getQueueURL(ctx context.Context, arn string) (string, error) {
	_, account, name := parseQueueFromArn(arn)
	out, err := sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName:              aws.String(name),
		QueueOwnerAWSAccountId: aws.String(account),
	})
	if err != nil {
		return "", err
	}
	return aws.ToString(out.QueueUrl), nil
}

// ---------- Dynatrace ingest ----------

func postDynatrace(ctx context.Context, ev dynatraceEvent) error {
	b, _ := json.Marshal(ev)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, dynatraceEventsIngest, strings.NewReader(string(b)))
	req.Header.Set("Authorization", "Api-Token "+dtToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dynatrace http error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dynatrace ingest failed: status=%d", resp.StatusCode)
	}
	return nil
}

// ---------- per-record ----------

func processRecord(ctx context.Context, rec events.SQSMessage, isDLQQueue bool) error {
	now := time.Now().UnixMilli()
	raw := rec.Body

	// Try to parse once
	var parsed map[string]interface{}
	_ = json.Unmarshal([]byte(raw), &parsed)

	origHint, hasOrig := extractOriginalQueueHint(rec, parsed)
	dlqDetected := isDLQQueue || hasOrig

	if looksLikeDynatraceEvent(parsed) {
		// Enrich existing event (keep title)
		props := ensureProps(parsed)
		for k, v := range baseEnrichment(rec) {
			props[k] = v
		}
		if hasOrig {
			props["originalQueue"] = origHint
		}
		// defaults if missing
		if _, ok := parsed["startTime"]; !ok {
			parsed["startTime"] = now
		}
		if _, ok := parsed["endTime"]; !ok {
			parsed["endTime"] = now
		}
		if _, ok := parsed["timeout"]; !ok {
			parsed["timeout"] = dtTimeoutMS
		}
		if _, ok := parsed["entitySelector"]; !ok {
			parsed["entitySelector"] = dtEntitySelector
		}
		// title prefix if DLQ
		if t, ok := parsed["title"].(string); ok {
			parsed["title"] = prefixTitleIfDLQ(t, dlqDetected)
		}

		// Convert back to struct for send
		out := dynatraceEvent{}
		b, _ := json.Marshal(parsed)
		if err := json.Unmarshal(b, &out); err != nil {
			return fmt.Errorf("normalize event failed: %w", err)
		}
		return postDynatrace(ctx, out)
	}

	// Build new event from scratch
	title := truncate(compactFromAny(raw), dtTitleMax)
	ev := dynatraceEvent{
		StartTime:      now,
		EndTime:        now,
		Timeout:        dtTimeoutMS,
		EntitySelector: dtEntitySelector,
		EventType:      dtEventType,
		Title:          prefixTitleIfDLQ(title, dlqDetected),
		Properties:     baseEnrichment(rec),
	}
	if hasOrig {
		ev.Properties["originalQueue"] = origHint
	}

	if os.Getenv("DEBUG") == "true" {
		b, _ := json.MarshalIndent(ev, "", "  ")
		fmt.Printf("📦 Sending to Dynatrace:\n%s\n", string(b))
	}

	return postDynatrace(ctx, ev)
}

// ---------- handler ----------

func handler(ctx context.Context, e events.SQSEvent) (events.SQSEventResponse, error) {
	// init SDK client once per cold start
	if sqsClient == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return events.SQSEventResponse{}, fmt.Errorf("aws config: %w", err)
		}
		sqsClient = sqs.NewFromConfig(cfg)
	}

	resp := events.SQSEventResponse{BatchItemFailures: []events.SQSBatchItemFailure{}}

	var queueArn string
	if len(e.Records) > 0 {
		queueArn = e.Records[0].EventSourceARN
	}
	isDLQ := isQueueADLQ(ctx, queueArn)

	for _, rec := range e.Records {
		if err := processRecord(ctx, rec, isDLQ); err != nil {
			// mark only this record as failed => partial batch failure
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{
				ItemIdentifier: rec.MessageId,
			})
		}
	}
	return resp, nil
}

func main() { lambda.Start(handler) }
