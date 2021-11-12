package org.fiware.sidecar.model;

public record MustacheEndpoint(String id, String domain, String path, String httpsPort, int port) {
}
