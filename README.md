# Dynatrace SQS Event Forwarder

This Lambda solution forwards messages from **Amazon SQS** queues to the **Dynatrace Events API** — converting queue messages into Dynatrace events and enriching them with SQS metadata.

---

## 🚀 Deployment

Deployment is handled entirely through **Terraform** using the configuration provided in this repository.

### Prerequisites

Before applying Terraform:
- Have a working **AWS account** and credentials configured.
- Have a **Dynatrace environment** and an **API token** with the `events.ingest` permission.
- Have the [Terraform CLI installed](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli#install-terraform) on your workstation
- Ensure you can access the repo’s **Terraform configuration** (contains the Lambda deployment logic).

### Variables to set

| Variable | Description | Example |
|-----------|--------------|----------|
| `aws_region` | AWS region to deploy the Lambda | `us-east-1` |
| `function_name` | Name of the Lambda function | `sqs-to-dynatrace` |
| `queue_arn` | ARN of the SQS queue to subscribe | `arn:aws:sqs:us-east-1:123456789012:my-queue` |
| `dt_url` | Dynatrace environment URL | `https://abc123.live.dynatrace.com` |
| `dt_token` | Dynatrace API token (events.ingest) | `dt0c01.abc…` |
| `github_token` | *(Optional)* A GitHub Token to avoid rate limitations | `github_pat_11BK...` |

## Set variables for Terraform
* Copy `infra/terraform/terraform.tfvars.example` from this repo to `infra/terraform/terraform.tfvars`.
* Open `infra/terraform/terraform.tfvars` and fill in your environment values`

## Create your Lambda Function using Terraform

Within the folder `infra/terraform/` then run:
```bash
terraform init
terraform apply
```

💡Confirm the plan with `yes` when prompted.  

You can safely ignore any warnings regarding `UTF-8` at this stage

Terraform automatically:
- Downloads the **latest release artifact** (`function-vX.Y.Z.zip`) from GitHub.
- Creates or updates **and** configures the Lambda function via Environment Variables.
- Subscribes it to the specified SQS queue.

---

## ⚙️ Lambda Behavior

### 1. Raw messages (text)
If an SQS message body is **plain text**, it is sent to Dynatrace as a new event:

| Field | Value |
|--------|--------|
| **Title** | Message body (truncated if necessary) |
| **Event type** | `AVAILABILITY_EVENT` (default) |
| **Properties** | Includes queue ARN, message ID, and receive count |

### 2. Event JSON messages
If the message body is already a **valid Dynatrace event JSON**, the Lambda forwards it directly — only adding metadata.

The function:
- **Preserves** your title, event type, and properties.
- **Adds** queue information (`queueArn`, `messageId`, etc.).
- **Adds** timing fields (`startTime`, `endTime`) if missing.

### 3. DLQ messages
If the Lambda is subscribed to a **Dead Letter Queue (DLQ)**:
- It automatically detects this and prefixes event titles with **`[DLQ]`**.
- The forwarded event includes `properties.originalQueue` when available.

If you include a property or SQS message attribute (default key: `originalQueueArn`) that identifies the original queue, the Lambda will also include it in the event metadata.

💡Your Lambda function allows for reconfiguring the key `originalQueueArn`.

---

## 🧠 Message Processing Summary

| Scenario | Behavior | Dynatrace Event Title |
|-----------|-----------|-----------------------|
| Plain text body | Creates a new Dynatrace event | Message content |
| JSON body (not event format) | Creates new Dynatrace event | JSON as string |
| Dynatrace event JSON body | Forwards as-is with enrichment | Existing `title` |
| DLQ message | Adds `[DLQ]` prefix to title | `[DLQ] …` |

All events are enriched with:
- Queue ARN
- Message ID
- Approximate receive count
- Timestamps
- (Optional) Original queue ARN (if provided)

---

## 🔐 Permissions

Terraform attaches an IAM role granting:
- Lambda execution (`logs:*`)
- SQS read and delete (`ReceiveMessage`, `DeleteMessage`, etc.)
- DLQ inspection (`ListDeadLetterSourceQueues`)

---

## 🪵 Logging

Lambda writes diagnostic logs to **Amazon CloudWatch**.

When the `DEBUG` environment variable is set to `true`, each outgoing Dynatrace payload is printed before sending.

---

## 💡 Example Use Cases

| Use Case | Description |
|-----------|--------------|
| **Forward operational alerts** | Forward application or infrastructure alerts as Dynatrace events for correlation. |
| **Monitor DLQs** | Trigger Dynatrace events whenever a message lands in a DLQ. |
| **Integrate external systems** | Push system events to SQS from other services, and automatically surface them in Dynatrace. |

---

## 🧰 Configuration Reference

| Environment Variable | Description | Default |
|-----------------------|--------------|----------|
| `DT_URL` | Dynatrace environment URL | — |
| `DT_TOKEN` | Dynatrace API token | — |
| `DT_ENTITY_SELECTOR` | Default entity selector | |
| `DT_EVENT_TYPE` | Default event type for raw messages | `AVAILABILITY_EVENT` |
| `DT_TIMEOUT_MS` | Auto-close timeout in ms | `0` |
| `DT_TITLE_MAX` | Max title length | `500` |
| `DT_ORIGINAL_QUEUE_PROP` | Key for original queue (in body or attributes) | `originalQueueArn` |
| `DT_DLQ_PREFIX` | Prefix for DLQ event titles | `[DLQ]` |
| `DEBUG` | Enables verbose logs | unset |

---

## ✅ Summary

| Feature | Description |
|----------|--------------|
| **Raw messages → Events** | Converts SQS text into Dynatrace events |
| **Event JSON passthrough** | Forwards existing event JSON unchanged |
| **DLQ detection** | Prefixes title and enriches metadata |
| **Terraform deployment** | Uses latest GitHub release automatically |
| **No build required** | Prebuilt binary is downloaded for you |

---

> Once deployed, any message arriving in your queue will be automatically transformed and sent to Dynatrace as an event — enriched with full SQS context and ready for analysis in your dashboards and alerting workflows.
