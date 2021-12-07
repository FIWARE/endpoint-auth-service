package org.fiware.sidecar.service;

import com.github.mustachejava.DefaultMustacheFactory;
import com.github.mustachejava.MustacheFactory;
import org.fiware.sidecar.configuration.ProxyProperties;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.mapping.EndpointMapperImpl;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.yaml.snakeyaml.Yaml;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;
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
	public void applyConfiguration() throws Exception {
		envoyUpdateService = new EnvoyUpdateService(MUSTACHE_FACTORY, endpointRepository, ENDPOINT_MAPPER, proxyProperties, Executors.newSingleThreadScheduledExecutor());
		envoyUpdateService.setupTemplates();

		UUID endpointID = UUID.randomUUID();
		when(endpointRepository.findAll()).thenReturn(List.of(getEndpoint(endpointID, "domain", "/", 6060, AuthType.ISHARE, true)));
		envoyUpdateService.applyConfiguration();
		Map<String, Object> expectedListener = getYamlAsMap(Path.of("src/test/resources/expectations/single-endpoint/listener.yaml"), Map.of("expected", endpointID.toString()));
		Map<String, Object> generatedListener = getYamlAsMap(Path.of("myTest/listener.yaml"), null);
//
//		Assertions.assertEquals(expectedListener, generatedListener, "Generated listener should be as expected.");
	}

	private Map<String, Object> getYamlAsMap(Path path, Map<String, String> replacementMap) throws Exception {
		String expectedString = Files.readString(path);
		if (replacementMap != null) {
			for (Map.Entry<String, String> replacementEntry : replacementMap.entrySet()) {
				expectedString = expectedString.replace(replacementEntry.getKey(), replacementEntry.getValue());
			}
		}
		Yaml yaml = new Yaml();
		return yaml.load(expectedString);
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