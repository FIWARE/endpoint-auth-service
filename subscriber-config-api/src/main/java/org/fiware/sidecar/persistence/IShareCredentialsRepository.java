package org.fiware.sidecar.persistence;

public interface IShareCredentialsRepository {

	void saveCredentialsByDomainAndPath(String domain, String path, String key, String certChain);
	void deleteCredentialsByDomainAndPath(String domain, String path);
	void updateSigningKeyByDomainAndPath(String domain, String path, String key);
	void updateCertificateChainByDomainAndPath(String domain, String path, String certChain);
}
