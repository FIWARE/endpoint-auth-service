package org.fiware.sidecar.rest;

import io.micronaut.http.HttpRequest;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.MutableHttpResponse;
import io.micronaut.http.annotation.Controller;
import io.micronaut.http.annotation.Delete;
import io.micronaut.http.annotation.Get;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Controller(value = "/last", port = "${inspector.port}")
public class RequestInspectorController {

	private List<HttpRequest> lastRequests = new ArrayList<>();

	public void addLastRequest(HttpRequest request) {
		lastRequests.add(request);
	}

	private HttpRequest getLastStoredRequest() {
		if(lastRequests.isEmpty()) {
			return null;
		}
		return lastRequests.get(lastRequests.size() - 1);
	}

	@Get("/all")
	public HttpResponse<List> getAllRequests() {

		return HttpResponse.ok(lastRequests
				.stream()
				.map(this::mapRequest)
				.collect(Collectors.toList()));
	}

	private StoredRequest mapRequest(HttpRequest request) {
		Object body = request.getBody().orElse(null);
		Map<String, String> headers = request.getHeaders().asMap().entrySet()
				.stream()
				.collect(Collectors.toMap(Map.Entry::getKey, e -> e.getValue().stream().findFirst().orElse("")));
		return new StoredRequest(headers, body);
	}

	@Get
	public HttpResponse<Object> getLastRequests() {
		if (getLastStoredRequest() == null) {
			return HttpResponse.notFound();
		}

		MutableHttpResponse response = HttpResponse.ok();
		if (getLastStoredRequest().getBody().isPresent()) {
			response.body(getLastStoredRequest().getBody());
		}
		Map<String, String> headers = getLastStoredRequest().getHeaders().asMap().entrySet()
				.stream()
				.collect(Collectors.toMap(Map.Entry::getKey, e -> e.getValue().stream().findFirst().orElse("")));

		headers.put("Requested-Path", getLastStoredRequest().getPath());
		response.headers(headers);

		return response;
	}

	@Delete
	public HttpResponse<Object> deleteLastRequest() {
		lastRequests = new ArrayList<>();
		return HttpResponse.noContent();
	}

	@Get("/headers/{header}")
	public HttpResponse<String> getHeaderFromLastRequest(String header) {
		if (lastRequests == null) {
			return HttpResponse.notFound();
		}
		return HttpResponse.ok(getLastStoredRequest().getHeaders().get(header));
	}

	@Get("/body")
	public HttpResponse<Object> getBodyFromLastRequest() {
		if (lastRequests == null) {
			return HttpResponse.notFound();
		}
		return HttpResponse.ok(getLastStoredRequest().getBody().orElse(null));
	}
}
