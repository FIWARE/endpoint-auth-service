package org.fiware.sidecar.persistence;

import io.micronaut.data.repository.CrudRepository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

public interface EndpointRepository extends CrudRepository<Endpoint, UUID> {

	Optional<Endpoint> findByDomainAndPath(String domain, String path);

	List<Endpoint> findByDomain(String domain);

}
