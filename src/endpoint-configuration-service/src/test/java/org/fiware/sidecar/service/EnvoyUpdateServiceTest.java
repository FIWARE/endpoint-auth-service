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
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.List;
import java.util.UUID;
import java.util.concurrent.Executors;
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
		proxyProperties = new ProxyProperties();
		proxyProperties.setListenerYamlPath("./myTest/listener.yaml");
		proxyProperties.setClusterYamlPath("./myTest/cluster.yaml");
		ProxyProperties.AddressConfig authService = new ProxyProperties.AddressConfig();
		authService.setPort(7070);
		authService.setAddress("auth-service");
		ProxyProperties.AddressConfig socketAddress = new ProxyProperties.AddressConfig();
		socketAddress.setPort(15001);
		socketAddress.setAddress("0.0.0.0");
		proxyProperties.setExternalAuth(authService);
		proxyProperties.setSocketAddress(socketAddress);
	}

	@Test
	public void applyConfiguration() {
//		envoyUpdateService = new EnvoyUpdateService(MUSTACHE_FACTORY, endpointRepository, ENDPOINT_MAPPER, proxyProperties, Executors.newSingleThreadScheduledExecutor());
//		envoyUpdateService.setupTemplates();
//
//		when(endpointRepository.findAll()).thenReturn(List.of(getEndpoint(UUID.randomUUID(), "10.5.0.2", "/", 6060, AuthType.ISHARE, true)));
//		envoyUpdateService.applyConfiguration();

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