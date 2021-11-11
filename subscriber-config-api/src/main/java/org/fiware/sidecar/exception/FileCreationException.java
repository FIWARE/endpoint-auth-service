package org.fiware.sidecar.exception;

public class FileCreationException extends RuntimeException {

	private final String filePath;

	public FileCreationException(String message, String filePath) {
		super(message);
		this.filePath = filePath;
	}

	public FileCreationException(String message, Throwable cause, String filePath) {
		super(message, cause);
		this.filePath = filePath;
	}
}
