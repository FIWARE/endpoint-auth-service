package org.fiware.sidecar.exception;

public class EnvoyUpdateException extends RuntimeException {

	public EnvoyUpdateException(String message) {
		super(message);
	}

	public EnvoyUpdateException(String message, Throwable cause) {
		super(message, cause);
	}
}
