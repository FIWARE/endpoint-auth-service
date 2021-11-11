package org.fiware.sidecar.persistence;

import io.micronaut.data.annotation.Repository;
import io.micronaut.data.repository.CrudRepository;

@Repository
public interface SubscriberRepository extends CrudRepository<Subscriber, String> {

	Subscriber findByDomainAndPath(String domain, String path);
}
