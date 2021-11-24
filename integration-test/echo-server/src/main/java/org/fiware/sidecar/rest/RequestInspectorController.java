package org.fiware.sidecar.rest;

import io.micronaut.http.HttpRequest;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.MutableHttpResponse;
import io.micronaut.http.annotation.Controller;
import io.micronaut.http.annotation.Get;
import lombok.Getter;
import lombok.Setter;

import java.util.Map;
import java.util.stream.Collectors;

@Controller(value = "/last", port = "${inspector.port}")
public class RequestInspectorController {

	@Setter
	private HttpRequest lastRequest = null;

	@Get
	public HttpResponse<Object> getLastRequest() {
		if (lastRequest == null) {
			return HttpResponse.notFound();
		}

		MutableHttpResponse response = HttpResponse.ok();
		if (lastRequest.getBody().isPresent()) {
			response.body(lastRequest.getBody());
		}
		Map<String, String> headers = lastRequest.getHeaders().asMap().entrySet()
				.stream()
				.collect(Collectors.toMap(e -> e.getKey(), e -> e.getValue().stream().findFirst().orElse("")));

		response.headers(headers);

		return response;
	}

	@Get("/headers/{header}")
	public HttpResponse<String> getHeaderFromLastRequest(String header) {
		return HttpResponse.ok(lastRequest.getHeaders().get(header));
	}

	@Get("/body")
	public HttpResponse<Object> getBodyFromLastRequest() {
		return HttpResponse.ok(lastRequest.getBody().orElse(null));
	}
}
