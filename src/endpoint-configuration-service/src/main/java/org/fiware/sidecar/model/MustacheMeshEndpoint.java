package org.fiware.sidecar.model;

import java.util.List;

public record MustacheMeshEndpoint(String authType, List<MustacheEndpointDomain> domains) {
}
