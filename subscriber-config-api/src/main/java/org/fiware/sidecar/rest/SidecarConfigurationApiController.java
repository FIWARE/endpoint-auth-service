package org.fiware.sidecar.rest;

import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.Controller;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.api.SidecarConfigurationApi;
import org.fiware.sidecar.mapping.SubscriberMapper;
import org.fiware.sidecar.model.SubscriberInfoVO;
import org.fiware.sidecar.model.SubscriberRegistrationVO;
import org.fiware.sidecar.persistence.SubscriberRepository;
import org.fiware.sidecar.repository.VirtualHostRepository;

import java.io.IOException;
import java.net.URI;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

@Slf4j
@Controller
@RequiredArgsConstructor
public class SidecarConfigurationApiController implements SidecarConfigurationApi {

	private final VirtualHostRepository virtualHostRepository;
	private final SubscriberRepository subscriberRepository;
	private final SubscriberMapper subscriberMapper;

	@Override
	public HttpResponse<Object> createSubscriber(SubscriberRegistrationVO subscriberRegistrationVO) {
		return HttpResponse.created(URI.create(subscriberRepository.save(subscriberMapper.subscriberRegistrationVoToSubscriber(subscriberRegistrationVO)).getId()));
	}

	@Override
	public HttpResponse<Object> deleteSubscriber(UUID id) {
		return null;
	}

	@Override
	public HttpResponse<SubscriberInfoVO> getSubscriberInfo(UUID id) {
		return null;
	}

	@Override
	public HttpResponse<List<SubscriberInfoVO>> getSubscribers() {
		return HttpResponse.ok(
				StreamSupport
						.stream(subscriberRepository.findAll().spliterator(), true)
						.map(subscriberMapper::subscriberToSubscriberInfoVo)
						.collect(Collectors.toList()));
	}
}
