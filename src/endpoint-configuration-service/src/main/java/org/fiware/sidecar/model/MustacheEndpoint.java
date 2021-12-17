package org.fiware.sidecar.model;

/**
 * Mustache representation of an endpoint
 */
public record MustacheEndpoint(String id, String domain, String path, String httpsPort, String passthrough, int port, String authType) {
}
