package org.fiware.sidecar.service;

import com.github.mustachejava.Mustache;
import com.github.mustachejava.MustacheFactory;
import io.micronaut.context.annotation.Context;
import io.micronaut.context.annotation.Requires;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.configuration.EnvoyProperties;
import org.fiware.sidecar.configuration.GeneralProperties;
import org.fiware.sidecar.configuration.MeshExtensionProperties;
import org.fiware.sidecar.exception.EnvoyUpdateException;
import org.fiware.sidecar.exception.MeshExtensionUpdateException;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.model.MustacheEndpoint;
import org.fiware.sidecar.model.MustacheEndpointDomain;
import org.fiware.sidecar.model.MustacheMeshEndpoint;
import org.fiware.sidecar.model.MustacheMetaData;
import org.fiware.sidecar.model.MustachePath;
import org.fiware.sidecar.model.MustacheVirtualHost;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;

import javax.annotation.PostConstruct;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ScheduledExecutorService;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

/**
 * Service to create and update the service mesh extension configuration
 */
@Slf4j
@Context
@Requires(property = "meshExtension.enabled", value = "true")
public class ServiceMeshUpdateService extends MustacheUpdateService {

	private final MeshExtensionProperties meshExtensionProperties;

	private Mustache meshExtensionTemplate;

	public ServiceMeshUpdateService(MustacheFactory mustacheFactory, EndpointRepository endpointRepository, EndpointMapper endpointMapper, ScheduledExecutorService executorService, GeneralProperties generalProperties, MeshExtensionProperties meshExtensionProperties) {
		super(mustacheFactory, endpointRepository, endpointMapper, executorService, generalProperties);
		this.meshExtensionProperties = meshExtensionProperties;
	}

	@PostConstruct
	public void setupTemplates() {
		meshExtensionTemplate = mustacheFactory.compile("./templates/service-mesh-extension.yaml.mustache");
	}

	/**
	 * Apply the actual configuration, retrieved from the repository
	 */
	@Override
	void applyConfiguration() {
		Map<String, Map<String, List<Endpoint>>> endpointsByAuthType = new LinkedHashMap<>();

		StreamSupport
				// do not stream in parallel, will create duplicate keys in the map
				.stream(endpointRepository.findAll().spliterator(), false)
				.forEach(endpoint -> {
					String authType = endpoint.getAuthType().toString();
					Map<String, List<Endpoint>> endpointsByDomain = endpointsByAuthType.getOrDefault(authType, new LinkedHashMap<>());
					List<Endpoint> endpointList = endpointsByDomain.getOrDefault(endpoint.getDomain(), new ArrayList<>());
					endpointList.add(endpoint);
					endpointsByDomain.put(endpoint.getDomain(), endpointList);
					endpointsByAuthType.put(authType, endpointsByDomain);
				});

		List<MustacheMeshEndpoint> mustacheMeshEndpoints = endpointsByAuthType
				.entrySet()
				.stream()
				.map(entry -> new MustacheMeshEndpoint(entry.getKey(), endpointsByDomainToMustache(entry.getValue()))).collect(Collectors.toList());

		List<MustacheMetaData> annotations = meshExtensionProperties.getAnnotations().stream()
				.map(endpointMapper::metaDataToMustacheMetadata)
				.collect(Collectors.toList());

		List<MustacheMetaData> labels = meshExtensionProperties.getLabels().stream()
				.map(endpointMapper::metaDataToMustacheMetadata)
				.collect(Collectors.toList());

		Map<String, Object> mustacheRenderContext = new HashMap<>();
		mustacheRenderContext.put("extensionName", meshExtensionProperties.getExtensionName());
		mustacheRenderContext.put("extensionNamespace", meshExtensionProperties.getExtensionNamespace());
		mustacheRenderContext.put("authProviderName", meshExtensionProperties.getAuthProviderName());
		mustacheRenderContext.put("selectorLabel", meshExtensionProperties.getWorkloadSelector().getName());
		mustacheRenderContext.put("selectorValue", meshExtensionProperties.getWorkloadSelector().getValue());
		mustacheRenderContext.put("filterVersion", meshExtensionProperties.getFilterVersion());
		mustacheRenderContext.put("labels", labels);
		mustacheRenderContext.put("annotations", annotations);
		mustacheRenderContext.put("meshEndpoints", mustacheMeshEndpoints);

		if (!Files.exists(Path.of(meshExtensionProperties.getMeshExtensionYamlPath()))) {
			try {
				Files.createFile(Path.of(meshExtensionProperties.getMeshExtensionYamlPath()));
			} catch (IOException e) {
				throw new EnvoyUpdateException("Was not able to create mesh-extension yaml.", e);
			}
		}

		try {
			FileWriter meshExtensionFileWriter = new FileWriter(meshExtensionProperties.getMeshExtensionYamlPath());
			meshExtensionTemplate.execute(meshExtensionFileWriter, mustacheRenderContext).flush();
			meshExtensionFileWriter.close();
		} catch (IOException e) {
			throw new MeshExtensionUpdateException("Was not able to generate and write the service-mesh-extension yaml-file.", e);
		}
	}

	private List<MustacheEndpointDomain> endpointsByDomainToMustache(Map<String, List<Endpoint>> endpointsByDomainMap) {
		return endpointsByDomainMap.entrySet().stream()
				.map(entry -> new MustacheEndpointDomain(entry.getKey(), endpointsToMustachePaths(entry.getValue())))
				.collect(Collectors.toList());
	}

	private List<MustachePath> endpointsToMustachePaths(List<Endpoint> endpoints) {
		return endpoints.stream().map(e -> new MustachePath(e.getPath())).collect(Collectors.toList());
	}

	@Override
	List<MustacheEndpoint> extendEndpointList(List<MustacheEndpoint> endpointList) {
		//NO-OP, only to fulfill the interface
		return endpointList;
	}
}
