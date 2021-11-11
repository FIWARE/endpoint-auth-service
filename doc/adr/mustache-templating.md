# Use mustache for config file templating

## Status

- proposed

## Context

To update the envoy configuration, a mechanism to generate the required yaml files is needed. The config will contain some static parts(f.e. the 
passthrough clusters or https handling) and the dynamic configuration set by the configuration-server. 

## Decision

The [listener-configuration](https://www.envoyproxy.io/docs/envoy/latest/start/quick-start/configuration-dynamic-filesystem#resources-listeners)
and the [cluster-configuration](https://www.envoyproxy.io/docs/envoy/latest/start/quick-start/configuration-dynamic-filesystem#resources-clusters) 
will be templated with [mustache](https://mustache.github.io/). The template files will live inside the 
configuration-server project and the rendered files will be served through shared volumes.

## Rational

- templating allows to manage the files in a readable and close to final-format fashion, in contrast to f.e. generating them completely from code
- static parts stay completely untouched and can easily be update in case of changes in the envoy config api
- mustache is an easy to use and understand logic-less templating language
- mustache is widely adopted and used in various projects
- easy integration into java