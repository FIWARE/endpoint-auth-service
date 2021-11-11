package org.fiware.sidecar.persistence;

import lombok.RequiredArgsConstructor;
import org.fiware.sidecar.configuration.IShareProperties;
import org.fiware.sidecar.exception.DeletionException;
import org.fiware.sidecar.exception.FileCreationException;
import org.fiware.sidecar.exception.FileUpdateException;
import org.fiware.sidecar.exception.FolderCreationException;
import org.fiware.sidecar.model.ishare.IShareAuthCredentialType;

import javax.inject.Singleton;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

@Singleton
@RequiredArgsConstructor
public class FileBasedCertificateRepository implements IShareCredentialsRepository {

	private static final String FILE_PATH_TEMPLATE = "%s/%s";
	private static final String FOLDER_PATH = "%s/%s/%s";

	private final IShareProperties iShareProperties;

	@Override
	public void saveCredentialsByDomainAndPath(String domain, String path, String key, String certChain) {
		Path folderPath = buildFolderPath(domain, path);
		if (Files.exists(buildFilePath(folderPath.toString(), IShareAuthCredentialType.KEY.getFileName())) || Files.exists(buildFilePath(folderPath.toString(), IShareAuthCredentialType.CERT_CHAIN.getFileName()))) {
			throw new IllegalArgumentException(String.format("A certificate for %s/%s already exists.", domain, path));
		}
		if (!Files.exists(folderPath)) {
			try {
				Files.createDirectories(folderPath);
			} catch (IOException e) {
				throw new FolderCreationException("Was not able to create the requested folder.", e, folderPath.toString());
			}
		}
		storeStringAt(IShareAuthCredentialType.KEY, domain, path, key);
		storeStringAt(IShareAuthCredentialType.CERT_CHAIN, domain, path, certChain);
	}

	private void storeStringAt(IShareAuthCredentialType fileType, String domain, String path, String toStore) {
		Path filePath = buildFilePath(buildFolderPath(domain, path).toString(), fileType.getFileName());
		try {
			Files.writeString(filePath, toStore);
		} catch (IOException e) {
			throw new FileCreationException("Was not able to create the requested file.", e, filePath.toString());
		}
	}

	@Override
	public void deleteCredentialsByDomainAndPath(String domain, String path) {
		Path folderPath = buildFolderPath(domain, path);
		deleteFileAt(IShareAuthCredentialType.CERT_CHAIN, folderPath);
		deleteFileAt(IShareAuthCredentialType.KEY, folderPath);
	}

	@Override
	public void updateSigningKeyByDomainAndPath(String domain, String path, String key) {
		try {
			storeStringAt(IShareAuthCredentialType.KEY, domain, path, key);
		} catch (RuntimeException e) {
			throw new FileUpdateException("Was not able to update file.", e);
		}
	}

	@Override
	public void updateCertificateChainByDomainAndPath(String domain, String path, String certChain) {
		try {
			storeStringAt(IShareAuthCredentialType.CERT_CHAIN, domain, path, certChain);
		} catch (RuntimeException e) {
			throw new FileUpdateException("Was not able to update file.", e);
		}
	}

	private void deleteFileAt(IShareAuthCredentialType fileType, Path folderPath) {
		try {
			Files.delete(buildFilePath(folderPath.toString(), fileType.getFileName()));
		} catch (IOException e) {
			throw new DeletionException(String.format("Was not able to delete %s.", folderPath), e, folderPath.toString());
		}
	}

	private Path buildFolderPath(String domain, String path) {
		return Path.of(
				String.format(FOLDER_PATH,
						stripTrailingSlashes(iShareProperties.getCertificateFolderPath()),
						stripTrailingSlashes(domain),
						stripTrailingSlashes(path)));
	}

	private String stripTrailingSlashes(String inputString) {
		if (inputString.endsWith("/")) {
			return inputString.substring(0, inputString.length() - 1);
		}
		return inputString;
	}

	private Path buildFilePath(String folderPath, String filename) {
		return Path.of(String.format(FILE_PATH_TEMPLATE, stripTrailingSlashes(folderPath), filename));
	}
}
