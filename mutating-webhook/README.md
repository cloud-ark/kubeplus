# k8s-mutating-webhook

## Status
In progress

## Description
Works kind of like helm, except at runtime it modifies and replaces Fn::ImportValue stuff from annotations of other Custom resource objects.
This is similar to AWS CloudFormation, and is aiming to give operators better interopability.


## Steps
1. Generate certs
    `make gen-certs`
2. Deploy Mutating Webhook Using Docker
    `dep ensure`
    Set environment variables
    ${PROJECT_ID} -> gcr project id
    ${IMAGE_NAME} -> crd-hook
    `make docker`
    `make deploy`
3. Install Operators
    `make install-operators`
4. Deploy application
    `make cluster`
    `make moodle`
5. Clean up
delete webhook: `make delete`
delete instances: `make delall`
