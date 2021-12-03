package org.fiware.sidecar.rest;

import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.mapping.EndpointMapperImpl;
import org.fiware.sidecar.model.AuthCredentialsVO;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.AuthTypeVO;
import org.fiware.sidecar.model.EndpointInfoVO;
import org.fiware.sidecar.model.EndpointRegistrationVO;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;
import org.fiware.sidecar.service.IShareEndpointWriteService;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

class EndpointConfigurationApiControllerTest {

	private EndpointConfigurationApiController endpointConfigurationApiController;
	private EndpointRepository endpointRepository;
	private static final EndpointMapper ENDPOINT_MAPPER = new EndpointMapperImpl();


	@BeforeEach
	public void setup() {
		endpointRepository = mock(EndpointRepository.class);
		endpointConfigurationApiController = new EndpointConfigurationApiController(List.of(new IShareEndpointWriteService()), endpointRepository, ENDPOINT_MAPPER);
	}

	@ParameterizedTest
	@MethodSource("endpointConfigurationValues")
	public void createEndpoint_success(AuthTypeVO authTypeVO, String domain, Integer port, String path, Boolean useHttps) {
		EndpointRegistrationVO endpointRegistrationVO = new EndpointRegistrationVO()
				.authType(authTypeVO)
				.domain(domain)
				.port(port)
				.path(path)
				.useHttps(useHttps)
				.authCredentials(new AuthCredentialsVO()
						.iShareClientId("clientId")
						.iShareIdpId("idpId")
						.iShareIdpAddress("http://ishard-idp.de")
						.requestGrantType("client_credentials"));

		Endpoint e = new Endpoint();
		e.setId(UUID.randomUUID());
		when(endpointRepository.save(any())).thenReturn(e);

		HttpResponse<Object> response = endpointConfigurationApiController.createEndpoint(endpointRegistrationVO);
		Assertions.assertEquals(HttpStatus.CREATED, response.getStatus(), "The endpoint should have been created.");
		response.getHeaders().findFirst("Location").equals(e.getId().toString());
	}


	@Test
	public void createEndpoint_duplicateDomain() {
		EndpointRegistrationVO endpointRegistrationVO = new EndpointRegistrationVO()
				.authType(AuthTypeVO.ISHARE)
				.domain("domain");

		when(endpointRepository.findByDomainAndPath(any(), any())).thenReturn(Optional.of(new Endpoint()));

		HttpResponse<Object> response = endpointConfigurationApiController.createEndpoint(endpointRegistrationVO);
		Assertions.assertEquals(HttpStatus.CONFLICT, response.getStatus(), "The endpoint should not have been created.");
	}

	@Test
	public void deleteEndpoint_success() {
		var endpointUUID = UUID.randomUUID();
		var endpoint = new Endpoint();
		endpoint.setAuthType(AuthType.ISHARE);
		endpoint.setId(endpointUUID);

		when(endpointRepository.findById(any())).thenReturn(Optional.of(endpoint));
		HttpResponse<Object> response = endpointConfigurationApiController.deleteEndpoint(UUID.randomUUID());
		Assertions.assertEquals(HttpStatus.NO_CONTENT, response.getStatus(), "The endpoint should have been deleted.");
	}

	@Test
	public void deleteEndpoint_notFound() {

		when(endpointRepository.findById(any())).thenReturn(Optional.empty());
		HttpResponse<Object> response = endpointConfigurationApiController.deleteEndpoint(UUID.randomUUID());
		Assertions.assertEquals(HttpStatus.NOT_FOUND, response.getStatus(), "The endpoint does not exist.");
	}

	@Test
	public void getEndpointInfo_success() {
		var uuid = UUID.randomUUID();
		var endpoint = new Endpoint();
		endpoint.setId(uuid);
		endpoint.setDomain("domain.org");
		endpoint.setIShareClientId("clientId");
		endpoint.setPath("/");
		endpoint.setPort(6060);
		endpoint.setAuthType(AuthType.ISHARE);
		endpoint.setRequestGrantType("client_credentials");
		endpoint.setUseHttps(true);

		when(endpointRepository.findById(any())).thenReturn(Optional.of(endpoint));

		HttpResponse<EndpointInfoVO> response = endpointConfigurationApiController.getEndpointInfo(uuid);
		Assertions.assertEquals(HttpStatus.OK, response.getStatus(), "An endpoint info should be successfully returned.");
		EndpointInfoVO endpointInfoVO = response.body();
		Assertions.assertEquals(AuthTypeVO.ISHARE, endpointInfoVO.getAuthType(), "Correct auth type should be returend.");
		Assertions.assertEquals("domain.org", endpointInfoVO.getDomain(), "Correct domain should be returend.");
		Assertions.assertEquals("/", endpointInfoVO.getPath(), "Correct path should be returend.");
		Assertions.assertEquals(true, endpointInfoVO.getUseHttps(), "Correct http should be returend.");
	}

	@Test
	public void getEndpointInfo_notFound() {
		when(endpointRepository.findById(any())).thenReturn(Optional.empty());

		Assertions.assertNull(endpointConfigurationApiController.getEndpointInfo(UUID.randomUUID()), "No such endpoint should be found.");
	}

	@Test
	public void getEndpoints_success() {
		List<Endpoint> endpoints = List.of(getEndpoint(UUID.randomUUID(), "domain1", "/", 7070, AuthType.ISHARE, true),
				getEndpoint(UUID.randomUUID(), "domain1", "/sub", 8080, AuthType.ISHARE, true),
				getEndpoint(UUID.randomUUID(), "localhost", "/", 7070, AuthType.ISHARE, false));

		when(endpointRepository.findAll()).thenReturn(endpoints);
		HttpResponse<List<EndpointInfoVO>> response = endpointConfigurationApiController.getEndpoints();
		Assertions.assertEquals(HttpStatus.OK, response.getStatus(), "A list of endpoints should be returned.");
		Assertions.assertEquals(3, response.body().size(), "All 3 endpoints should be returend.");
	}

	private Endpoint getEndpoint(UUID uuid, String domain, String path, int port, AuthType authType, boolean https) {
		var endpoint = new Endpoint();
		endpoint.setId(uuid);
		endpoint.setDomain(domain);
		endpoint.setPath(path);
		endpoint.setPort(port);
		endpoint.setAuthType(authType);
		endpoint.setUseHttps(https);
		return endpoint;
	}

	static Stream<Arguments> endpointConfigurationValues() {

		Arguments domains = Arguments.of("test.domain", "localhost", "127.0.0.1");
		Arguments ports = Arguments.of(6060, null);
		Arguments paths = Arguments.of("/", "/sub", "/sub/sub", null);
		Arguments useHttps = Arguments.of(true, false);

		List<Arguments> argumentsList = List.of(Arguments.of(AuthTypeVO.ISHARE));
		argumentsList = permutateArguments(argumentsList, domains);
		argumentsList = permutateArguments(argumentsList, ports);
		argumentsList = permutateArguments(argumentsList, paths);
		argumentsList = permutateArguments(argumentsList, useHttps);
		return argumentsList.stream();
	}

	private static List<Arguments> permutateArguments(List<Arguments> argList, Arguments newArgs) {
		return Arrays.stream(newArgs.get()).toList().stream().flatMap(p -> argList.stream().map(a -> appendArgument(a, p))).collect(Collectors.toList());
	}

	private static <T> Arguments appendArgument(Arguments a, Object o) {
		List argList = new ArrayList(Arrays.asList(a.get()));
		argList.add(o);
		return Arguments.of(argList.toArray());
	}
}