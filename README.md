# logstream-tail

Real-time log aggregator that fans in from multiple cloud sources into a single terminal stream.

## Overview

`logstream-tail` connects to CloudWatch Logs and GCP Cloud Logging simultaneously, multiplexing output into a unified, color-coded stream in your terminal — like `tail -f` for the cloud.

## Installation

```bash
go install github.com/yourorg/logstream-tail@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/logstream-tail.git
cd logstream-tail
make build
```

## Usage

```bash
# Tail a CloudWatch log group and a GCP log bucket simultaneously
logstream-tail \
  --cloudwatch /aws/lambda/my-function \
  --gcp projects/my-project/logs/my-service

# Filter output with a keyword
logstream-tail --cloudwatch /aws/lambda/my-function --filter "ERROR"

# Set a custom poll interval (default: 5s)
logstream-tail --cloudwatch /aws/lambda/my-function --interval 2s
```

### Configuration

Credentials are picked up from the standard provider chains:

- **AWS**: `AWS_PROFILE`, `~/.aws/credentials`, or instance role
- **GCP**: `GOOGLE_APPLICATION_CREDENTIALS` or application default credentials

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--cloudwatch` | — | CloudWatch log group ARN or name |
| `--gcp` | — | GCP log resource path |
| `--filter` | — | Keyword filter applied to all sources |
| `--interval` | `5s` | Polling interval per source |
| `--no-color` | `false` | Disable colorized output |

## Requirements

- Go 1.21+
- AWS credentials (for CloudWatch)
- GCP credentials (for GCP Logging)

## License

MIT © 2024 yourorg