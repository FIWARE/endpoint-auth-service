package org.fiware.sidecar.exception;

public class DeletionException extends RuntimeException{


	public DeletionException(String message) {
		super(message);
	}

	public DeletionException(String message, Throwable cause) {
		super(message, cause);
	}
}
