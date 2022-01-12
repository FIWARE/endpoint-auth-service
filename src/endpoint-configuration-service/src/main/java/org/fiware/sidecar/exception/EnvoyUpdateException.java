package org.fiware.sidecar.exception;

/**
 * Exception to be thrown if updating the envoy config failed.
 */
public class EnvoyUpdateException extends RuntimeException {

	public EnvoyUpdateException(String message, Throwable cause) {
		super(message, cause);
	}
}
