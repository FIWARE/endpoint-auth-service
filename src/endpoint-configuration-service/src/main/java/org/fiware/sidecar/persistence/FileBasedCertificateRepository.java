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
	private static final String FOLDER_PATH_TEMPLATE = "%s/%s";

	private final IShareProperties iShareProperties;


	@Override
	public void saveCredentialsById(String id, String key, String certChain) {
		Path folderPath = buildFolderPath(id);
		if (Files.exists(buildFilePath(folderPath.toString(), IShareAuthCredentialType.KEY.getFileName())) || Files.exists(buildFilePath(folderPath.toString(), IShareAuthCredentialType.CERT_CHAIN.getFileName()))) {
			throw new IllegalArgumentException(String.format("Credentials for %s already exists.", id));
		}
		if (!Files.exists(folderPath)) {
			try {
				Files.createDirectories(folderPath);
			} catch (IOException e) {
				throw new FolderCreationException("Was not able to create the requested folder.", e, folderPath.toString());
			}
		}
		storeStringAt(IShareAuthCredentialType.KEY, id, key);
		storeStringAt(IShareAuthCredentialType.CERT_CHAIN, id, certChain);
	}

	@Override
	public void deleteCredentialsById(String id) {
		Path folderPath = buildFolderPath(id);
		deleteFileAt(IShareAuthCredentialType.CERT_CHAIN, folderPath);
		deleteFileAt(IShareAuthCredentialType.KEY, folderPath);
	}

	@Override
	public void updateSigningKeyById(String id, String key) {
		try {
			storeStringAt(IShareAuthCredentialType.KEY, id, key);
		} catch (RuntimeException e) {
			throw new FileUpdateException("Was not able to update file.", e);
		}
	}

	@Override
	public void updateCertificateChainById(String id, String certChain) {
		try {
			storeStringAt(IShareAuthCredentialType.CERT_CHAIN, id, certChain);
		} catch (RuntimeException e) {
			throw new FileUpdateException("Was not able to update file.", e);
		}
	}

	private void storeStringAt(IShareAuthCredentialType fileType, String id, String toStore) {
		Path filePath = buildFilePath(buildFolderPath(id).toString(), fileType.getFileName());
		try {
			Files.writeString(filePath, toStore);
		} catch (IOException e) {
			throw new FileCreationException("Was not able to create the requested file.", e, filePath.toString());
		}
	}

	private void deleteFileAt(IShareAuthCredentialType fileType, Path folderPath) {
		try {
			Files.delete(buildFilePath(folderPath.toString(), fileType.getFileName()));
		} catch (IOException e) {
			throw new DeletionException(String.format("Was not able to delete %s.", folderPath), e, folderPath.toString());
		}
	}

	private Path buildFolderPath(String id) {
		return Path.of(
				String.format(FOLDER_PATH_TEMPLATE,
						stripTrailingSlashes(iShareProperties.getCertificateFolderPath()),
						id));
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
