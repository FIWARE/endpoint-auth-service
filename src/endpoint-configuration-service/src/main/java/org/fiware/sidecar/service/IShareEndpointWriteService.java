package org.fiware.sidecar.service;

import org.fiware.sidecar.exception.CredentialsConfigNotFound;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.EndpointRegistrationVO;

import javax.inject.Singleton;
import java.util.UUID;

@Singleton
public class IShareEndpointWriteService implements EndpointWriteService{
	@Override
	public AuthType supportedAuthType() {
		return AuthType.ISHARE;
	}

	@Override
	public void createEndpoint(UUID uuid, EndpointRegistrationVO endpointRegistrationVO) {
		//noop - nothing special to do for iShare
	}

	@Override
	public void deleteEndpoint(UUID uuid) {
		//noop - nothing special to do for iShare
	}

	@Override
	public void updateEndpointCredential(UUID id, String credentialType, String credentialBody) throws CredentialsConfigNotFound {
		//noop - nothing special to do for iShare
	}
}
