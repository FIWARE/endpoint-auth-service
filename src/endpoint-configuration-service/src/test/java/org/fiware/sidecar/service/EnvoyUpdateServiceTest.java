package org.fiware.sidecar.service;

import com.github.mustachejava.DefaultMustacheFactory;
import com.github.mustachejava.MustacheFactory;
import org.fiware.sidecar.configuration.EnvoyProperties;
import org.fiware.sidecar.configuration.GeneralProperties;
import org.fiware.sidecar.configuration.MeshExtensionProperties;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.mapping.EndpointMapperImpl;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
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
	private GeneralProperties generalProperties;
	private String testId;

	@BeforeEach
	public void setup() throws Exception {
		generalProperties = new GeneralProperties();
		generalProperties.setUpdateDelayInS(2);
		endpointRepository = mock(EndpointRepository.class);
		testId = UUID.randomUUID().toString();
		Files.createDirectory(Path.of(String.format("./%s", testId)));
	}

	private MeshExtensionProperties getMeshExtensionProperties() {

		MeshExtensionProperties meshExtensionProperties = new MeshExtensionProperties();

		meshExtensionProperties.setAuthProviderName("outbount|8080||auth-provider");
		meshExtensionProperties.setFilterVersion("1.0.0");
		meshExtensionProperties.setExtensionName("my-extension");
		meshExtensionProperties.setExtensionNamespace("my-extension-namespace");
		meshExtensionProperties.setMeshExtensionYamlPath(String.format("./%s/service-mesh-extension.yaml", testId));

		MeshExtensionProperties.MetaData label = new MeshExtensionProperties.MetaData();
		label.setName("my-label");
		label.setValue("my-label-value");

		MeshExtensionProperties.MetaData annotation = new MeshExtensionProperties.MetaData();
		annotation.setName("my-annotation");
		annotation.setValue("my-annotation-value");
		meshExtensionProperties.setLabels(List.of(label));
		meshExtensionProperties.setAnnotations(List.of(annotation));

		MeshExtensionProperties.MetaData workloadSelector = new MeshExtensionProperties.MetaData();
		workloadSelector.setName("my-workload");
		workloadSelector.setValue("selected");
		meshExtensionProperties.setWorkloadSelector(workloadSelector);

		return meshExtensionProperties;
	}

	private EnvoyProperties getEnvoyProperties() {

		EnvoyProperties envoyProperties = new EnvoyProperties();
		envoyProperties.setListenerYamlPath(String.format("./%s/listener.yaml", testId));
		envoyProperties.setClusterYamlPath(String.format("./%s/cluster.yaml", testId));
		EnvoyProperties.AddressConfig authService = new EnvoyProperties.AddressConfig();
		authService.setPort(7070);
		authService.setAddress("auth-service");
		EnvoyProperties.AddressConfig socketAddress = new EnvoyProperties.AddressConfig();
		socketAddress.setPort(15001);
		socketAddress.setAddress("0.0.0.0");
		envoyProperties.setExternalAuth(authService);
		envoyProperties.setSocketAddress(socketAddress);
		return envoyProperties;
	}

//	// deletes all generated results. Disable this mechanism if you need to debug them.
//	@AfterEach
//	public void cleanUpResults() throws Exception {
//		try (Stream<Path> walk = Files.walk(Path.of(String.format("./%s", testId)))) {
//			walk.sorted(Comparator.reverseOrder())
//					.map(Path::toFile)
//					.forEach(File::delete);
//		}
//	}

	@ParameterizedTest
	@MethodSource("getTestConfig")
	public void applyConfigurationToServiceMesh(List<Endpoint> endpoints, String expectationsFolder) throws Exception {
		ServiceMeshUpdateService meshUpdateService = new ServiceMeshUpdateService(MUSTACHE_FACTORY, endpointRepository, ENDPOINT_MAPPER, Executors.newSingleThreadScheduledExecutor(), generalProperties, getMeshExtensionProperties());
		meshUpdateService.setupTemplates();

		when(endpointRepository.findAll()).thenReturn(endpoints);
		meshUpdateService.applyConfiguration();

		Map<String, String> replacementMap = buildIdReplacementMap(endpoints);

		Map<String, Object> expectedListener = getYamlAsMap(Path.of(String.format("%s/service-mesh-extension.yaml", String.format(expectationsFolder, "meshExtension"))), null);
		Map<String, Object> generatedListener = getYamlAsMap(Path.of(String.format("%s/service-mesh-extension.yaml", testId)), replacementMap);
		Assertions.assertEquals(expectedListener, generatedListener, "Generated service-mesh-extension should be as expected.");

	}

	@ParameterizedTest
	@MethodSource("getTestConfig")
	public void applyConfigurationToEnvoy(List<Endpoint> endpoints, String expectationsFolder) throws Exception {
		EnvoyUpdateService envoyUpdateService = new EnvoyUpdateService(MUSTACHE_FACTORY, endpointRepository, ENDPOINT_MAPPER, Executors.newSingleThreadScheduledExecutor(), generalProperties, getEnvoyProperties());
		envoyUpdateService.setupTemplates();

		when(endpointRepository.findAll()).thenReturn(endpoints);
		envoyUpdateService.applyConfiguration();

		Map<String, String> replacementMap = buildIdReplacementMap(endpoints);

		Map<String, Object> expectedListener = getYamlAsMap(Path.of(String.format("%s/listener.yaml", String.format(expectationsFolder, "envoy"))), null);
		Map<String, Object> generatedListener = getYamlAsMap(Path.of(String.format("%s/listener.yaml", testId)), replacementMap);
		Assertions.assertEquals(expectedListener, generatedListener, "Generated listener should be as expected.");

		Map<String, Object> expectedCluster = getYamlAsMap(Path.of(String.format("%s/cluster.yaml", String.format(expectationsFolder, "envoy"))), null);
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