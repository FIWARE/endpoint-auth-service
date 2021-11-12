package org.fiware.sidecar.exception;

public class DeletionException extends RuntimeException{

	private final String path;

	public DeletionException(String message, String path) {
		super(message);
		this.path = path;
	}

	public DeletionException(String message, Throwable cause, String path) {
		super(message, cause);
		this.path = path;
	}
}
