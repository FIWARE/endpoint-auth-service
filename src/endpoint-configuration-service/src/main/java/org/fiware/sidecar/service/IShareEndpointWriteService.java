package org.fiware.sidecar.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.exception.CredentialsConfigNotFound;
import org.fiware.sidecar.exception.DeletionException;
import org.fiware.sidecar.exception.FileCreationException;
import org.fiware.sidecar.exception.FolderCreationException;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.EndpointRegistrationVO;
import org.fiware.sidecar.model.ishare.IShareAuthCredentialType;
import org.fiware.sidecar.persistence.Endpoint;
import org.fiware.sidecar.persistence.EndpointRepository;
import org.fiware.sidecar.persistence.IShareCredentialsRepository;

import javax.inject.Singleton;
import java.util.Arrays;
import java.util.Optional;
import java.util.UUID;

@Slf4j
@RequiredArgsConstructor
@Singleton
public class IShareEndpointWriteService implements EndpointWriteService {

	private final IShareCredentialsRepository iShareCredentialsRepository;
	private final EndpointRepository endpointRepository;

	@Override
	public AuthType supportedAuthType() {
		return AuthType.ISHARE;
	}

	@Override
	public void createEndpoint(UUID id, EndpointRegistrationVO subscriberRegistrationVO) {
		try {
			iShareCredentialsRepository.saveCredentialsById(
					id.toString(),
					subscriberRegistrationVO.getAuthCredentials().signingKey(),
					subscriberRegistrationVO.getAuthCredentials().certificateChain());
		} catch (FolderCreationException | FileCreationException e) {
			try {
				// we explicitly delete, in case it was only paritally created
				iShareCredentialsRepository.deleteCredentialsById(id.toString());
			} catch (DeletionException deletionException) {
				log.warn("Rollback deletion failed, we will bubble the original exception and log the deletion to debug.");
				log.debug("Deletion exception: ", deletionException);
			}
			// bubble the exception to allow the db to rollback.
			throw e;
		}
	}

	@Override
	public void deleteEndpoint(UUID id) {
		Optional<Endpoint> optionalEndpoint = endpointRepository.findById(id);
		if (optionalEndpoint.isPresent()) {
			iShareCredentialsRepository
					.deleteCredentialsById(
							optionalEndpoint.get().getId().toString());
		}
	}

	@Override
	public void updateEndpointCredential(UUID id, String credentialType, String credentialBody) throws CredentialsConfigNotFound {

		Endpoint endpoint = endpointRepository
				.findById(id).orElseThrow(() -> new IllegalArgumentException(String.format("Endpoint %s not found.", id)));

		switch (IShareAuthCredentialType.getForCredentialType(credentialType)) {
			case CERT_CHAIN -> iShareCredentialsRepository.updateCertificateChainById(endpoint.getId().toString(), credentialBody);
			case KEY -> iShareCredentialsRepository.updateSigningKeyById(endpoint.getId().toString(), credentialBody);
			default -> throw new CredentialsConfigNotFound(
					"IShare does not support updating the requested credentials.",
					credentialType,
					Arrays.stream(IShareAuthCredentialType.values())
							.map(IShareAuthCredentialType::getCredentialType)
							.toList());
		}
	}
}
