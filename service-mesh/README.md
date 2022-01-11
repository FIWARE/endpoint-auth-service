# Service Mesh integration

If you already have a service mesh running on your cluster, a sidecar-proxy is probably also already in place. In order to not interfer with the mesh functionality, its not recommended to run the sidecar-envoy, but instead use the [wasm-filter directly](../src/cached-auth-filter) to integrate with the mesh.
In general, every [Service-Mesh](https://en.wikipedia.org/wiki/Service_mesh) that uses a proxy supporting the [proxy-wasm ABI](https://github.com/proxy-wasm/spec). Currently, the solution is only tested with the [OpenShift Service Mesh](https://cloud.redhat.com/learn/topics/service-mesh) and therefor also [istio](https://istio.io/). 

## Precondition

For the solution to properly work, the [endpoint-configuration-service](https://quay.io/repository/fiware/endpoint-configuration-service) and the [auth-provider](../doc/AUTHPROVIDER.md)(currently only [iShare](https://quay.io/repository/fiware/ishare-auth-provider)) should be deployed. 
The [helm-chart](https://github.com/FIWARE/helm-charts/tree/main/charts/endpoint-auth-service can be used for that, only the [sidecar-injection](https://github.com/FIWARE/helm-charts/tree/main/charts/endpoint-auth-service#sidecar-injection) needs to be disabled.

## OpenShift Service Mesh

The OpenShift Service Mesh supports extensions via [WebAssembly](https://docs.openshift.com/container-platform/4.6/service_mesh/v2x/ossm-extensions.html). In order to install the [cached-auth-filter](../src/cached-auth-filter), a ServiceMeshExtension compatible container is provided: [cached-auth-filter-extension](https://quay.io/repository/fiware/cached-auth-filter-extension)

An example for installing the mesh-extension can be found under the [example-folder](./openshift/example/extension.yaml). To allow the plugin to talk with the auth-provider, two ways are possible:
- the auth-provider is already included into the mesh, then nothing beside providing its address to the plugin via config is required
- if the auth-provider is not prat of the mesh, a [ServiceEntry](https://docs.openshift.com/container-platform/4.6/service_mesh/v2x/ossm-traffic-manage.html#ossm-routing-se_routing-traffic) is required. See the [example-folder](./openshift/example/service-entry.yaml) for an example.

## Envoy in the mesh

The current stable version of [envoy](https://www.envoyproxy.io/) is 1.20.1. Since the [MatchingApi](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api) is not productive yet, the plugin either needs to be configured to handle every call or the ```general.enableEndpointMatching``` feature needs to be enabled and the endpoints to be handle. See the [plugin documentation](../src/cached-auth-filter/README.md) for more.