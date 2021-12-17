package org.fiware.sidecar.service;

import com.github.mustachejava.DefaultMustacheFactory;
import com.github.mustachejava.MustacheFactory;
import org.fiware.sidecar.configuration.ProxyProperties;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.mapping.EndpointMapperImpl;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;
import org.junit.jupiter.api.AfterEach;
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
import java.util.Comparator;
import java.util.HashMap;
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
	private String testId;

	@BeforeEach
	public void setup() throws Exception {
		endpointRepository = mock(EndpointRepository.class);
		proxyProperties = new ProxyProperties();
		testId = UUID.randomUUID().toString();
		Files.createDirectory(Path.of(String.format("./%s", testId)));

		proxyProperties.setListenerYamlPath(String.format("./%s/listener.yaml", testId));
		proxyProperties.setClusterYamlPath(String.format("./%s/cluster.yaml", testId));
		ProxyProperties.AddressConfig authService = new ProxyProperties.AddressConfig();
		authService.setPort(7070);
		authService.setAddress("auth-service");
		ProxyProperties.AddressConfig socketAddress = new ProxyProperties.AddressConfig();
		socketAddress.setPort(15001);
		socketAddress.setAddress("0.0.0.0");
		proxyProperties.setExternalAuth(authService);
		proxyProperties.setSocketAddress(socketAddress);
	}

	// deletes all generated results. Disable this mechanism if you need to debug them.
	@AfterEach
	public void cleanUpResults() throws Exception {
		try (Stream<Path> walk = Files.walk(Path.of(String.format("./%s", testId)))) {
			walk.sorted(Comparator.reverseOrder())
					.map(Path::toFile)
					.forEach(File::delete);
		}
	}

	@ParameterizedTest
	@MethodSource("getTestConfig")
	public void applyConfiguration(List<Endpoint> endpoints, String expectationsFolder) throws Exception {
		envoyUpdateService = new EnvoyUpdateService(MUSTACHE_FACTORY, endpointRepository, ENDPOINT_MAPPER, proxyProperties, Executors.newSingleThreadScheduledExecutor());
		envoyUpdateService.setupTemplates();

		when(endpointRepository.findAll()).thenReturn(endpoints);
		envoyUpdateService.applyConfiguration();

		Map<String, String> replacementMap = buildIdReplacementMap(endpoints);

		Map<String, Object> expectedListener = getYamlAsMap(Path.of(String.format("%s/listener.yaml", expectationsFolder)), null);
		Map<String, Object> generatedListener = getYamlAsMap(Path.of(String.format("%s/listener.yaml", testId)), replacementMap);
		Assertions.assertEquals(expectedListener, generatedListener, "Generated listener should be as expected.");

		Map<String, Object> expectedCluster = getYamlAsMap(Path.of(String.format("%s/cluster.yaml", expectationsFolder)), null);
		Map<String, Object> generatedCluster = getYamlAsMap(Path.of(String.format("%s/cluster.yaml", testId)), replacementMap);
		Assertions.assertEquals(expectedCluster, generatedCluster, "Generated cluster should be as expected.");
	}

	private Map<String, String> buildIdReplacementMap(List<Endpoint> endpoints) {
		Map<String, String> replacementMap = new HashMap<>();
		for (int i = 0; i < endpoints.size(); i++) {
			replacementMap.put(endpoints.get(i).getId().toString(), String.format("expected-%s", i));
		}
		return replacementMap;
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

	private static Stream<Arguments> getTestConfig() {
		return Stream.of(
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/", 6060, AuthType.ISHARE, true)),
						"src/test/resources/expectations/single-endpoint/"),
				Arguments.of(
						List.of(),
						"src/test/resources/expectations/empty/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain-2", "/", 6070, AuthType.ISHARE, false)),
						"src/test/resources/expectations/multi-endpoint/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/", 6060, AuthType.ISHARE, false)),
						"src/test/resources/expectations/single-endpoint-no-ssl/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/nonRoot", 6060, AuthType.ISHARE, true)),
						"src/test/resources/expectations/single-non-root-path/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/path1", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain", "/path2", 6060, AuthType.ISHARE, true)),
						"src/test/resources/expectations/single-endpoint-multi-path/"),
				Arguments.of(
						List.of(getEndpoint(UUID.randomUUID(), "domain", "/path1", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain", "/path2", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain-2", "/", 6060, AuthType.ISHARE, true),
								getEndpoint(UUID.randomUUID(), "domain-2", "/nonRoot", 6070, AuthType.ISHARE, false)),
						"src/test/resources/expectations/multi-endpoint-multi-path/")
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