# Docker Build and Test Guide

This guide contains reusable commands for building and testing GitStream Docker images.

## Build All Images

```sh
# Build all service images
docker build -f cmd/ingest/Dockerfile -t gitstream-ingest .
docker build -f cmd/processor/Dockerfile -t gitstream-processor .
docker build -f cmd/api/Dockerfile -t gitstream-api .
```

## Verify Image Sizes

```sh
# Check built image sizes
docker images | grep gitstream
```

Expected sizes:
- gitstream-ingest: ~47MB
- gitstream-processor: ~62MB 
- gitstream-api: ~58MB

## Test Image Health Checks

```sh
# Test ingest health check
docker run --rm -p 8080:8080 gitstream-ingest &
sleep 5
curl -i http://localhost:8080/health
docker stop $(docker ps -q --filter ancestor=gitstream-ingest)

# Test processor metrics endpoint
docker run --rm -p 8091:8091 gitstream-processor &
sleep 5
curl -i http://localhost:8091/metrics
docker stop $(docker ps -q --filter ancestor=gitstream-processor)

# Test API health check
docker run --rm -p 8090:8090 gitstream-api &
sleep 5
curl -i http://localhost:8090/health
docker stop $(docker ps -q --filter ancestor=gitstream-api)
```

## Clean Up

```sh
# Remove built images
docker rmi gitstream-ingest gitstream-processor gitstream-api

# Clean up build cache
docker builder prune
```

## Production Deployment Notes

For Fly.io deployment:
- Each Dockerfile expects environment variables for configuration
- Images run as non-root user (gitstream, UID 1001)
- Health checks are enabled on service endpoints
- Images include ca-certificates for HTTPS connections
- Static linking ensures binaries work without libc dependencies