package org.fiware.sidecar.persistence;

import io.micronaut.data.jdbc.annotation.JdbcRepository;
import io.micronaut.data.model.query.builder.sql.Dialect;
import io.micronaut.data.repository.CrudRepository;

import java.util.Optional;
import java.util.UUID;

@JdbcRepository(dialect = Dialect.H2)
 public interface SubscriberRepository extends CrudRepository<Subscriber, UUID> {

	Optional<Subscriber> findByDomainAndPath(String domain, String path);
}
