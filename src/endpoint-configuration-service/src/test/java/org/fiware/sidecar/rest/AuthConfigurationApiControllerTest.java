package org.fiware.sidecar.rest;

import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.mapping.EndpointMapperImpl;
import org.fiware.sidecar.model.AuthInfoVO;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

class AuthConfigurationApiControllerTest {

	private AuthConfigurationApiController authConfigurationApiController;
	private EndpointRepository endpointRepository;
	private static final EndpointMapper ENDPOINT_MAPPER = new EndpointMapperImpl();


	@BeforeEach
	public void setup() {
		endpointRepository = mock(EndpointRepository.class);
		authConfigurationApiController = new AuthConfigurationApiController(endpointRepository, ENDPOINT_MAPPER);
	}

	@ParameterizedTest
	@MethodSource("providePathMatchConfig")
	public void findClosestMatch(String expectedPath, String pathToAsk, List<String> pathList) {
		Endpoint expectedEndpoint = createEndpointWithPath(expectedPath);

		List<Endpoint> endpoints = pathList.stream().map(path -> createEndpointWithPath(path)).collect(Collectors.toList());
		endpoints.add(expectedEndpoint);
		when(endpointRepository.findByDomain("test.de"))
				.thenReturn(endpoints);
		HttpResponse<AuthInfoVO> response = authConfigurationApiController.getEndpointByDomainAndPath("test.de", pathToAsk);
		assertEquals(HttpStatus.OK, response.getStatus(), "An authInfo should be responded");
		assertEquals(expectedEndpoint.getIShareClientId(), response.body().getAdditionalProperties().get("iShareClientId"), String.format("The endpoint with path %s should be returned.", expectedPath));
	}

	private static Stream<Arguments> providePathMatchConfig() {
		return Stream.of(
				Arguments.of("/", "/", List.of("/test", "/test/path", "/path/test")),
				Arguments.of("/", "/tes", List.of("/test", "/test/path", "/path/test")),
				Arguments.of("/", "/p", List.of("/test", "/test/path", "/path/test")),
				Arguments.of("/", "/p/t", List.of("/test", "/test/path", "/path/test")),
				Arguments.of("/test", "/test", List.of("/", "/test/path", "/path/test")),
				Arguments.of("/test", "/test/p", List.of("/", "/test/path", "/path/test")),
				Arguments.of("/test", "/test/something/down/the/line", List.of("/", "/test/path", "/path/test")),
				Arguments.of("/test/path", "/test/path/down/the/line", List.of("/", "/test", "/path/test")),
				Arguments.of("/test/path", "/test/path/d", List.of("/", "/test", "/path/test"))
		);
	}

	private Endpoint createEndpointWithPath(String path) {
		Endpoint endpoint = new Endpoint();
		endpoint.setAuthType(AuthType.ISHARE);
		endpoint.setId(UUID.randomUUID());
		// use the path also as clientId, to have it visible in the test result
		endpoint.setIShareClientId(path);
		endpoint.setIShareIdpAddress("http://myIdp.de");
		endpoint.setPort(80);
		endpoint.setIShareIdpId("idpId");
		endpoint.setRequestGrantType("client_credentials");
		endpoint.setPath(path);
		return endpoint;
	}

}