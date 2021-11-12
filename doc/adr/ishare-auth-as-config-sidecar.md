# Deploy the iShare-auth-provider as a sidecar to the configuration-service

## Status

- proposed

## Context

The iShare-auth-provider requires confidential credentials-information(e.g. the signing-key) to be provided by the configuration-service. 

## Decision

The auth-provider will be deployed as a sidecar to the configuration-service and read the credentials from the shared file-system.

## Rational

- the key never has to leave the endpoint-auth-service, therefore reduces potential attacking points