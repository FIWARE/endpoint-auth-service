package org.fiware.sidecar.service;

import com.github.mustachejava.DefaultMustacheFactory;
import com.github.mustachejava.Mustache;
import com.github.mustachejava.MustacheFactory;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.configuration.ProxyProperties;
import org.fiware.sidecar.exception.EnvoyUpdateException;
import org.fiware.sidecar.mapping.SubscriberMapper;
import org.fiware.sidecar.model.MustacheSubscriber;
import org.fiware.sidecar.persistence.SubscriberRepository;

import javax.annotation.PostConstruct;
import javax.inject.Singleton;
import java.io.FileWriter;
import java.io.IOException;
import java.io.StringWriter;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

@Slf4j
@RequiredArgsConstructor
@Singleton
public class EnvoyUpdateService {

	private final MustacheFactory mustacheFactory;
	private final SubscriberRepository subscriberRepository;
	private final SubscriberMapper subscriberMapper;
	private final ProxyProperties proxyProperties;

	private Mustache listenerTemplate;
	private Mustache clusterTemplate;

	@PostConstruct
	public void setupTemplates() {
		listenerTemplate = mustacheFactory.compile("./templates/listener.yaml.mustache");
		clusterTemplate = mustacheFactory.compile("./templates/cluster.yaml.mustache");
	}

	public void applyConfiguration() {
		List<MustacheSubscriber> mustacheSubscriberList = StreamSupport
				.stream(subscriberRepository.findAll().spliterator(), true)
				.map(subscriberMapper::subscriberToMustacheSubscriber)
				.collect(Collectors.toList());

		Map<String, Object> mustacheRenderContext = Map.of("subscribers", mustacheSubscriberList);

		try {
			FileWriter listenerFileWriter = new FileWriter(proxyProperties.getListenerYamlPath());
			listenerTemplate.execute(listenerFileWriter, mustacheRenderContext).flush();
			listenerFileWriter.close();
		} catch (IOException e) {
			throw new EnvoyUpdateException("Was not able to update listener.yaml", e);
		}

		try {
			FileWriter clusterFileWriter = new FileWriter(proxyProperties.getClusterYamlPath());
			clusterTemplate.execute(clusterFileWriter, mustacheRenderContext).flush();
			clusterFileWriter.close();
		} catch (IOException e) {
			throw new EnvoyUpdateException("Was not able to update cluster.yaml", e);
		}
	}

}
