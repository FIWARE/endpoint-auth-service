package org.fiware.sidecar.service;

import com.github.mustachejava.Mustache;
import com.github.mustachejava.MustacheFactory;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.configuration.ProxyProperties;
import org.fiware.sidecar.exception.EnvoyUpdateException;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.model.MustacheEndpoint;
import org.fiware.sidecar.persistence.EndpointRepository;

import javax.annotation.PostConstruct;
import javax.inject.Singleton;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.StreamSupport;

@Slf4j
@RequiredArgsConstructor
@Singleton
public class EnvoyUpdateService {

	private final MustacheFactory mustacheFactory;
	private final EndpointRepository endpointRepository;
	private final EndpointMapper endpointMapper;
	private final ProxyProperties proxyProperties;

	private Mustache listenerTemplate;
	private Mustache clusterTemplate;

	@PostConstruct
	public void setupTemplates() {
		listenerTemplate = mustacheFactory.compile("./templates/listener.yaml.mustache");
		clusterTemplate = mustacheFactory.compile("./templates/cluster.yaml.mustache");
	}

	public void applyConfiguration() {

		List<MustacheEndpoint> mustacheSubscriberList = StreamSupport
				.stream(endpointRepository.findAll().spliterator(), true)
				.map(endpointMapper::endpointToMustacheEndpoint)
				.toList();


		ProxyProperties.AddressConfig socketAddress = proxyProperties.getSocketAddress();
		ProxyProperties.AddressConfig authAddress = proxyProperties.getExternalAuth();

		Map<String, Object> mustacheRenderContext = new HashMap<>();
		mustacheRenderContext.put("socket-address", socketAddress.getAddress());
		mustacheRenderContext.put("socket-port", socketAddress.getPort());
		mustacheRenderContext.put("auth-service-address", authAddress.getAddress());
		mustacheRenderContext.put("auth-service-port", authAddress.getPort());
		mustacheRenderContext.put("endpoints", mustacheSubscriberList);

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
