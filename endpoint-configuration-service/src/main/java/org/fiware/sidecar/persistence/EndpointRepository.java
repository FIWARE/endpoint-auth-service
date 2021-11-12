package org.fiware.sidecar.persistence;

import io.micronaut.data.jdbc.annotation.JdbcRepository;
import io.micronaut.data.model.query.builder.sql.Dialect;
import io.micronaut.data.repository.CrudRepository;

import java.util.Optional;
import java.util.UUID;

@JdbcRepository(dialect = Dialect.H2)
 public interface EndpointRepository extends CrudRepository<Endpoint, UUID> {

	Optional<Endpoint> findByDomainAndPath(String domain, String path);
}
