# Endpoint-Configuration-Service

The endpoint-configuration-service is the central service to manage and provide configuration about the endpoints to be handled by the sidecar-proxy.
The service stores information about the endpoints and provides them via [REST-Api](../../api/endpoint-configuration-api.yaml) and generates [dynamic resources](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/operations/dynamic_configuration)
for configuring [listners](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/lds#config-listeners-lds) and [clusters](https://www.envoyproxy.io/docs/envoy/latest/configuration/upstream/cluster_manager/cds#config-cluster-manager-cds)
on envoy.

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

| Property                       | Env-Var                        | Description                                             | Default                                                     |
|--------------------------------|--------------------------------|---------------------------------------------------------|-------------------------------------------------------------|
| `micronaut.server.port`        | `MICRONAUT_SERVER_PORT`        | Server port to be used for mintaka                      | 8080                                                        |
| `micronaut.metrics.enabled`    | `MICRONAUT_METRICS_ENABLED`    | Enable the metrics gathering                            | true                                                        |
| `endpoints.all.port`           | `ENDPOINTS_ALL_PORT`           | Port to provide the management endpoints                | 8080                                                        |
| `endpoints.metrics.enabled`    | `ENDPOINTS_METRICS_ENABLED`    | Enable the metrics endpoint                             | true                                                        |
| `endpoints.health.enabled`     | `ENDPOINTS_HEALTH_ENABLED`     | Enable the health endpoint                              | true                                                        | 
| `datasources.default.host`     | `DATASOURCES_DEFAULT_URL`      | URL for accessing db                                    | jdbc:h2:mem:devDb;LOCK_TIMEOUT=10000;DB_CLOSE_ON_EXIT=FALSE |
| `datasources.default.username` | `DATASOURCES_DEFAULT_USERNAME` | Username to be used for db connections                  | sa                                                          | 
| `datasources.default.password` | `DATASOURCES_DEFAULT_PASSWORD` | Password to be used for db connections                  |                                                             | 
| `proxy.externalAuth.address`   | `PROXY_EXTERNAL_AUTH_ADDRESS`  | Domain of the auth-provider                             | auth-service                                                |
| `proxy.externalAuth.port`      | `PROXY_EXTERNAL_AUTH_PORT`     | Port of the auth-provider                               | 7070                                                        |
| `proxy.listenerYamlPath`       | `PROXY_LISTENER_YAML_PATH`     | Path to store the generated listener.yaml               | ./listener.yaml                                             |
| `proxy.clusterYamlPath`        | `PROXY_CLUSTER_YAML_PATH`      | Path to store the generated cluster.yaml                | ./cluster.yaml                                              |
| `proxy.updateDelayInS`         | `PROXY_UPDATE_DELAY_IN_S`      | How much delay until the config generation shoudl start | 2                                                           |
| `proxy.wasmFilterPath`         | `PROXY_WASM_FILTER_PATH`       | Path the cached-auth-filter wasm-file.                  | /cache-filter/cache-filter.wasm                                                           |

### Coverage

Code-coverage reports are automatically created by [Jacoco](https://www.eclemma.org/jacoco/) when the test are executed by maven. Public
reports are available at [Coveralls.io](https://coveralls.io/github/fiware/endpoint-auth-service).

### Static analyzes

Static code analyzes("linting") are provided via [Spotbugs](https://spotbugs.github.io/).
Reports can be created via: ```mvn -B verify spotbugs:spotbugs -DskipTests```

## Documentation

The code is documented in the [Javadoc comments format](https://docs.oracle.com/javase/1.5.0/docs/tooldocs/solaris/javadoc.html).