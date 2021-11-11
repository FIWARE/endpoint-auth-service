# Outgoing-Request Auth-Sidecar

In various use-cases, there is a need to apply authn/z to outgoing requests for components that do not handle this them-self(f.e. notifications in
[NGSI-LD brokers](https://github.com/FIWARE/context.Orion-LD)). This sidecar provides that by adding an [envoy-proxy](https://www.envoyproxy.io) 
as [sidecar](https://www.oreilly.com/library/view/designing-distributed-systems/9781491983638/ch02.html) to the component that gets forwarded all 
outgoing requests via ip-tables(see [iptables-init](./iptables-init)). Another sidecar([auth-provider](./auth-provider)) provides target specific auth-tokens
that envoy adds to the requests. To configure the whole setup, the [configuration-api](./subscriber-config-api) can be used.

## Overview

![Proxy-Architecture](./arch-overview.svg)

## ADRs

- [Add authentication via a sidecar-proxy](./doc/adr/sidecar-based-auth.md)
- [Use envoy as proxy](./doc/adr/choose-proxy.md)
- [Use mustache templating for envoy config](./doc/adr/mustache-templating.md)

## APIs

- [Configuration-API](./subscriber-config-api/api/api.yaml)