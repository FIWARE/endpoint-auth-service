package org.fiware.sidecar.rest;

import io.micronaut.http.HttpRequest;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.MutableHttpResponse;
import io.micronaut.http.annotation.Controller;
import io.micronaut.http.annotation.Delete;
import io.micronaut.http.annotation.Get;
import io.micronaut.http.annotation.Post;
import io.micronaut.http.annotation.Put;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Slf4j
@RequiredArgsConstructor
@Controller
public class EchoController {

	private final RequestInspectorController requestInspectorController;

	@Get("/{+path}")
	public HttpResponse<Object> getEcho(HttpRequest request, String path) {
		return echoRequest(request);
	}

	@Post("/{+path}")
	public HttpResponse<Object> postEcho(HttpRequest request, String path) {
		return echoRequest(request);
	}

	@Put("/{+path}")
	public HttpResponse<Object> putEcho(HttpRequest request, String path) {
		return echoRequest(request);
	}

	@Delete("/{+path}")
	public HttpResponse<Object> deleteEcho(HttpRequest request, String path) {
		return echoRequest(request);
	}

	private MutableHttpResponse echoRequest(HttpRequest request) {
		//store for inspection
		requestInspectorController.setLastRequest(request);

		Map<String, String> headers = request.getHeaders().asMap().entrySet()
				.stream()
				.collect(Collectors.toMap(e -> e.getKey(), e -> e.getValue().stream().findFirst().orElse("")));

		log.debug("{}  {}: Body: {}, Headers: {}",
				request.getMethod(),
				request.getPath(),
				request.getBody().map(b -> b.toString()).orElse("null"),
				headers);

		MutableHttpResponse response = HttpResponse.ok();
		if (request.getBody().isPresent()) {
			response.body(request.getBody());
		}

		response.headers(headers);

		return response;
	}

}
