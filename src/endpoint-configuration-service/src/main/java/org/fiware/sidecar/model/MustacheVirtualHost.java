package org.fiware.sidecar.model;

import java.util.List;
import java.util.Set;

public record MustacheVirtualHost(String id, String domain, Set<MustachePort> ports, List<MustacheEndpoint> endpoints) {
}
