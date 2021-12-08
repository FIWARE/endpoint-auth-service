package org.fiware.sidecar.persistence;

import io.micronaut.data.repository.CrudRepository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * Repository interface for managing Endpoints
 */
public interface EndpointRepository extends CrudRepository<Endpoint, UUID> {

	/**
	 * Get an endpoint by its domain-path combination
	 * @param domain - domain to be retrieved
	 * @param path - path to be retrieved
	 * @return the optional endpoint
	 */
	Optional<Endpoint> findByDomainAndPath(String domain, String path);

	/**
	 * Return all endpoints configured for the domain
	 * @param domain - the domain to find
	 * @return the endpoints for the domain
	 */
	List<Endpoint> findByDomain(String domain);

}
