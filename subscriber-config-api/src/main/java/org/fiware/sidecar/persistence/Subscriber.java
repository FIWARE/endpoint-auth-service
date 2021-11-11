package org.fiware.sidecar.persistence;

import lombok.Data;
import org.fiware.sidecar.model.AuthType;

import javax.persistence.Entity;
import javax.persistence.Id;

@Entity
@Data
public class Subscriber {

	@Id
	private String id;

	private String domain;
	private String path;
	private boolean useHttps;
	private AuthType authType;
	private String iShareClientId;
	private String iShareIdpId;
	private String iShareIdpAddress;
	private String requestGrantType;
}
