package org.fiware.sidecar.model;

public record MustacheSubscriber(String id, String domain, String path, String httpsPort,  int port) {
}
