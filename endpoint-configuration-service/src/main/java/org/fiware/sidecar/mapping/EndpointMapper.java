package org.fiware.sidecar.mapping;

import org.fiware.sidecar.model.AuthType;
import org.fiware.sidecar.model.AuthTypeVO;
import org.fiware.sidecar.model.EndpointInfoVO;
import org.fiware.sidecar.model.EndpointRegistrationVO;
import org.fiware.sidecar.model.MustacheEndpoint;
import org.fiware.sidecar.persistence.Endpoint;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.Mappings;
import org.mapstruct.Named;

import java.util.UUID;

@Mapper(componentModel = "jsr330")
public interface EndpointMapper {

	EndpointInfoVO endpointToEndpointInfoVo(Endpoint subscriber);

	@Mappings({
			@Mapping(source = "authCredentials.iShareClientId", target = "IShareClientId"),
			@Mapping(source = "authCredentials.iShareIdpId", target = "IShareIdpId"),
			@Mapping(source = "authCredentials.iShareIdpAddress", target = "IShareIdpAddress"),
			@Mapping(source = "authCredentials.requestGrantType", target = "requestGrantType"),
	})
	Endpoint endpointRegistrationVoToEndpoint(EndpointRegistrationVO endpointRegistrationVO);

	AuthType authTypeVoToAuthType(AuthTypeVO authTypeVO);
	@Mappings({
			@Mapping(source = "useHttps", target = "httpsPort", qualifiedByName = "useHttpsMustacheMapping")
	})
	MustacheEndpoint endpointToMustacheEndpoint(Endpoint endpoint);

	@Named("useHttpsMustacheMapping")
	static String useHttpsMustacheMapping(boolean useHttps) {
		return useHttps ? "https" : null;
	}

	default UUID StringToUUID(String value) {
		return UUID.fromString(value);
	}

	default String map(UUID value) {
		return value.toString();
	}
}
