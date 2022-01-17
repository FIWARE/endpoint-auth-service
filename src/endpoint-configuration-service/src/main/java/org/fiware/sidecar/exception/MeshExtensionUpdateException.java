package org.fiware.sidecar.exception;

/**
 * Exception to be thrown when updating the mesh extension file fails.
 */
public class MeshExtensionUpdateException extends RuntimeException{

	public MeshExtensionUpdateException(String message, Throwable cause) {
		super(message, cause);
	}
}
