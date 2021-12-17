package org.fiware.sidecar;

import io.micronaut.runtime.Micronaut;

/**
 * Base application as starting point
 */
public class Application {

	public static void main(String[] args) {
		Micronaut.run(Application.class, args);
	}

}
