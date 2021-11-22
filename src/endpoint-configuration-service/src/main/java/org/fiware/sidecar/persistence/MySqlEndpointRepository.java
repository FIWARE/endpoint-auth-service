package org.fiware.sidecar.persistence;

import io.micronaut.context.annotation.Requires;
import io.micronaut.data.jdbc.annotation.JdbcRepository;
import io.micronaut.data.model.query.builder.sql.Dialect;

@Requires(property = "datasources.default.dialect", value = "MySql")
@JdbcRepository(dialect = Dialect.MYSQL)
public interface MySqlEndpointRepository extends EndpointRepository {
}
