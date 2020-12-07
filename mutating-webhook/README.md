# KubePlus Mutating webhook

## Description

KubePlus mutating webhook performs following functions:
- track user/account who is creating Kubernetes resources (the identity of the creator is added as an annotation on the resource spec)
- create instances of custom services registered as Custom Resources in a cluster
- resolve binding functions (ImportValue, AddLabel, AddAnnotations)


## Development steps

1. Setup:
   - Use go 1.13
    - source setgopath.sh

2. Build:
   - Update Docker registry coordinates in Makefile (docker and docker1 rules)
   - make docker

3. Deploy:
   - make deploy

4. Delete:
  - make delete