package org.fiware.sidecar.model.ishare;

import lombok.Getter;

import java.util.Arrays;

public enum IShareAuthCredentialType {

	KEY("key.pem", "signingKey"),
	CERT_CHAIN("cert.cer", "certificateChain");

	@Getter
	private final String fileName;
	@Getter
	private final String credentialType;

	IShareAuthCredentialType(String fileName, String credentialType) {
		this.fileName = fileName;
		this.credentialType = credentialType;
	}

	public static IShareAuthCredentialType getForCredentialType(String credentialType) {
		return Arrays.stream(values())
				.filter(v -> v.credentialType.equals(credentialType))
				.findFirst()
				.orElseThrow(() -> new IllegalArgumentException(String.format("No credentialsType %s exists.", credentialType)));
	}
}
