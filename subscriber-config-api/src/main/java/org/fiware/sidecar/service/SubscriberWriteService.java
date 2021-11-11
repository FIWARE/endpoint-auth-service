package org.fiware.sidecar.service;

import org.fiware.sidecar.exception.CredentialsConfigNotFound;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.SubscriberRegistrationVO;

import java.util.UUID;

/**
 * Service interface for all operations that alter a subscriber. While read is equal for all subscriber types, the credentials will differ and therefor the
 * write operations have to depend on the auth-type.
 */
public interface SubscriberWriteService {

	/**
	 * Should return the auth-type supported by this write-service.
	 * @return the {@link AuthType}
	 */
	AuthType supportedAuthType();

	/**
	 * Do all creations that are required for this specific subscriber type
	 * @param subscriberRegistrationVO the subscriber registration to create the subscriber of
	 */
	void createSubscriber(SubscriberRegistrationVO subscriberRegistrationVO);

	/**
	 * Delete everything specific for this subscriber type
	 * @param uuid id of the subscriber to be deleted.
	 */
	void  deleteSubscriber(UUID uuid);

	/**
	 * Update type specific credential parts
	 * @param id id of susbcriber to be updated
	 * @param credentialType type of the credential to be updated(f.e. signingKey, certificateChain, username, password)
	 * @param credentialBody body holding the credential information
	 */
	void updateSubscriberCredential(UUID id, String  credentialType, String  credentialBody) throws CredentialsConfigNotFound;
}
