package org.fiware.sidecar.model;

import java.util.List;
import java.util.Set;

/**
 * Mustache representation of an envoy virtual host
 */
public record MustacheVirtualHost(String domain, Set<MustachePort> ports, List<MustacheEndpoint> endpoints) {
}
