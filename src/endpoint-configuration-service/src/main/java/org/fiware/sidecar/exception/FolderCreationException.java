package org.fiware.sidecar.exception;

public class FolderCreationException extends RuntimeException {

	private final String folderPath;

	public FolderCreationException(String message, String folderPath) {
		super(message);
		this.folderPath = folderPath;
	}

	public FolderCreationException(String message, Throwable cause, String folderPath) {
		super(message, cause);
		this.folderPath = folderPath;
	}
}
