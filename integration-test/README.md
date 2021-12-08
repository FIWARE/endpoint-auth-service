# Integration tests

Since the endpoint-auth-service consists of multiple components working together, a dedicated integration-test suite is used. In order to provide readability,
[cucumber](https://cucumber.io) with the [gherkin-syntax](https://cucumber.io/docs/gherkin/) was choosen. The test-files can be found here:
- [gherkin-files](./integration-test/src/test/resources/it/)
The actual test-implementation is done with [junit-5](https://junit.org/junit5/docs/current/user-guide/) and the [okHttp-Library](https://github.com/square/okhttp).

## Run the tests

> Precondition: The tests require docker-compose and maven to be installed. See https://maven.apache.org/install.html and https://docs.docker.com/compose/install/ for instructions

The tests expect an already running system, configured according to the provided [docker-compose](../docker-compose/README.md). Please read the documentation of 
the compose-setup before starting. The tests make use of echo-servers as endpoints to mock the actual request-endpoints and to inspect the request. The code for the 
echo-server is here: [echo-server](echo-server).

To execute the test, run ```mvn clean verify```. The test will check if an endpoint-configuration-service is running in the background-step. Test-results will be printed out on the commandline aned
can be found after the test under ```/target/surefire-reports/```.

## Extend the tests

Every change should be tested via unit-test. Integration-tests are close to the top of the [testing-pyramid](https://www.browserstack.com/guide/testing-pyramid-for-test-automation) and therefor 
hard to write and expensive to run. 
However, if a new feature, especially one the includes multiple components, is added, an integration-test should be created.
* add a [cucumber.feature](https://cucumber.io/docs/gherkin/reference/) under [integration-test/src/test/resources/it](integration-test/src/test/resources/it) - detailed and expressive wording is very welcome
* try to reuse existing steps as much as possible(see [StepDefinitions](integration-test/src/test/java/it/StepDefinitions.java)), increases readability and decreases maintenance
* implement the additional steps(e.g. Given, When, Then) if needed (see [StepDefinitions](integration-test/src/test/java/it/StepDefinitions.java))

Execution in the test-pipeline will happen automatically.