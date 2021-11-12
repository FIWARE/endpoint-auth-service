package org.fiware.sidecar.persistence;

public interface IShareCredentialsRepository {

	void saveCredentialsById(String id, String key, String certChain);
	void deleteCredentialsById(String id);
	void updateSigningKeyById(String id, String key);
	void updateCertificateChainById(String id, String certChain);
}
