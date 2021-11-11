package org.fiware.sidecar.configuration;

import io.micronaut.context.annotation.ConfigurationProperties;
import lombok.Data;

@ConfigurationProperties("iShare")
@Data
public class IShareAuth {

	/**
	 * Path to the root folder for storing the
	 */
	private String certificateFolderPath;
}
