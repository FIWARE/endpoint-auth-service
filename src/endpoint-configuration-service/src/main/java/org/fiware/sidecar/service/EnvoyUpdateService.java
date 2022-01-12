package org.fiware.sidecar.service;

import com.github.mustachejava.Mustache;
import com.github.mustachejava.MustacheFactory;
import io.micronaut.context.annotation.Context;
import io.micronaut.context.annotation.Requires;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.configuration.EnvoyProperties;
import org.fiware.sidecar.configuration.GeneralProperties;
import org.fiware.sidecar.exception.EnvoyUpdateException;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.model.MustacheAuthType;
import org.fiware.sidecar.model.MustacheEndpoint;
import org.fiware.sidecar.model.MustacheVirtualHost;
import org.fiware.sidecar.persistence.EndpointRepository;

import javax.annotation.PostConstruct;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ScheduledExecutorService;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

/**
 * Service to create and update the envoy configuration
 */
@Slf4j
@Context
@Requires(property = "envoy.enabled", value = "true")
public class EnvoyUpdateService extends MustacheUpdateService{

	private static final MustacheEndpoint PASSTHROUGH_ENDPOINT = new MustacheEndpoint("passthrough", "not-used", "/", null, "true", 0, "");

	private final EnvoyProperties envoyProperties;

	private Mustache listenerTemplate;
	private Mustache clusterTemplate;

	public EnvoyUpdateService(MustacheFactory mustacheFactory, EndpointRepository endpointRepository, EndpointMapper endpointMapper, ScheduledExecutorService executorService, GeneralProperties generalProperties, EnvoyProperties envoyProperties) {
		super(mustacheFactory, endpointRepository, endpointMapper, executorService, generalProperties);
		this.envoyProperties = envoyProperties;
	}


	@PostConstruct
	public void setupTemplates() {
		listenerTemplate = mustacheFactory.compile("./templates/listener.yaml.mustache");
		clusterTemplate = mustacheFactory.compile("./templates/cluster.yaml.mustache");
	}

	/**
	 * Apply the actual configuration, retrieved from the repository
	 */
	void applyConfiguration() {

		List<MustacheEndpoint> mustacheEndpoints = StreamSupport
				.stream(endpointRepository.findAll().spliterator(), true)
				.map(endpointMapper::endpointToMustacheEndpoint)
				.toList();

		List<MustacheAuthType> mustacheAuthTypes = mustacheEndpoints
				.stream()
				.map(MustacheEndpoint::authType)
				.distinct()
				.map(MustacheAuthType::new)
				.collect(Collectors.toList());

		List<MustacheVirtualHost> mustacheVirtualHosts = getMustacheVirtualHosts();


		EnvoyProperties.AddressConfig socketAddress = envoyProperties.getSocketAddress();
		EnvoyProperties.AddressConfig authAddress = envoyProperties.getExternalAuth();

		Map<String, Object> mustacheRenderContext = new HashMap<>();
		mustacheRenderContext.put("socketAddress", socketAddress.getAddress());
		mustacheRenderContext.put("socketPort", socketAddress.getPort());
		mustacheRenderContext.put("authServiceAddress", authAddress.getAddress());
		mustacheRenderContext.put("authServicePort", authAddress.getPort());
		mustacheRenderContext.put("wasmFilterPath", envoyProperties.getWasmFilterPath());
		mustacheRenderContext.put("virtualHosts", mustacheVirtualHosts);
		mustacheRenderContext.put("endpoints", mustacheEndpoints);
		mustacheRenderContext.put("authTypes", mustacheAuthTypes);
		mustacheRenderContext.put("enableWasmFilter", mustacheAuthTypes.isEmpty() ? null : "true");

		if (!Files.exists(Path.of(envoyProperties.getListenerYamlPath()))) {
			try {
				Files.createFile(Path.of(envoyProperties.getListenerYamlPath()));
			} catch (IOException e) {
				throw new EnvoyUpdateException("Was not able to create listener.yaml", e);
			}
		}
		if (!Files.exists(Path.of(envoyProperties.getClusterYamlPath()))) {
			try {
				Files.createFile(Path.of(envoyProperties.getClusterYamlPath()));
			} catch (IOException e) {
				throw new EnvoyUpdateException("Was not able to create cluster.yaml", e);
			}
		}

		// BE AWARE: Order matters here. Due to the dynamic resource updates of envoy, first updating the listeners can lead to illegally referencing clusters
		// that do not yet exist.
		updateEnvoyConfig(envoyProperties.getClusterYamlPath(), clusterTemplate, mustacheRenderContext, "Was not able to update cluster.yaml");
		updateEnvoyConfig(envoyProperties.getListenerYamlPath(), listenerTemplate, mustacheRenderContext, "Was not able to update listener.yaml");
	}

	/*
	 * If only sub-paths are configured, add a passthrough route match for the domain
	 */
	@Override
	List<MustacheEndpoint> extendEndpointList(List<MustacheEndpoint> endpointList) {
		Optional<MustacheEndpoint> optionalRootEndpoint = endpointList.stream()
				.filter(mustacheEndpoint -> mustacheEndpoint.path().equals(PASSTHROUGH_ENDPOINT.path()))
				.findAny();
		if (optionalRootEndpoint.isEmpty()) {
			endpointList.add(PASSTHROUGH_ENDPOINT);
		}
		return endpointList;
	}

	private void updateEnvoyConfig(String configFilename, Mustache clusterTemplate, Map<String, Object> mustacheRenderContext, String message) {
		try {
			FileWriter clusterFileWriter = new FileWriter(configFilename);
			clusterTemplate.execute(clusterFileWriter, mustacheRenderContext).flush();
			clusterFileWriter.close();
		} catch (IOException e) {
			throw new EnvoyUpdateException(message, e);
		}
	}

}
