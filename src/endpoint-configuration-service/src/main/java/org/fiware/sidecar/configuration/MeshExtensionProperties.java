package org.fiware.sidecar.configuration;

import io.micronaut.context.annotation.ConfigurationBuilder;
import io.micronaut.context.annotation.ConfigurationProperties;
import lombok.Data;
import lombok.Getter;
import lombok.Setter;

import java.util.List;

@ConfigurationProperties("meshExtension")
@Data
public class MeshExtensionProperties {

	/**
	 * Should the generation of service-mesh extension configurations be enabled?
	 */
	private boolean enabled;

	/**
	 * Address to access the authProvider. Should be a service address available at the mesh.
	 * Format looks like: 'outbound|80||ext-authz' (e.g. _direction_|_port_||_service-name_)
	 */
	private String authProviderName;

	/**
	 * Selector used for applying the mesh-extension to a specific workload. It should be a label assigned to the target service.
	 */
	@ConfigurationBuilder(configurationPrefix = "workloadSelector")
	private MeshExtensionProperties.MetaData workloadSelector = new MeshExtensionProperties.MetaData();

	/**
	 * Version of the filter to be included in the meshExtension file.
	 */
	private String filterVersion;

	/**
	 * Name to be used for the mesh extension when deployed to k8s/openshift.
	 */
	private String extensionName;

	/**
	 * Namespace to deploy the mesh extension into.
	 */
	private String extensionNamespace;

	/**
	 * Path to which the extension file will be written.
	 */
	private String meshExtensionYamlPath;

	/**
	 * Annotations that should be applied to the mesh extension's metadata.
	 */
	private List<MetaData> annotations;

	/**
	 * Labels that should be applied to the mesh extension's metadata.
	 */
	private List<MetaData> labels;

	/**
	 * Pojo holding a k8s metadata entry like label or annotation
	 */
	@Getter
	@Setter
	public static class MetaData {

		private String name;
		private String value;
	}
}
