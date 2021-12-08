package org.fiware.sidecar.model;

import lombok.Getter;

/**
 * Enum for the supported auth-types.
 * Needs to be extended for new types.
 */
public enum AuthType {

	ISHARE("iShare");

	@Getter
	private final String value;

	AuthType(String value) {
		this.value = value;
	}

}
