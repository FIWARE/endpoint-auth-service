package org.fiware.sidecar.rest;

import io.micronaut.core.util.AntPathMatcher;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.Controller;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.api.AuthConfigurationApi;
import org.fiware.sidecar.mapping.EndpointMapper;
import org.fiware.sidecar.model.AuthInfoVO;
import org.fiware.sidecar.persistence.EndpointRepository;

import java.util.Comparator;

@Slf4j
@Controller
@RequiredArgsConstructor
public class AuthConfigurationApiController implements AuthConfigurationApi {

	private static final PathPartMatcher PATH_PART_MATCHER = new PathPartMatcher();

	private final EndpointRepository endpointRepository;
	private final EndpointMapper endpointMapper;

	@Override
	public HttpResponse<AuthInfoVO> getEndpointByDomainAndPath(String domain, String path) {

		return endpointRepository.findByDomain(domain).stream()
				.filter(endpoint -> PATH_PART_MATCHER.matchesPartly(path, endpoint.getPath()))
				// we want the longest match(e.g. /test instead of / when /test/path is asked)
				.max(Comparator.comparingInt(endpoint -> endpoint.getPath().length()))
				.map(endpointMapper::endpointToAuthInfoVo)
				.map(HttpResponse::ok)
				.orElse(HttpResponse.notFound());
	}
}
