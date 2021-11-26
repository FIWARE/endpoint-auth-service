package org.fiware.sidecar.rest;

import io.micronaut.core.util.AntPathMatcher;

public class PathPartMatcher extends AntPathMatcher {
	public boolean matchesPartly(String pattern, String source) {
		return doMatch(pattern, source, false);
	}
}
