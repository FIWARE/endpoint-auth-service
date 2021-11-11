package org.fiware.sidecar.exception;

public class FileUpdateException extends RuntimeException {

	public FileUpdateException(String message) {
		super(message);
	}

	public FileUpdateException(String message, Throwable cause) {
		super(message, cause);
	}
}
