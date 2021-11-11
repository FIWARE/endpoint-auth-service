package org.fiware.sidecar.rest;

import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import io.micronaut.http.annotation.Controller;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.api.SidecarConfigurationApi;
import org.fiware.sidecar.exception.CredentialsConfigNotFound;
import org.fiware.sidecar.exception.DeletionException;
import org.fiware.sidecar.exception.FileCreationException;
import org.fiware.sidecar.exception.FolderCreationException;
import org.fiware.sidecar.mapping.SubscriberMapper;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.AuthTypeVO;
import org.fiware.sidecar.model.SubscriberInfoVO;
import org.fiware.sidecar.model.SubscriberRegistrationVO;
import org.fiware.sidecar.persistence.IShareCredentialsRepository;
import org.fiware.sidecar.persistence.Subscriber;
import org.fiware.sidecar.persistence.SubscriberRepository;
import org.fiware.sidecar.service.EnvoyUpdateService;
import org.fiware.sidecar.service.SubscriberWriteService;

import javax.transaction.Transactional;
import java.net.URI;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

@Slf4j
@Controller
@RequiredArgsConstructor
public class SidecarConfigurationApiController implements SidecarConfigurationApi {

	private final List<SubscriberWriteService> subscriberWriteServices;
	private final SubscriberRepository subscriberRepository;
	private final SubscriberMapper subscriberMapper;
	private final EnvoyUpdateService envoyUpdateService;

	@Transactional
	@Override
	public HttpResponse<Object> createSubscriber(SubscriberRegistrationVO subscriberRegistrationVO) {

		if (!subscriberRegistrationVO.authType().equals(AuthTypeVO.ISHARE)) {
			throw new UnsupportedOperationException("Currently only iShare-authentication is supported.");
		}

		if (subscriberRepository.findByDomainAndPath(subscriberRegistrationVO.getDomain(), subscriberRegistrationVO.getPath()).isPresent()) {
			return HttpResponse.status(HttpStatus.CONFLICT);
		}

		Subscriber subscriber = subscriberRepository.save(subscriberMapper.subscriberRegistrationVoToSubscriber(subscriberRegistrationVO));

		// type specific creations
		getServiceForAuthType(subscriberMapper.authTypeVoToAuthType(subscriberRegistrationVO.authType()))
				.createSubscriber(subscriberRegistrationVO);

		// update the envoy configuration
		envoyUpdateService.applyConfiguration();

		return HttpResponse.created(URI.create(subscriber.getId().toString()));
	}

	@Transactional
	@Override
	public HttpResponse<Object> deleteSubscriber(UUID id) {
		Optional<Subscriber> optionalSubscriber = subscriberRepository.findById(id);
		if (optionalSubscriber.isPresent()) {
			subscriberRepository.deleteById(id);
			getServiceForAuthType(optionalSubscriber.get().getAuthType()).deleteSubscriber(id);

			// update the envoy configuration
			envoyUpdateService.applyConfiguration();

			return HttpResponse.noContent();
		}
		return HttpResponse.notFound();
	}

	@Override
	public HttpResponse<SubscriberInfoVO> getSubscriberInfo(UUID id) {
		Optional<SubscriberInfoVO> optionalSubscriberInfoVO = subscriberRepository
				.findById(id)
				.map(subscriberMapper::subscriberToSubscriberInfoVo);
		if (optionalSubscriberInfoVO.isPresent()) {
			return HttpResponse.ok(optionalSubscriberInfoVO.get());
		}
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

	@Override
	public HttpResponse<Object> updateCredentialConfiguration(UUID id, String credential, String body) {
		Optional<Subscriber> optionalSubscriber = subscriberRepository.findById(id);
		if (!optionalSubscriber.isPresent()) {
			HttpResponse.notFound(String.format("Subscriber %s does not exist.", id));
		}
		try {
			getServiceForAuthType(optionalSubscriber.get().getAuthType()).updateSubscriberCredential(id, credential, body);
		} catch (CredentialsConfigNotFound e) {
			HttpResponse.notFound(
					String.format("Credential %s does not exist for subscriber %s. Only %s are supported.",
							e.getCredential(),
							id,
							e.getSupportedCredentialConfigs()));
		}

		// update of credentials do not demand an update of the envoy configuration, since envoy stays free of security concerns.

		return HttpResponse.noContent();
	}

	private SubscriberWriteService getServiceForAuthType(AuthType authType) {
		return subscriberWriteServices
				.stream()
				.filter(sws -> sws.supportedAuthType()
						.equals(authType))
				.findFirst()
				.orElseThrow(() -> new UnsupportedOperationException(String.format("Auth type %s is not supported by this instance of the sidecar.", authType.getValue())));
	}
}
