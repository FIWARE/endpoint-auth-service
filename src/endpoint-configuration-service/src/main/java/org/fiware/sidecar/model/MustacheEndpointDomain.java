package org.fiware.sidecar.model;

import java.util.List;

public record MustacheEndpointDomain(String domain, List<MustachePath> paths) {
}
