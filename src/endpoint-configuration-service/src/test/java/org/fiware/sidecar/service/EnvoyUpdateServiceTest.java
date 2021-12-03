package org.fiware.sidecar.service;

import com.github.mustachejava.DefaultMustacheFactory;
import com.github.mustachejava.MustacheFactory;
import org.fiware.sidecar.configuration.ProxyProperties;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.mapping.EndpointMapperImpl;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.List;
import java.util.UUID;
import java.util.stream.Stream;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

class EnvoyUpdateServiceTest {

	private static final MustacheFactory MUSTACHE_FACTORY = new DefaultMustacheFactory();
	private static final EndpointMapper ENDPOINT_MAPPER = new EndpointMapperImpl();
	private EndpointRepository endpointRepository;
	private ProxyProperties proxyProperties;
	private EnvoyUpdateService envoyUpdateService;


	@BeforeEach
	public void setup() {
		endpointRepository = mock(EndpointRepository.class);
	}
//
//	@ParameterizedTest
//	@MethodSource("provideConfig")
//	public void applyConfiguration(ProxyProperties proxyProperties, List<Endpoint> endpoints, String expectedResultPath) {
//		envoyUpdateService = new EnvoyUpdateService(MUSTACHE_FACTORY, endpointRepository, ENDPOINT_MAPPER, proxyProperties);
//
//		when(endpointRepository.findAll()).thenReturn(endpoints);
//		envoyUpdateService.applyConfiguration();
//
//	}
//
//	static Stream<Arguments> provideConfig() {
//		return Stream.of(
//				Arguments.of(
//						new ProxyProperties("listener.yaml", "cluster.yaml",
//								new ProxyProperties.AddressConfig("ext-auth", 7070),
//								new ProxyProperties.AddressConfig("0.0.0.0", 15001)),
//						List.of(getEndpoint(UUID.randomUUID(), "domain", "/", 8080, AuthType.ISHARE, true)),
//						"test/expected"));
//	}
//
//	private static Endpoint getEndpoint(UUID uuid, String domain, String path, int port, AuthType authType, boolean https) {
//		var endpoint = new Endpoint();
//		endpoint.setId(uuid);
//		endpoint.setDomain(domain);
//		endpoint.setPath(path);
//		endpoint.setPort(port);
//		endpoint.setAuthType(authType);
//		endpoint.setUseHttps(https);
//		return endpoint;
//	}

}