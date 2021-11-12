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
import java.util.stream.Collectors;

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
	public void createEndpoint(EndpointRegistrationVO subscriberRegistrationVO) {
		try {
			iShareCredentialsRepository.saveCredentialsByDomainAndPath(
					subscriberRegistrationVO.getDomain(),
					subscriberRegistrationVO.getPath(),
					subscriberRegistrationVO.getAuthCredentials().signingKey(),
					subscriberRegistrationVO.getAuthCredentials().certificateChain());
		} catch (FolderCreationException | FileCreationException e) {
			try {
				iShareCredentialsRepository.deleteCredentialsByDomainAndPath(subscriberRegistrationVO.getDomain(), subscriberRegistrationVO.getPath());
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
					.deleteCredentialsByDomainAndPath(
							optionalEndpoint.get().getDomain(),
							optionalEndpoint.get().getPath());
		}
	}

	@Override
	public void updateEndpointCredential(UUID id, String credentialType, String credentialBody) throws CredentialsConfigNotFound {

		Endpoint endpoint = endpointRepository
				.findById(id).orElseThrow(() -> new IllegalArgumentException(String.format("Endpoint %s not found.", id)));

		switch (IShareAuthCredentialType.getForCredentialType(credentialType)) {
			case CERT_CHAIN -> iShareCredentialsRepository.updateCertificateChainByDomainAndPath(endpoint.getDomain(), endpoint.getPath(), credentialBody);
			case KEY -> iShareCredentialsRepository.updateSigningKeyByDomainAndPath(endpoint.getDomain(), endpoint.getPath(), credentialBody);
			default -> throw new CredentialsConfigNotFound(
					"IShare does not support updating the requested credentials.",
					credentialType,
					Arrays.stream(IShareAuthCredentialType.values())
							.map(IShareAuthCredentialType::getCredentialType)
							.collect(Collectors.toList()));
		}
	}
}
