package org.fiware.sidecar.service;

import com.github.mustachejava.Mustache;
import com.github.mustachejava.MustacheFactory;
import io.micronaut.context.annotation.Context;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.configuration.ProxyProperties;
import org.fiware.sidecar.exception.EnvoyUpdateException;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.model.MustacheAuthType;
import org.fiware.sidecar.model.MustacheEndpoint;
import org.fiware.sidecar.model.MustachePort;
import org.fiware.sidecar.model.MustacheVirtualHost;
import org.fiware.sidecar.persistence.EndpointRepository;

import javax.annotation.PostConstruct;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

/**
 * Service to update the envoy configuration
 */
@Slf4j
@RequiredArgsConstructor
@Context
public class EnvoyUpdateService {

	private static final MustacheEndpoint PASSTHROUGH_ENDPOINT = new MustacheEndpoint("passthrough", "not-used", "/", null, "true", 0, "");

	private final MustacheFactory mustacheFactory;
	private final EndpointRepository endpointRepository;
	private final EndpointMapper endpointMapper;
	private final ProxyProperties proxyProperties;
	private final ScheduledExecutorService executorService;

	private Mustache listenerTemplate;
	private Mustache clusterTemplate;

	@PostConstruct
	public void setupTemplates() {
		listenerTemplate = mustacheFactory.compile("./templates/listener.yaml.mustache");
		clusterTemplate = mustacheFactory.compile("./templates/cluster.yaml.mustache");
	}

	/**
	 * Schedule a new configuration generation
	 */
	public void scheduleConfigUpdate() {
		executorService.schedule(this::applyConfiguration, proxyProperties.getUpdateDelayInS(), TimeUnit.SECONDS);
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

		Map<String, List<MustacheEndpoint>> endpointMap = StreamSupport
				.stream(endpointRepository.findAll().spliterator(), true)
				.map(endpointMapper::endpointToMustacheEndpoint)
				.collect(Collectors.toMap(MustacheEndpoint::domain, me -> new ArrayList<>(List.of(me)), (v1, v2) -> {
					v1.addAll(v2);
					return v1;
				}));

		List<MustacheVirtualHost> mustacheVirtualHosts = endpointMap
				.entrySet().stream()
				.map(entry -> new MustacheVirtualHost(
						entry.getKey(),
						entry.getValue().stream()
								.map(MustacheEndpoint::port)
								.distinct()
								.map(MustachePort::new)
								.collect(Collectors.toSet()),
						addPassThroughIfNoRoot(entry.getValue())))
				.toList();


		ProxyProperties.AddressConfig socketAddress = proxyProperties.getSocketAddress();
		ProxyProperties.AddressConfig authAddress = proxyProperties.getExternalAuth();

		Map<String, Object> mustacheRenderContext = new HashMap<>();
		mustacheRenderContext.put("socket-address", socketAddress.getAddress());
		mustacheRenderContext.put("socket-port", socketAddress.getPort());
		mustacheRenderContext.put("auth-service-address", authAddress.getAddress());
		mustacheRenderContext.put("auth-service-port", authAddress.getPort());
		mustacheRenderContext.put("virtualHosts", mustacheVirtualHosts);
		mustacheRenderContext.put("endpoints", mustacheEndpoints);
		mustacheRenderContext.put("authTypes", mustacheAuthTypes);

		if (!Files.exists(Path.of(proxyProperties.getListenerYamlPath()))) {
			try {
				Files.createFile(Path.of(proxyProperties.getListenerYamlPath()));
			} catch (IOException e) {
				throw new EnvoyUpdateException("Was not able to create listener.yaml", e);
			}
		}
		if (!Files.exists(Path.of(proxyProperties.getClusterYamlPath()))) {
			try {
				Files.createFile(Path.of(proxyProperties.getClusterYamlPath()));
			} catch (IOException e) {
				throw new EnvoyUpdateException("Was not able to create cluster.yaml", e);
			}
		}

		// BE AWARE: Order matters here. Due to the dynamic resource updates of envoy, first updating the listeners can lead to illegally referencing clusters
		// that do not yet exist.
		updateEnvoyConfig(proxyProperties.getClusterYamlPath(), clusterTemplate, mustacheRenderContext, "Was not able to update cluster.yaml");
		updateEnvoyConfig(proxyProperties.getListenerYamlPath(), listenerTemplate, mustacheRenderContext, "Was not able to update listener.yaml");
	}

	/*
	 * If only sub-paths are configured, add a passthrough route match for the domain
	 */
	private List<MustacheEndpoint> addPassThroughIfNoRoot(List<MustacheEndpoint> originalList) {
		Optional<MustacheEndpoint> optionalRootEndpoint = originalList.stream()
				.filter(mustacheEndpoint -> mustacheEndpoint.path().equals(PASSTHROUGH_ENDPOINT.path()))
				.findAny();
		if (optionalRootEndpoint.isEmpty()) {
			originalList.add(PASSTHROUGH_ENDPOINT);
		}
		return originalList;
	}

	private void updateEnvoyConfig(String proxyProperties, Mustache clusterTemplate, Map<String, Object> mustacheRenderContext, String message) {
		try {
			FileWriter clusterFileWriter = new FileWriter(proxyProperties);
			clusterTemplate.execute(clusterFileWriter, mustacheRenderContext).flush();
			clusterFileWriter.close();
		} catch (IOException e) {
			throw new EnvoyUpdateException(message, e);
		}
	}

}
