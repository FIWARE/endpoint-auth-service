#Feature: Handle multiple tokens
#  iShare token's should only added to configured consumers.
#
#  Background: The data-provider uses a sidecar-proxy for adding authorization headers.
#    Given The Data-provider is running with the endpoint-authentication-service as a sidecar-proxy.
#
#  Scenario: Only Data-Consumer receives authorized requests at root, when nothing else is configured.
#    Given Data-Consumer's root path is configured as an iShare endpoint.
#    And No other endpoint is configured.
#    When Data-Provider sends a request to the data-consumer's root path.
#    And Data-Provider sends a request to a sub-path of the data-consumer.
#    And Data-Provider sends a request to the data-consumer-2's root path.
#    And Data-Provider sends a request to a sub-path of the data-consumer-2.
#    Then Data-Consumer should receive requests with an authorization-header.
#    And Data-Consumer-2 should receive requests without an authorization-header.
#
#  Scenario: Only Data-Consumer receives authorized requests at subpath, when nothing else is configured.
#    Given Data-Consumer subpath is configured as an iShare endpoint.
#    And No other endpoint is configured.
#    When Data-Provider sends a request to a sub-path of the data-consumer.
#    When Data-Provider sends a request to the data-consumer's root path.
#    And Data-Provider sends a request to the data-consumer-2's root path.
#    And Data-Provider sends a request to a sub-path of the data-consumer-2.
#    Then Data-Consumer should receive requests with an authorization-header at the subpath.
#    And Data-Consumer-2 should receive requests without an authorization-header.
#    And Data-Consumer should receive requests without an authorization-header at root path.