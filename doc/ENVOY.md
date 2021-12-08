# Envoy

[Envoy](https://www.envoyproxy.io) is used as the side-car to add auth-information for outgoing requests. 
See the ["Use envoy as proxy"-ADR](adr/choose-proxy.md) for the rational of that choice.

In order to support the required [lua-functionality](https://www.lua.org/), dedicated container for [envoy](https://quay.io/repository/wi_stefan/envoy) is provided.
Its a small extension of the official [envoy-docker](https://hub.docker.com/r/envoyproxy/envoy) with the [json-lua library](https://raw.githubusercontent.com/rxi/json.lua/v0.1.2/json.lua) 
already installed. Any other envoy can be used, as long as the library is provided.