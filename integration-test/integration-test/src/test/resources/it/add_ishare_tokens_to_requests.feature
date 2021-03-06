Feature: Add iShare tokens to requests.
  iShare token's should be added to every configured outgoing request.

  Background: The data-provider uses a sidecar-proxy for adding authorization headers.
    Given The Data-provider is running with the endpoint-authentication-service as a sidecar-proxy.

  Scenario: Everything can pass-through untouched.
    Given No endpoint is configured.
    When Data-Provider sends a request to the data-consumer's root path.
    And Data-Provider sends a request to a sub-path of the data-consumer.
    And Data-Provider sends a request to the data-consumer-2's root path.
    And Data-Provider sends a request to a sub-path of the data-consumer-2.
    Then Data-Consumer should receive requests without an authorization-header.
    And Data-Consumer-2 should receive requests without an authorization-header.

  Scenario: Data-Consumer receives only authorized requests.
    Given Data-Consumer's root path is configured as an iShare endpoint.
    When Data-Provider sends a request to the data-consumer's root path.
    Then Data-Consumer should receive a request with an authorization-header.

  Scenario: Data-Consumer receives authorized requests at every sub-path.
    Given Data-Consumer's root path is configured as an iShare endpoint.
    When Data-Provider sends a request to a sub-path of the data-consumer.
    Then Data-Consumer should receive a request with an authorization-header.

  Scenario: Data-Consumer receive authorized requests at a sub-path.
    Given Data-Consumer subpath is configured as an iShare endpoint.
    When Data-Provider sends a request to the data-consumer's root path.
    Then Data-Consumer should receive a request without an authorization-header.

  Scenario: Data-Consumer receive authorized requests at a sub-path with multiple configured paths.
    Given Data-Consumer subpath is configured as an iShare endpoint.
    And Data-Consumer anotherpath is configured as an iShare endpoint.
    When Data-Provider sends a request to the data-consumer's root path.
    Then Data-Consumer should receive a request without an authorization-header.

  Scenario: Data-Consumer receives authorized requests only at a sub-path, root stays untouched.
    Given Data-Consumer subpath is configured as an iShare endpoint.
    When Data-Provider sends a request to a sub-path of the data-consumer.
    Then Data-Consumer should receive a request with an authorization-header.

  Scenario: Data-Consumer receives authorized requests only at multiple sub-path, root stays untouched.
    Given Data-Consumer subpath is configured as an iShare endpoint.
    And Data-Consumer anotherpath is configured as an iShare endpoint.
    When Data-Provider sends a request to a sub-path of the data-consumer.
    Then Data-Consumer should receive a request with an authorization-header.