package org.fiware.sidecar.mapping;

import org.fiware.sidecar.model.SubscriberInfoVO;
import org.fiware.sidecar.model.SubscriberRegistrationVO;
import org.fiware.sidecar.persistence.Subscriber;
import org.mapstruct.Mapper;

import java.util.UUID;

@Mapper(componentModel = "jsr330")
public interface SubscriberMapper {

	SubscriberInfoVO subscriberToSubscriberInfoVo(Subscriber subscriber);
	Subscriber subscriberRegistrationVoToSubscriber(SubscriberRegistrationVO subscriberRegistrationVO);
	default UUID StringToUUID(String value){
		return UUID.fromString(value);
	}
}
