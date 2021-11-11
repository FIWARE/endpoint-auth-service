package org.fiware.sidecar.model;

import java.util.UUID;

public record Subscriber(UUID id, String domain, int port, String path, boolean useHttps, AuthType authType, IShareAuthCredentials iShareAuthCredentials){};
