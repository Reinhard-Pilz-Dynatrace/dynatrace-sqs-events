# IAM role
data "aws_iam_policy_document" "assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "${var.function_name}-role"
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

# CloudWatch Logs
resource "aws_iam_role_policy_attachment" "logs" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# SQS least-privilege (scope to queue_arn)
data "aws_iam_policy_document" "sqs" {
  statement {
    sid    = "SqsConsume"
    effect = "Allow"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:ChangeMessageVisibility",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ListDeadLetterSourceQueues"
    ]
    resources = [var.queue_arn]
  }
}

resource "aws_iam_policy" "sqs" {
  name   = "${var.function_name}-sqs"
  policy = data.aws_iam_policy_document.sqs.json
}

resource "aws_iam_role_policy_attachment" "sqs" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.sqs.arn
}

# Event source mapping (SQS trigger)
resource "aws_lambda_event_source_mapping" "sqs" {
  event_source_arn                   = var.queue_arn
  function_name                      = aws_lambda_function.this.arn
  batch_size                         = 10
  maximum_batching_window_in_seconds = 0
  enabled                            = true
}


# Get the latest release
data "github_release" "latest" {
  owner       = var.github_owner
  repository  = var.github_repo
  retrieve_by = "latest"
}

# Pick the asset that matches your name pattern
locals {
  selected_asset = one([
    for a in data.github_release.latest.assets : a
    if can(regex("^function-.*\\.zip$", a.name))
  ])
  release_zip_url = local.selected_asset.browser_download_url

  # Compute the effective path and hash
  effective_zip_path = local_file.release_zip.filename
  release_zip_hash = base64sha256(data.http.release_zip.response_body_base64)
}

# Download the ZIP (use the base64 body for binaries)
data "http" "release_zip" {
  url   = local.release_zip_url

  # Tell GitHub we're after a binary
  request_headers = {
    Accept = "application/octet-stream"
  }
}

# Write the ZIP to a known path inside this module; no parent dirs required
resource "local_file" "release_zip" {
  filename       = "${path.module}/function-from-release.zip"
  content_base64 = sensitive(data.http.release_zip.response_body_base64)  
}

resource "aws_lambda_function" "this" {
  function_name = var.function_name
  role          = aws_iam_role.lambda.arn

  filename         = local.effective_zip_path
  source_code_hash = local.release_zip_hash

  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = [var.architecture]

  environment {
    variables = {
      DT_URL                 = var.dt_url
      DT_TOKEN               = var.dt_token
      DT_ENTITY_SELECTOR     = var.dt_entity_selector
      DT_EVENT_TYPE          = var.dt_event_type
      DT_TIMEOUT_MS          = tostring(var.dt_timeout_ms)
      DT_TITLE_MAX           = tostring(var.dt_title_max)
      DT_ORIGINAL_QUEUE_PROP = var.dt_original_queue_prop
      DT_DLQ_PREFIX          = var.dt_dlq_prefix
    }
  }
}
