Feature: Handle multiple tokens
  iShare token's should only added to configured consumers.

  Background: The data-provider uses a sidecar-proxy for adding authorization headers.
    Given The Data-provider is running with the endpoint-authentication-service as a sidecar-proxy.

  Scenario: Data-Consumer receives only authorized requests.
    Given Data-Consumer's root path is configured as an iShare endpoint.
    When Data-Provider sends a request to the data-consumer's root path.
    And Data-Provider sends a request to the data-consumer-2's root path.
    Then Data-Consumer should receive a request with an authorization-header.
    And Data-Consumer-2 should receive a request without an authorization-header.