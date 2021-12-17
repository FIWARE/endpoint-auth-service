package org.fiware.sidecar.exception;

import lombok.Getter;

import java.util.List;

/**
 * Exception to be thrown if a credential is requested that was not found
 */
public class CredentialsConfigNotFound extends Exception {

	@Getter
	private final String credential;
	@Getter
	private final List<String> supportedCredentialConfigs;

	public CredentialsConfigNotFound(String message, String credential, List<String> supportedCredentialConfigs) {
		super(message);
		this.credential = credential;
		this.supportedCredentialConfigs = supportedCredentialConfigs;
	}

	public CredentialsConfigNotFound(String message, Throwable cause, String credential, List<String> supportedCredentialConfigs) {
		super(message, cause);
		this.credential = credential;
		this.supportedCredentialConfigs = supportedCredentialConfigs;
	}
}
