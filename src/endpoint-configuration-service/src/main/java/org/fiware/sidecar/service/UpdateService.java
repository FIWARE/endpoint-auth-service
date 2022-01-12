package org.fiware.sidecar.service;

/**
 * Interface for updating configurations to be used by different sidecar solutions(f.e. envoy, service-mesh)
 */
public interface UpdateService {
	/**
	 * Schedule the actual update. It's not directly triggered, to allow solution-specific update frequencies
	 */
	void scheduleConfigUpdate();
}
