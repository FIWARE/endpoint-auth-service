package org.fiware.sidecar.configuration;

import io.micronaut.context.annotation.ConfigurationProperties;
import lombok.Data;

@ConfigurationProperties("proxy")
@Data
public class ProxyProperties {

	/**
	 * Path to the listener.yaml.mustache file used for configuration of the listeners in the envoy proxy.
	 */
	private String listenerYamlPath;

	/**
	 * Path to the cluster.yaml file used for configuration of the clusters in the envoy proxy.
	 */
	private String clusterYamlPath;
}
