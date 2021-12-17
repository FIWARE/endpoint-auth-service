package org.fiware.sidecar.configuration;

import io.micronaut.context.annotation.ConfigurationBuilder;
import io.micronaut.context.annotation.ConfigurationProperties;
import lombok.Data;
import lombok.Getter;
import lombok.Setter;

/**
 * Configuration of proxy(e.g. envoy) related properties
 */
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

	/**
	 * Delay to be used when updating envoy. Should prevent to frequent config changes.
	 */
	private long updateDelayInS;

	/**
	 * Address of the authentication provider
	 */
	@ConfigurationBuilder(configurationPrefix = "externalAuth")
	private AddressConfig externalAuth = new AddressConfig();

	/**
	 * Socket configuration to be used for envoy
	 * typical values for the address will be:
	 * * 0.0.0.0 if used in a sidecar/shared network approach
	 * * localhost if inside a container(f.e. docker-compose setups)
	 * standard port to be used for envoy is 15001
	 */
	@ConfigurationBuilder(configurationPrefix = "socketAddress")
	private AddressConfig socketAddress = new AddressConfig();

	@Setter
	@Getter
	public static class AddressConfig {

		private String address;
		private int port;
	}
}
