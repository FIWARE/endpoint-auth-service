package org.fiware.sidecar.persistence;

import lombok.Data;
import org.fiware.sidecar.model.AuthType;

import javax.persistence.Entity;
import javax.persistence.GeneratedValue;
import javax.persistence.Id;
import java.util.UUID;

/**
 * DB entity for an endpoint
 */
@Entity
@Data
public class Endpoint {

	@Id
	@GeneratedValue
	private UUID id;


	private String domain;
	private String path;
	private int port;
	private int targetPort;
	private boolean useHttps;
	private AuthType authType;
	private String iShareClientId;
	private String iShareIdpId;
	private String iShareIdpAddress;
	private String requestGrantType;
}
