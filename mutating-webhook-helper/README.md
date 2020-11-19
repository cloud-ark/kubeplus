# KubePlus Mutating webhook helper

## Description

Continuously scans a cluster to check for new CRDs and adds them to
MutatingWebhookConfiguration object so that KubePlus Mutating webhook can
intercept requests for the new Custom Resources.

## Development steps

1. Setup:
   - Use go 1.13
    - source setgopath.sh

2. Build:
   - Update versions.txt
   - Update Docker registry coordinates in build-artifact.sh

3. Deploy:
   - Update the version in deployment.yaml
   - kubectl apply -f deployment.yaml


