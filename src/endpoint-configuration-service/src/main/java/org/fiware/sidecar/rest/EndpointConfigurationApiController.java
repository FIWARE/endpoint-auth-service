package org.fiware.sidecar.rest;

import io.micronaut.core.util.AntPathMatcher;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import io.micronaut.http.annotation.Controller;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.api.EndpointConfigurationApi;
import org.fiware.sidecar.exception.CredentialsConfigNotFound;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.AuthTypeVO;
import org.fiware.sidecar.model.EndpointInfoVO;
import org.fiware.sidecar.model.EndpointRegistrationVO;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;
import org.fiware.sidecar.service.EndpointWriteService;
import org.fiware.sidecar.service.EnvoyUpdateService;

import javax.transaction.Transactional;
import javax.validation.Valid;
import java.net.URI;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.StreamSupport;

@Slf4j
@Controller
@RequiredArgsConstructor
public class EndpointConfigurationApiController implements EndpointConfigurationApi {

	private final List<EndpointWriteService> subscriberWriteServices;
	private final EndpointRepository endpointRepository;
	private final EndpointMapper endpointMapper;

	@Transactional
	@Override
	public HttpResponse<Object> createEndpoint(EndpointRegistrationVO endpointRegistrationVO) {

		if (!endpointRegistrationVO.authType().equals(AuthTypeVO.ISHARE)) {
			throw new UnsupportedOperationException("Currently only iShare-authentication is supported.");
		}

		if (endpointRepository.findByDomainAndPath(endpointRegistrationVO.getDomain(), endpointRegistrationVO.getPath()).isPresent()) {
			return HttpResponse.status(HttpStatus.CONFLICT);
		}

		Endpoint endpoint = endpointRepository.save(endpointMapper.endpointRegistrationVoToEndpoint(endpointRegistrationVO));

		// type specific creations
		getServiceForAuthType(endpointMapper.authTypeVoToAuthType(endpointRegistrationVO.authType()))
				.createEndpoint(endpoint.getId(), endpointRegistrationVO);

		return HttpResponse.created(URI.create(endpoint.getId().toString()));
	}

	@Transactional
	@Override
	public HttpResponse<Object> deleteEndpoint(UUID id) {
		Optional<Endpoint> optionalSubscriber = endpointRepository.findById(id);
		if (optionalSubscriber.isPresent()) {
			endpointRepository.deleteById(id);
			getServiceForAuthType(optionalSubscriber.get().getAuthType()).deleteEndpoint(id);

			return HttpResponse.noContent();
		}
		return HttpResponse.notFound();
	}

	@Override
	public HttpResponse<EndpointInfoVO> getEndpointInfo(UUID id) {
		Optional<EndpointInfoVO> optionalSubscriberInfoVO = endpointRepository
				.findById(id)
				.map(endpointMapper::endpointToEndpointInfoVo);
		return optionalSubscriberInfoVO.<HttpResponse<EndpointInfoVO>>map(HttpResponse::ok).orElse(null);
	}

	@Override
	public HttpResponse<List<EndpointInfoVO>> getEndpoints() {
		return HttpResponse.ok(
				StreamSupport
						.stream(endpointRepository.findAll().spliterator(), true)
						.map(endpointMapper::endpointToEndpointInfoVo)
						.toList());
	}

	@Override
	public HttpResponse<Object> updateCredentialConfiguration(UUID id, String credential, String body) {
		Optional<Endpoint> optionalEndpoint = endpointRepository.findById(id);
		if (optionalEndpoint.isEmpty()) {
			return HttpResponse.notFound(String.format("Subscriber %s does not exist.", id));
		}
		try {
			getServiceForAuthType(optionalEndpoint.get().getAuthType()).updateEndpointCredential(id, credential, body);
		} catch (CredentialsConfigNotFound e) {
			return HttpResponse.notFound(
					String.format("Credential %s does not exist for subscriber %s. Only %s are supported.",
							e.getCredential(),
							id,
							e.getSupportedCredentialConfigs()));
		}

		// update of credentials do not demand an update of the envoy configuration, since envoy stays free of security concerns.
		return HttpResponse.noContent();
	}

	private EndpointWriteService getServiceForAuthType(AuthType authType) {
		return subscriberWriteServices
				.stream()
				.filter(sws -> sws.supportedAuthType()
						.equals(authType))
				.findFirst()
				.orElseThrow(() -> new UnsupportedOperationException(String.format("Auth type %s is not supported by this instance of the sidecar.", authType.getValue())));
	}
}
