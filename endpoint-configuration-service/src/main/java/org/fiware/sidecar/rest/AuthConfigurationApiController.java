package org.fiware.sidecar.rest;

import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.Controller;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.api.AuthConfigurationApi;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.model.AuthInfoVO;
import org.fiware.sidecar.persistence.EndpointRepository;

@Slf4j
@Controller
@RequiredArgsConstructor
public class AuthConfigurationApiController implements AuthConfigurationApi {

	private final EndpointRepository endpointRepository;
	private final EndpointMapper endpointMapper;

	@Override
	public HttpResponse<AuthInfoVO> getEndpointByDomainAndPath(String domain, String path) {
		return endpointRepository.findByDomainAndPath(domain, path)
				.map(endpointMapper::endpointToAuthInfoVo)
				.map(HttpResponse::ok)
				.orElse(HttpResponse.notFound());
	}
}
