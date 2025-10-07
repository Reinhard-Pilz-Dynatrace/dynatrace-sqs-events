package main

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func extractOriginalQueueHint(rec events.SQSMessage, parsed map[string]any) (string, bool) {
	// 1) body property
	if parsed != nil {
		if v, ok := parsed[dtOriginalQueueProp]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s, true
			}
		}
	}
	// 2) message attribute
	if s, ok := messageAttr(rec, dtOriginalQueueProp); ok && s != "" {
		return s, true
	}
	return "", false
}

func prefixTitleIfDLQ(title string, isDLQ bool) string {
	if !isDLQ {
		return title
	}
	if strings.HasPrefix(title, dlqPrefix) {
		return title
	}
	return strings.TrimSpace(dlqPrefix + " " + title)
}

func isQueueADLQ(ctx context.Context, arn string) bool {
	if arn == "" {
		return false
	}
	if v, ok := dlqCache[arn]; ok {
		return v
	}
	url, err := getQueueURL(ctx, arn)
	if err != nil {
		// if we can't tell, assume not a DLQ
		dlqCache[arn] = false
		return false
	}
	out, err := sqsClient.ListDeadLetterSourceQueues(ctx, &sqs.ListDeadLetterSourceQueuesInput{
		QueueUrl:   aws.String(url),
		MaxResults: aws.Int32(10),
	})
	if err != nil {
		dlqCache[arn] = false
		return false
	}
	is := len(out.QueueUrls) > 0
	dlqCache[arn] = is
	return is
}
