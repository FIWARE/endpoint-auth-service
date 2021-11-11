package org.fiware.sidecar;

import com.github.mustachejava.DefaultMustacheFactory;
import com.github.mustachejava.MustacheFactory;
import io.micronaut.context.annotation.Bean;
import io.micronaut.context.annotation.Factory;
import io.micronaut.runtime.Micronaut;

/**
 * Base application as starting point
 */
@Factory
public class Application {

	public static void main(String[] args) {
		Micronaut.run(Application.class, args);
	}

	@Bean
	public MustacheFactory mustacheFactory() {
		return new DefaultMustacheFactory();
	}

}
