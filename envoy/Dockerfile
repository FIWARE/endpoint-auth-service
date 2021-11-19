FROM ubuntu:20.04 as build

RUN apt-get update
RUN apt-get install -y wget
RUN wget https://raw.githubusercontent.com/rxi/json.lua/v0.1.2/json.lua --directory-prefix /tmp/

FROM envoyproxy/envoy:v1.20.0
RUN mkdir /lua-json
RUN chmod a+rw /lua-json
COPY --from=build /tmp/json.lua /lua-json/json.lua
RUN chmod a+rx /lua-json/json.lua
