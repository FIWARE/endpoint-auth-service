package org.fiware.sidecar.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.fiware.sidecar.exception.CredentialsConfigNotFound;
import org.fiware.sidecar.exception.DeletionException;
import org.fiware.sidecar.exception.FileCreationException;
import org.fiware.sidecar.exception.FolderCreationException;
import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.SubscriberRegistrationVO;
import org.fiware.sidecar.model.ishare.IShareAuthCredentialType;
import org.fiware.sidecar.persistence.IShareCredentialsRepository;
import org.fiware.sidecar.persistence.Subscriber;
import org.fiware.sidecar.persistence.SubscriberRepository;

import javax.inject.Singleton;
import java.util.Arrays;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@Slf4j
@RequiredArgsConstructor
@Singleton
public class IShareSubscriberWriteService implements SubscriberWriteService {

	private final IShareCredentialsRepository iShareCredentialsRepository;
	private final SubscriberRepository subscriberRepository;

	@Override
	public AuthType supportedAuthType() {
		return AuthType.ISHARE;
	}

	@Override
	public void createSubscriber(SubscriberRegistrationVO subscriberRegistrationVO) {
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
	public void deleteSubscriber(UUID id) {
		Optional<Subscriber> optionalSubscriber = subscriberRepository.findById(id);
		if (optionalSubscriber.isPresent()) {
			iShareCredentialsRepository
					.deleteCredentialsByDomainAndPath(
							optionalSubscriber.get().getDomain(),
							optionalSubscriber.get().getPath());
		}
	}

	@Override
	public void updateSubscriberCredential(UUID id, String credentialType, String credentialBody) throws CredentialsConfigNotFound {

		Subscriber subscriber = subscriberRepository
				.findById(id).orElseThrow(() -> new IllegalArgumentException(String.format("Subscriber %s not found.", id)));

		switch (IShareAuthCredentialType.getForCredentialType(credentialType)) {
			case CERT_CHAIN -> iShareCredentialsRepository.updateCertificateChainByDomainAndPath(subscriber.getDomain(), subscriber.getPath(), credentialBody);
			case KEY -> iShareCredentialsRepository.updateSigningKeyByDomainAndPath(subscriber.getDomain(), subscriber.getPath(), credentialBody);
			default -> throw new CredentialsConfigNotFound(
					"IShare does not support updating the requested credentials.",
					credentialType,
					Arrays.stream(IShareAuthCredentialType.values())
							.map(IShareAuthCredentialType::getCredentialType)
							.collect(Collectors.toList()));
		}
	}
}
