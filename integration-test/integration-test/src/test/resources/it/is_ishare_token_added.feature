Feature: Is iShare token added?
  iShare token's should be added to every configured outgoing request.

  Scenario: Echo-Server is configured to receive authorized requests.
    Given Echo-server is configured as an iShare endpoint.
    When Client sends a request to the echo-server.
    Then Echo-server should receive a request with an authorization-header.