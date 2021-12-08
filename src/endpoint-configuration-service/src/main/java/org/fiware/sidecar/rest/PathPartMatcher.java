package org.fiware.sidecar.rest;

import io.micronaut.core.util.AntPathMatcher;

/**
 * Extension of the {@link  AntPathMatcher} to make its partly-match funtionallity accessible
 */
public class PathPartMatcher extends AntPathMatcher {
	public boolean matchesPartly(String pattern, String source) {
		return doMatch(pattern, source, false);
	}
}
