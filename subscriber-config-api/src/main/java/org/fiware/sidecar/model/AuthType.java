package org.fiware.sidecar.model;

import lombok.Getter;

public enum AuthType {

	ISHARE("iShare");

	@Getter
	private final String value;

	AuthType(String value) {
		this.value = value;
	}

}
