# Implement Auth-Provider not as part of the configuration-service 

## Status

- proposed

## Context

Every auth-provider implementation requires specific logic and information for its authentication type. For example in case 
of iShare, the auth-provider needs access to a certificate-chain and a signing-key that are used to create a specifc JWT.

## Decision

The auth-provider(s) will be implemented as separate components.

## Rational

- flexible in terms of implementation language -> easier to contribute new providers
- plugin-ready-solution -> for a given use-case, only the providers that a required need to be deployed
- easier to test the functionality of single providers, than to always test the whole system.
- ability to use non-java libraries, if there already is support for an auth-type