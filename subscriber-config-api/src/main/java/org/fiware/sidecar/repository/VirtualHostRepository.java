package org.fiware.sidecar.repository;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import org.fiware.sidecar.configuration.ProxyProperties;
import org.fiware.sidecar.model.EnvoyVirtualHost;
import org.yaml.snakeyaml.Yaml;

import javax.inject.Singleton;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;

@RequiredArgsConstructor
@Singleton
public class VirtualHostRepository {

	private final ProxyProperties proxyProperties;

	public EnvoyVirtualHost getVirtualHostById(UUID hostId) throws IOException {

		InputStream inputStream = new FileInputStream(new File(proxyProperties.getListenerYamlPath()));
		Yaml yaml = new Yaml();
		Map<String, Object> data = yaml.load(inputStream);

		if (!data.containsKey("resources") || !(data.get("resources") instanceof List)) {
			throw new IllegalArgumentException("Not a valid listener.yaml");
		}
		List listenerList = (List) data.get("resources");
		Optional optionalFilterChain = listenerList
				.stream()
				.filter(l -> l instanceof Map)
				.filter(l -> ((Map) l).containsKey("name") && ((Map) l).get("name").equals("envoy_listener"))
				.map(l -> ((Map) l).get("filter_chains"))
				.findFirst();
		if (optionalFilterChain.isEmpty() || !(optionalFilterChain.get() instanceof List)) {
			throw new IllegalArgumentException("Not a valid listener.yaml. Does not contain the envoy listeners filter chains.");
		}
		List filterChains = (List) optionalFilterChain.get();
		Optional optionalHttpFilters = filterChains
				.stream()
				.filter(fc -> fc instanceof Map)
				.filter(fc -> ((Map) fc).containsKey("name") && ((Map) fc).get("name").equals("http_chain"))
				.map(l -> ((Map) l).get("filters"))
				.findFirst();
		if (optionalHttpFilters.isEmpty() || !(optionalHttpFilters.get() instanceof List)) {
			throw new IllegalArgumentException("Not a valid listener.yaml. Does not contain the filter list for the http_filter chain.");
		}
		List httpFilters = (List) optionalHttpFilters.get();
		Optional optionalTypedConfig = httpFilters.stream()
				.filter(hf -> hf instanceof Map)
				.filter(hf -> ((Map) hf).containsKey("name") && ((Map) hf).get("name").equals("envoy.filters.network.http_connection_manager"))
				.map(hf ->((Map) hf).get("typed_config"))
				.findFirst();
		if (optionalTypedConfig.isEmpty() || !(optionalTypedConfig.get() instanceof List)) {
			throw new IllegalArgumentException("Not a valid listener.yaml. Does not contain the typed config for the httpConnectionManager.");
		}

		return null;
	}


}
