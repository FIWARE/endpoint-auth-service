package org.fiware.sidecar.configuration;

import io.micronaut.context.annotation.ConfigurationProperties;
import lombok.Data;

/**
 * Configuration of general properties
 */
@ConfigurationProperties("general")
@Data
public class GeneralProperties {


	/**
	 * Delay used when updating configurations. Should prevent to frequent config changes.
	 */
	private long updateDelayInS;
}
