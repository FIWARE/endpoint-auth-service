# Endpoint-Configuration-Service

The endpoint-configuration-service is the central service to manage and provide configuration about the endpoints to be handled by the sidecar-proxy.
The service stores information about the endpoints and provides them via [REST-Api](../../api/endpoint-configuration-api.yaml). It generates [dynamic resources](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/operations/dynamic_configuration)
for configuring [listners](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/lds#config-listeners-lds) and [clusters](https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cds#config-cluster-manager-cds)
on envoy and [ServiceMeshExtension's](https://docs.openshift.com/container-platform/4.6/service_mesh/v2x/ossm-extensions.html) for applying the configured auth via the sidecar-proxy in [Openshift Service Mesh](https://cloud.redhat.com/learn/topics/service-mesh).

## Build

The project uses [maven](https://maven.apache.org/) for build and dependency management.

## Install

The endpoint-configuration-service is part of the endpoint-authentication-service, so that the other components also need to be taken into account.
It uses a SQL-Database as backend, currently are [MySQL](https://www.mysql.com/) and the in-memory db [H2](https://www.h2database.com/html/main.html) are supported.

### How to run

Start the service through its container: ```docker run quay.io/repository/fiware/endpoint-configuration-service```

### Configuration

Since it is built using the [Micronaut-Framework](https://micronaut.io/) all configurations can be provided either via configuration
file ([application.yaml](src/main/resources/application.yml)) or as environment variables. For detailed information about the configuration mechanism,
see the [framework documentation](https://docs.micronaut.io/3.1.3/guide/index.html#configurationProperties).

The following table concentrates on the most important configuration parameters:

| Property                              | Env-Var                                          | Description                                                                      | Default                                                     |
|---------------------------------------|--------------------------------------------------|----------------------------------------------------------------------------------|-------------------------------------------------------------|
| `micronaut.server.port`               | `MICRONAUT_SERVER_PORT`                          | Server port to be used for mintaka                                               | 8080                                                        |
| `micronaut.metrics.enabled`           | `MICRONAUT_METRICS_ENABLED`                      | Enable the metrics gathering                                                     | true                                                        |
| `endpoints.all.port`                  | `ENDPOINTS_ALL_PORT`                             | Port to provide the management endpoints                                         | 8080                                                        |
| `endpoints.metrics.enabled`           | `ENDPOINTS_METRICS_ENABLED`                      | Enable the metrics endpoint                                                      | true                                                        |
| `endpoints.health.enabled`            | `ENDPOINTS_HEALTH_ENABLED`                       | Enable the health endpoint                                                       | true                                                        | 
| `datasources.default.host`            | `DATASOURCES_DEFAULT_URL`                        | URL for accessing db                                                             | jdbc:h2:mem:devDb;LOCK_TIMEOUT=10000;DB_CLOSE_ON_EXIT=FALSE |
| `datasources.default.username`        | `DATASOURCES_DEFAULT_USERNAME`                   | Username to be used for db connections                                           | sa                                                          | 
| `datasources.default.password`        | `DATASOURCES_DEFAULT_PASSWORD`                   | Password to be used for db connections                                           |                                                             | 
| `general.updateDelayInS`              | `GENERAL_UPDATE_DELAY_IN_S`                      | How much delay until the config generation shoudl start                          | 2                                                           |
| `envoy.enabled`                       | `ENVOY_ENABLED`                                  | Should configuration for envoy be generated.                                     | true                                                        |
| `envoy.externalAuth.address`          | `ENVOY_EXTERNAL_AUTH_ADDRESS`                    | Domain of the auth-provider                                                      | auth-service                                                |
| `envoy.externalAuth.port`             | `ENVOY_EXTERNAL_AUTH_PORT`                       | Port of the auth-provider                                                        | 7070                                                        |
| `envoy.listenerYamlPath`              | `ENVOY_LISTENER_YAML_PATH`                       | Path to store the generated listener.yaml                                        | ./listener.yaml                                             |
| `envoy.clusterYamlPath`               | `ENVOY_CLUSTER_YAML_PATH`                        | Path to store the generated cluster.yaml                                         | ./cluster.yaml                                              |
| `envoy.wasmFilterPath`                | `ENVOY_WASM_FILTER_PATH`                         | Path the cached-auth-filter wasm-file.                                           | /cache-filter/cache-filter.wasm                             |
| `meshExtension.enabled`               | `MESH_EXTENSION_ENABLED`                         | Should configuration for Openshift Service Mesh extensions be generated.         | true                                                        |
| `meshExtension.authProviderName`      | `MESH_EXTENSION_AUTH_PROVIDER_NAME`              | Name to access the auth provider, as defined by the service mesh.                |                                                             |
| `meshExtension.workloadSelector`      | `MESH_EXTENSION_WORKLOAD_SELECTOR_(NAME/VALUE)`  | Label to be used in the mesh extension for selecting the workloads to be handled | app/app                                                     |
| `meshExtension.filterVersion`         | `MESH_EXTENSION_FILTER_VERSION`                  | Version of the cached-auth-filter to be used in the extension                    | ${project.version}                                          |
| `meshExtension.extensionName`         | `MESH_EXTENSION_EXTENSION_NAME`                  | Name of the extension resource to be created inside Openshift                    | cached-auth-filter-extension                                |
| `meshExtension.extensionNamespace`    | `MESH_EXTENSION_EXTENSION_NAMESPACE`             | Namespace(Project) to create the resource at inside Openshift                    | extension-namespace                                         |
| `meshExtension.meshExtensionYamlPath` | `MESH_EXTENSION_MESH_EXTENSION_YAML_PATH`        | Path to generate the ServiceMeshExtension yaml-file at.                          | ./service-mesh-extension.yaml                               |


### Coverage

Code-coverage reports are automatically created by [Jacoco](https://www.eclemma.org/jacoco/) when the test are executed by maven. Public
reports are available at [Coveralls.io](https://coveralls.io/github/fiware/endpoint-auth-service).

### Static analyzes

Static code analyzes("linting") are provided via [Spotbugs](https://spotbugs.github.io/).
Reports can be created via: ```mvn -B verify spotbugs:spotbugs -DskipTests```

## Documentation

The code is documented in the [Javadoc comments format](https://docs.oracle.com/javase/1.5.0/docs/tooldocs/solaris/javadoc.html).