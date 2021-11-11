package org.fiware.sidecar.model;

import java.net.URL;

public record IShareAuthCredentials(String certChain, String signingKey, String iShareClientId, String iShareIdpId, URL iShareIdpAddress, String grantType) {
}
