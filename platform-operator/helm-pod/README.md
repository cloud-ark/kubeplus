# Helmer

## Description

Helmer provides a REST API to deploy Helm charts and retrieve Prometheus metrics for the deployed chart (helm release).

## Development steps

1. Setup:
   Use Golang version 1.13 (1.14 is also okay)

2. Modify:
   - Making changes to Dockerfile
   - Make changes to main.go

3. Build:
   - Update Docker registry coordinates in build-artifact.sh
   - Update versions.txt before creating new versioned artifact.

  ./build-artifact.sh <latest | versioned>
