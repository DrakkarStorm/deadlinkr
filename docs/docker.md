# 🐳 Docker Guide for Deadlinkr

Simple guide for using Deadlinkr with Docker.

## Quick Start

### 1. Pull and Run

```bash
# Pull the latest image
docker pull ghcr.io/drakkarstorm/deadlinkr:latest

# Run a quick scan
docker run --rm ghcr.io/drakkarstorm/deadlinkr:latest scan https://example.com
```

### 2. Basic Usage Examples

```bash
# Scan with auto-detected format (from extension)
docker run --rm -v $(pwd)/reports:/app/reports \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com \
  -d 2 -c 50 -o /app/reports/scan.json

# Scan with shortcuts and authentication
docker run --rm ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://private-site.com \
  --auth-basic "username:password" \
  -d 1 -o report.csv

# Show all links including working ones
docker run --rm ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com \
  --show-all -o full-report.html
```

## Available Images

### Official Images

| Registry | Image | Description |
|----------|-------|-------------|
| GitHub Container Registry | `ghcr.io/drakkarstorm/deadlinkr` | Official images |
| Docker Hub | `deadlinkr/deadlinkr` | Mirror on Docker Hub |

### Tags

| Tag | Description | Architectures |
|-----|-------------|---------------|
| `latest` | Latest stable release | `linux/amd64`, `linux/arm64` |
| `v1.x.x` | Specific version | `linux/amd64`, `linux/arm64` |
| `develop` | Development branch | `linux/amd64`, `linux/arm64` |
| `weekly-YYYYMMDD` | Weekly builds | `linux/amd64`, `linux/arm64` |

## Configuration

### Environment Variables

The Docker image supports all Deadlinkr configuration via environment variables:

```bash
docker run --rm \
  -e LOG_LEVEL=debug \
  -e TIMEOUT=30 \
  -e CONCURRENCY=100 \
  -e RATE_LIMIT=5.0 \
  -e CACHE_ENABLED=true \
  -e DEADLINKR_AUTH_USER=username \
  -e DEADLINKR_AUTH_PASS=password \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://private-site.com
```

### Volume Mounts

#### Configuration Files

```bash
# Mount configuration directory
docker run --rm \
  -v $(pwd)/config:/app/config:ro \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com
```

#### Output Reports

```bash
# Save reports to host (format auto-detected)
docker run --rm \
  -v $(pwd)/reports:/app/reports:rw \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com \
  -o /app/reports/scan-report.json
```

#### Log Files

```bash
# Persist log files
docker run --rm \
  -v $(pwd)/logs:/app/logs:rw \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com \
  --log-level debug
```

## Scheduled Scanning

For automated scheduled scans, you can use cron or systemd timers directly with Docker:

```bash
# Add to crontab (crontab -e)
0 2 * * * docker run --rm -v /var/log/deadlinkr:/app/reports ghcr.io/drakkarstorm/deadlinkr:latest scan https://example.com -o /app/reports/scan-$(date +\%Y\%m\%d).json
```

## Advanced Usage

### CI/CD Integration

#### GitHub Actions

```yaml
# .github/workflows/link-check.yml
name: Link Check
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  link-check:
    runs-on: ubuntu-latest
    steps:
      - name: Run Deadlinkr
        run: |
          docker run --rm \
            -v ${{ github.workspace }}/reports:/app/reports \
            ghcr.io/drakkarstorm/deadlinkr:latest \
            scan https://example.com \
            -d 3 \
            -o /app/reports/link-check.json
      
      - name: Upload Report
        uses: actions/upload-artifact@v4
        with:
          name: link-check-report
          path: reports/link-check.json
```

#### GitLab CI

```yaml
# .gitlab-ci.yml
link-check:
  stage: test
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker run --rm 
        -v $PWD/reports:/app/reports 
        ghcr.io/drakkarstorm/deadlinkr:latest 
        scan https://example.com 
        -o /app/reports/report.json
  artifacts:
    paths:
      - reports/
    expire_in: 1 week
```

### Kubernetes Deployment

#### Basic Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deadlinkr
  labels:
    app: deadlinkr
spec:
  replicas: 1
  selector:
    matchLabels:
      app: deadlinkr
  template:
    metadata:
      labels:
        app: deadlinkr
    spec:
      containers:
      - name: deadlinkr
        image: ghcr.io/drakkarstorm/deadlinkr:latest
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: CONCURRENCY
          value: "50"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        volumeMounts:
        - name: reports
          mountPath: /app/reports
        - name: config
          mountPath: /app/config
          readOnly: true
      volumes:
      - name: reports
        persistentVolumeClaim:
          claimName: deadlinkr-reports
      - name: config
        configMap:
          name: deadlinkr-config
```

#### CronJob for Scheduled Scans

```yaml
# k8s/cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: deadlinkr-scanner
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: deadlinkr
            image: ghcr.io/drakkarstorm/deadlinkr:latest
            args:
            - "scan"
            - "https://example.com"
            - "-d"
            - "2"
            - "-o"
            - "/app/reports/scan.json"
            env:
            - name: LOG_LEVEL
              value: "info"
            - name: CONCURRENCY
              value: "30"
            volumeMounts:
            - name: reports
              mountPath: /app/reports
          volumes:
          - name: reports
            persistentVolumeClaim:
              claimName: deadlinkr-reports
          restartPolicy: OnFailure
```

## Security

### Non-Root User

The Docker image runs as a non-root user (`nonroot:nonroot`) for enhanced security:

```bash
# Check user in container
docker run --rm ghcr.io/drakkarstorm/deadlinkr:latest whoami
# Output: nonroot
```

### Read-Only Root Filesystem

For maximum security, you can run with a read-only root filesystem:

```bash
docker run --rm --read-only \
  --tmpfs /tmp:noexec,nosuid,size=50m \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com
```

### Security Scanning

The images are automatically scanned for vulnerabilities:

```bash
# Scan image locally with Trivy
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy:latest image ghcr.io/drakkarstorm/deadlinkr:latest
```

## Building Custom Images

### Local Build

```bash
# Build locally
./scripts/docker-build.sh --tag custom

# Build and push
./scripts/docker-build.sh --tag v1.0.0 --push

# Multi-architecture build
./scripts/docker-build.sh --tag latest --multi-arch --push
```

### Custom Dockerfile

```dockerfile
FROM ghcr.io/drakkarstorm/deadlinkr:latest

# Add custom configuration
COPY my-config.yaml /app/config/

# Override default settings
ENV LOG_LEVEL=debug
ENV CONCURRENCY=100

# Custom entrypoint
COPY custom-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/custom-entrypoint.sh"]
```

## Troubleshooting

### Common Issues

#### Permission Denied

```bash
# Fix volume permissions
sudo chown -R 65532:65532 ./reports ./logs

# Or use different user
docker run --rm --user $(id -u):$(id -g) \
  -v $(pwd)/reports:/app/reports \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com
```

#### Out of Memory

```bash
# Limit memory and adjust concurrency
docker run --rm --memory="512m" \
  -e CONCURRENCY=10 \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com
```

#### Network Issues

```bash
# Use host network for debugging
docker run --rm --network=host \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com
```

### Debug Mode

```bash
# Run with debug logging
docker run --rm \
  -e LOG_LEVEL=debug \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com
```

### Health Checks

```bash
# Check container health
docker run --name deadlinkr-test \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  tail -f /dev/null

# In another terminal
docker exec deadlinkr-test /usr/local/bin/deadlinkr --help

# Cleanup
docker rm -f deadlinkr-test
```

## Performance Optimization

### Resource Limits

```yaml
# docker-compose.yml
services:
  deadlinkr:
    image: ghcr.io/drakkarstorm/deadlinkr:latest
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '1.0'
        reservations:
          memory: 256M
          cpus: '0.25'
```

### Caching

```bash
# Use Docker BuildKit cache
export DOCKER_BUILDKIT=1

# Build with cache mount
docker build --cache-from ghcr.io/drakkarstorm/deadlinkr:latest .
```

## Best Practices

1. **Use specific tags** instead of `latest` in production
2. **Limit resource usage** with Docker resource constraints
3. **Use read-only filesystems** when possible
4. **Mount volumes** for persistent data (reports, logs, config)
5. **Set appropriate environment variables** for your use case
6. **Use multi-stage builds** for custom images
7. **Scan images** for vulnerabilities regularly
8. **Use non-root users** (already default in official images)
9. **Keep images updated** to get latest security patches
10. **Monitor resource usage** and adjust settings accordingly