package org.fiware.sidecar.rest;

import lombok.Getter;
import lombok.RequiredArgsConstructor;

import java.util.Map;

@RequiredArgsConstructor
@Getter
public class StoredRequest {

	private final Map<String, String> headers;
	private final Object body;
}
