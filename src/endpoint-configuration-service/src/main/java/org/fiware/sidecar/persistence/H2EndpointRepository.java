package org.fiware.sidecar.persistence;

import io.micronaut.context.annotation.Requires;
import io.micronaut.data.jdbc.annotation.JdbcRepository;
import io.micronaut.data.model.query.builder.sql.Dialect;

/**
 * Endpoint repository to use H2
 */
@Requires(property = "datasources.default.dialect", value = "H2")
@JdbcRepository(dialect = Dialect.H2)
public interface H2EndpointRepository extends EndpointRepository {
}
