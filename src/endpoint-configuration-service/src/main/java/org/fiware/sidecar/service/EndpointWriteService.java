package org.fiware.sidecar.service;

import org.fiware.sidecar.exception.CredentialsConfigNotFound;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.EndpointRegistrationVO;

import java.util.UUID;

/**
 * Service interface for all operations that alter a endpoint. While read is equal for all endpoint types, the credentials will differ and therefor the
 * write operations have to depend on the auth-type.
 */
public interface EndpointWriteService {

	/**
	 * Should return the auth-type supported by this write-service.
	 * @return the {@link AuthType}
	 */
	AuthType supportedAuthType();

	/**
	 * Do all creations that are required for this specific endpoint type
	 * @param  uuid id of the endpoint to be created
	 * @param endpointRegistrationVO the endpoint registration to create the endpoint of
	 */
	void createEndpoint(UUID uuid, EndpointRegistrationVO endpointRegistrationVO);

	/**
	 * Delete everything specific for this endpoint type
	 * @param uuid id of the endpoint to be deleted.
	 */
	void deleteEndpoint(UUID uuid);

	/**
	 * Update type specific credential parts
	 * @param id id of susbcriber to be updated
	 * @param credentialType type of the credential to be updated(f.e. signingKey, certificateChain, username, password)
	 * @param credentialBody body holding the credential information
	 */
	void updateEndpointCredential(UUID id, String  credentialType, String  credentialBody) throws CredentialsConfigNotFound;
}
