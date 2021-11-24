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

	// CATCH ALL MAPPINGS

	@Get
	public HttpResponse<Object> getEmptyEcho(HttpRequest request) {
		return echoRequest(request);
	}

	@Post
	public HttpResponse<Object> postEmptyEcho(HttpRequest request) {
		return echoRequest(request);
	}

	@Put
	public HttpResponse<Object> putEmptyEcho(HttpRequest request) {
		return echoRequest(request);
	}

	@Delete
	public HttpResponse<Object> deleteEmptyEcho(HttpRequest request) {
		return echoRequest(request);
	}

	@Get("/{+path}")
	public HttpResponse<Object> getEcho(HttpRequest request) {
		return echoRequest(request);
	}

	@Post("/{+path}")
	public HttpResponse<Object> postEcho(HttpRequest request) {
		return echoRequest(request);
	}

	@Put("/{+path}")
	public HttpResponse<Object> putEcho(HttpRequest request) {
		return echoRequest(request);
	}

	@Delete("/{+path}")
	public HttpResponse<Object> deleteEcho(HttpRequest request) {
		return echoRequest(request);
	}


	// REQUEST HANDLING

	private MutableHttpResponse echoRequest(HttpRequest request) {
		//store for inspection
		requestInspectorController.setLastRequest(request);

		Map<String, String> headers = request.getHeaders().asMap().entrySet()
				.stream()
				.collect(Collectors.toMap(Map.Entry::getKey, e -> e.getValue().stream().findFirst().orElse("")));

		log.info("{}  {}: Body: {}, Headers: {}",
				request.getMethod(),
				request.getPath(),
				request.getBody().map(Object::toString).orElse("null"),
				headers);

		MutableHttpResponse response = HttpResponse.ok();
		if (request.getBody().isPresent()) {
			response.body(request.getBody());
		}

		response.headers(headers);

		return response;
	}

}
