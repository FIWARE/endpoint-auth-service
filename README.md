[![](https://nexus.lab.fiware.org/repository/raw/public/badges/chapters/api-management.svg)](https://www.fiware.org/developers/catalogue/)
[![License badge](https://img.shields.io/github/license/FIWARE/endpoint-auth-service.svg)](https://opensource.org/licenses/AGPL-3.0)
[![Container Repository on Quay](https://img.shields.io/badge/quay.io-FIWARE-green "Container Repository on Quay")](https://quay.io/repository/fiware/endpoint-configuration-service?tab=tags)
[![](https://img.shields.io/badge/tag-fiware-orange.svg?logo=stackoverflow)](http://stackoverflow.com/questions/tagged/fiware)
<br>
![Status](https://nexus.lab.fiware.org/static/badges/statuses/incubating.svg)
[![Coverage Status](https://coveralls.io/repos/github/FIWARE/endpoint-auth-service/badge.svg?branch=main)](https://coveralls.io/github/FIWARE/endpoint-auth-service?branch=main)
[![Unit-Test](https://github.com/fiware/endpoint-auth-service/actions/workflows/unit.yml/badge.svg)](https://github.com/fiware/endpoint-auth-service/actions/workflows/unit.yml)
[![Integration-test](https://messages.cucumber.io/api/report-collections/28ab3c23-79eb-4497-89f0-429a11c0eeff/badge)](https://reports.cucumber.io/report-collections/28ab3c23-79eb-4497-89f0-429a11c0eeff)

-------

# Endpoint-Auth-Service

In various use-cases, there is a need to apply authn/z to outgoing requests for components that do not handle this them-self(f.e. notifications in
[NGSI-LD brokers](https://github.com/FIWARE/context.Orion-LD)). This service provides that by adding an [envoy-proxy](https://www.envoyproxy.io) 
as [sidecar](https://www.oreilly.com/library/view/designing-distributed-systems/9781491983638/ch02.html) to the component that gets forwarded all 
outgoing requests via ip-tables(see [iptables-init](./src/iptables-init)). The sidecar-proxy does request auth-information at the [auth-provider](./src/auth-provider) 
and adds it to the requests accordingly. The endpoints to be handled and there auth-information can be configured through
[endpoint-configuration-service](./src/endpoint-configuration-service).

## Overview

![Proxy-Architecture](./doc/img/arch-overview.svg)

The architecture consists of 2 main components:
- the sidecar-proxy to intercept and manipulate outgoing requests
- the endpoint-auth-service to provide configuration and authentication information

This architecture allows to separate the actual authentication flows from the proxy itself, thus giving more flexibility in terms of technology
and reduces the complexity of the lua-code inside the proxy. All lua-code placed there needs to be non-blocking to not kill the performance of the 
proxy(and therefore the proxied requests itself). Most auth-flows require contacting external services(f.e. the IDP in the iShare use-case) or some 
file-io(f.e. reading the certs in iShare), which is hard to implement in a non-blocking fashion. With using the built-in http-call method and moving 
the complexity into the domain of the auth-provider, this can be avoided. Besided that, it allows to use request-caching for thos calls to the auth-provider
in order to prevent a new auth-flow for every request.

## Run

To run the service locally, see [docker-compose](docker-compose/README.md)

## Testing

Unit-testing is dependent on the concrete component and described in the individual folders. See f.e. [iShare-auth-provider](src/ishare-auth-provider/README.md#Testing).
Integration-Testing is described in the [integration-test suite](integration-test/README.md).

## Development

For general development information, check the [contribution-guidelines](CONTRIBUTING.md).
For information about the individual components, see their folders.

## Deployment 

While its possible to use this service basically inside any environment(see for example [docker-compose](./docker-compose)), its highly 
recommended for [kubernetes-environments](https://kubernetes.io/). Since networking can be manipulated on a per-pod base, its much easier 
to set up the required ip-tables and prevent other components from beeing influenced by that.

To make the deployment easy, a helm-chart is provided here:
- https://github.com/FIWARE/helm-charts/tree/main/charts/endpoint-auth-service

See [kubernetes](kubernetes) for a full example.

## Component specific documentation

* [envoy](envoy/README.md)
* [auth-provider](doc/AUTHPROVIDER.md)
* [endpoint-configuration-servcie](src/endpoint-configuration-service/README.md)
* [iShare-auth-provider](src/ishare-auth-provider/README.md)
* [init-iptables](src/iptables-init/IPTABLES.md)
* [envoy-configmap-updater](src/envoy-configmap-updater/README.md)
* [envoy-resource-updater](src/envoy-resource-updater/README.md)
* [mesh-extension-updater](src/mesh-extension-updater/README.md)

## ADRs

- [Add authentication via a sidecar-proxy](./doc/adr/sidecar-based-auth.md)
- [Use envoy as proxy](./doc/adr/choose-proxy.md)
- [Use mustache templating for envoy config](./doc/adr/mustache-templating.md)
- [Implement auth-providers as separate components](./doc/adr/authprovider-as-separate-component.md)
- [Auth-provider config on path level](./doc/adr/auth-provider-on-path-level.md)
- [Auth-mechanism as wasm-filter](./doc/adr/wasm-filter.md)

## APIs

- [Endpoint-Configuration-API](./api/endpoint-configuration-api.yaml)
- [Auth-Provider-API](./api/auth-provider-api.yaml)
- [iShare-Credentials-Management-API](./api/ishare-credentials-management-api.yaml)

## HTTPS endpoints

Since the proxy intercepts traffic transparently, it can only handle http-traffic. Adding auth-information requires the manipulation of request-headers which is not possible for
encrypted traffic. Therefor the handled component, should use http to send its requests(f.e. [orion-ld notifications](https://github.com/FIWARE/context.Orion-LD) should use a http-target address).
In order to not deteriorate the security of the system, an endpoint can be configured to apply tls to the connection. In this case the sidecar-proxy will forward the request via https, even if it did
come in as an http-request.

Example configuration:

POST 
```shell
curl -X POST 'endpoint-configuration-service/endpoint' \
  -H 'Content-Type: application/json' \
  -d '{
      "domain": "myNotificationEndpoint.org",
      "port": 80,
      "path": "/receive",
      "useHttps": true,
      "authType": "iShare",
      "authCredentials": {"..."}
  }'
```

If the handled service now sends the following request:

```shell
    curl -X POST 'http://myNotificationEndpoint.org/receive' \
      -H 'Content-Type: application/json' \
      -d '{"notification":"hello"}'
```

the proxy will rewrite the request to be like:

```shell
    curl -X POST 'https://myNotificationEndpoint.org/receive' \
      -H 'Content-Type: application/json' \
      -H 'authorization: myToken' \
      -d '{"notification":"hello"}'
```


## Why not use mTLS?

[mTLS](https://en.wikipedia.org/wiki/Mutual_authentication#mTLS) is a method for mutual authentication. The parties on each end of the connection can be verified through TLS certificates. In contrast to that,
the endpoint-authentication-service solution targets the authentication-handling of only one side of the connection. The authentication of the connection-target is optionally available 
through https. 
In the general context of the endpoint-auth-service, participating in mTLS can be seen as an additional auth-method. In order to support that, the [lua-script in the listeners-template](src/endpoint-configuration-service/src/main/resources/templates/listener.yaml.mustache) 
would need to add the client-certificate to the request. Since the endpoint-auth-service uses [envoy](https://www.envoyproxy.io) and therefor can be integrated with a 
service-mesh like [istio](https://istio.io/), mTLS probably is easier to apply with the mesh.

As a conclusion, the endpoint-auth-service should not be seen as a competition or alternative to mTLS, but rather an option for supporting other and maybe additional
auth-methods, like the implemented [iShare-solution](src/ishare-auth-provider/README.md).
