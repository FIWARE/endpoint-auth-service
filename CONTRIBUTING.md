# Contributing to the Endpoint-Auth-Service

Thanks for checking out the Endpoint-Auth-Service. In order to contribute, please check the general [FIWARE development guidelines](https://fiware-requirements.readthedocs.io/en/latest/lifecycle/index.html).

## Coding guidelines

The Endpoint-Auth-Service consists of multiple components, using different technologies and languages. Please check the individual guidelines for them.
If a new one is added in the contribution, please add a rationale for that(f.e. in the form of an [ADR](../README.md#ADRs)).
For component specific information, check their folders. 

### Java

Your contributions should try to follow the [google java coding guidelines](https://google.github.io/styleguide/javaguide.html). The structure of your
code should fit the principles of [Domain Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html) and use the DI-mechanisms of
[Mirconaut](https://docs.micronaut.io/3.1.3/guide/index.html). Be aware of the framework and make use of its functionalities wherever it makes sense.
Additional tooling for code-generation is included in the project([lombok](https://projectlombok.org/), [openAPI-codegen](https://github.com/kokuwaio/micronaut-openapi-codegen),
[mapstruct](https://mapstruct.org/)) in order to reduce boiler-plate code.

### Golang

Your contributions should try to follow the idioms and recommendations from [effective_go](https://go.dev/doc/effective_go). [Gin](https://github.com/gin-gonic/gin)
is used as a web-framework. Be aware of the framework and make use of its functionalities wherever it makes sense.

### Cucumber

For the overall [integration-testing](../integration-test), [cucumber](https://cucumber.io/) is used together with [junit5](https://junit.org/junit5/docs/current/user-guide/).

## Testing



## Pull Request

Since this project uses automatic versioning, please apply one of the following labels to your pull request:
* patch - the PR contains a fix
* minor - the PR contains a new feature/improvement
* major - the PR contains a breaking change

The PRs enforce squash merge. Please provide a proper description on your squash, it will be used for release notes.

## Vulnerabilities

Please report vulnerabilities as [bugs](#bug) or email the authors.

## Bugs & Enhancements

If you find bug or searching for a new feature, please check the [issues](https://github.com/wistefan/endpoint-auth-service/issues) and [pull requests](https://github.com/wistefan/endpoint-auth-service/pulls)
first.

### Bug

If your bug is not already mentioned, please create either a [PR](#pull-request) or a new issue. The issue should contain a brief description on the
observed behaviour, your expectation and a description on how to reproduce it (bonus points if you provide a testcase for reproduction;).

### Enhancement

Create an issue including a proper description for the new feature. You can also start with a PR right away, but it would be easier to align on the details
before and  save unnessary work if discussed before.
If new functionality is added, the [integration-testsuite](../integration-test) needs to be extended.