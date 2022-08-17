package org.fiware.sidecar.service;

import com.github.mustachejava.MustacheFactory;
import lombok.RequiredArgsConstructor;
import org.fiware.sidecar.configuration.GeneralProperties;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.model.MustacheEndpoint;
import org.fiware.sidecar.model.MustachePort;
import org.fiware.sidecar.model.MustacheVirtualHost;
import org.fiware.sidecar.persistence.EndpointRepository;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

@RequiredArgsConstructor
public abstract class MustacheUpdateService implements UpdateService {

	protected final MustacheFactory mustacheFactory;
	protected final EndpointRepository endpointRepository;
	protected final EndpointMapper endpointMapper;
	protected final ScheduledExecutorService executorService;
	protected final GeneralProperties generalProperties;

	/**
	 * Schedule the update with a configurable delay.
	 */
	public void scheduleConfigUpdate() {
		executorService.schedule(this::applyConfiguration, generalProperties.getUpdateDelayInS(), TimeUnit.SECONDS);
	}

	/**
	 * Apply the actual configuration, retrieved from the repository
	 */
	abstract void applyConfiguration();

	protected List<MustacheVirtualHost> getMustacheVirtualHosts() {
		Map<String, List<MustacheEndpoint>> endpointMap = StreamSupport
				.stream(endpointRepository.findAll().spliterator(), true)
				.map(endpointMapper::endpointToMustacheEndpoint)
				.collect(Collectors.toMap(MustacheEndpoint::domain, me -> new ArrayList<>(List.of(me)), (v1, v2) -> {
					v1.addAll(v2);
					return v1;
				}));

		return endpointMap
				.entrySet().stream()
				.map(entry -> new MustacheVirtualHost(
						entry.getKey(),
						entry.getValue().stream()
								.map(MustacheEndpoint::port)
								.distinct()
								.map(MustachePort::new)
								.collect(Collectors.toSet()),
						extendEndpointList(entry.getValue())))
				.toList();
	}

	abstract List<MustacheEndpoint> extendEndpointList(List<MustacheEndpoint> endpointList);
}
