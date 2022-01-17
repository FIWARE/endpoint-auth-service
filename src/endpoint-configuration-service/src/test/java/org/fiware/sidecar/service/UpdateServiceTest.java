package org.fiware.sidecar.service;

import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.persistence.Endpoint;
import org.junit.jupiter.params.provider.Arguments;

import java.util.List;
import java.util.UUID;
import java.util.stream.Stream;

public abstract class UpdateServiceTest {

	public static Stream<Arguments> getTestConfig() {
		return Stream.of(
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/", 6060, AuthType.ISHARE, true)),
						"src/test/resources/expectations/%s/single-endpoint/"),
				Arguments.of(
						List.of(),
						"src/test/resources/expectations/%s/empty/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain-2", "/", 6070, AuthType.ISHARE, false)),
						"src/test/resources/expectations/%s/multi-endpoint/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/", 6060, AuthType.ISHARE, false)),
						"src/test/resources/expectations/%s/single-endpoint-no-ssl/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/nonRoot", 6060, AuthType.ISHARE, true)),
						"src/test/resources/expectations/%s/single-non-root-path/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/path1", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain", "/path2", 6060, AuthType.ISHARE, true)),
						"src/test/resources/expectations/%s/single-endpoint-multi-path/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/path1", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain", "/path2", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain-2", "/", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain-2", "/nonRoot", 6070, AuthType.ISHARE, false)),
						"src/test/resources/expectations/%s/multi-endpoint-multi-path/")
		);
	}

	private static Endpoint getEndpoint(UUID uuid, String domain, String path, int port, AuthType authType, boolean https) {
		var endpoint = new Endpoint();
		endpoint.setId(uuid);
		endpoint.setDomain(domain);
		endpoint.setPath(path);
		endpoint.setPort(port);
		endpoint.setAuthType(authType);
		endpoint.setUseHttps(https);
		return endpoint;
	}
}
