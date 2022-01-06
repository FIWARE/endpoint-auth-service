// Copyright 2020-2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/valyala/fastjson"
)

const (
	authorityKey = ":authority"
	pathKey      = ":path"
)

/**
* Global compare & set value for cache control
 */
var cas uint32 = 0
var domain string
var path string

/**
* Plugin configurations
 */
var config Configuration

// Default configurations

/**
* Default authtype to be used by the filter
 */
var defaultAuthType = "ISHARE"

/**
* Default (cluster)name to contact the auth-provider at
 */
var defaultAuthProviderName = "ext-authz"

/**
* Default timeout to be used when  requesting the auth provider.
 */
var defaultAuthRequestTimeout int64 = 5000

type Configuration struct {
	authType           string
	authProviderName   string
	authRequestTimeout int64
}

type CachedAuthInformation struct {
	expirationTime int64       `json:"expiration"`
	cachedHeaders  HeadersList `json:"cachedHeaders"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

/**
* Struct to contain headers to be returned by the auth provider.
 */
type HeadersList []Header

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type (
	vmContext     struct{}
	pluginContext struct {
		// Embed the default plugin context here,
		// so that we don't need to reimplement all the methods.
		types.DefaultPluginContext
	}

	httpContext struct {
		// Embed the default http context here,
		// so that we don't need to reimplement all the methods.
		types.DefaultHttpContext
	}
)

// Handle the start event of the wasm-vm
func (*vmContext) OnVMStart(vmConfigurationSize int) types.OnVMStartStatus {

	proxywasm.LogInfo("Successfully started VM.")
	return types.OnVMStartStatusOK
}

// Handle the plugin start and read the config
func (ctx pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	readAuthTypeFromPluginConfig()
	proxywasm.LogInfo("Successfully started plugin.")
	return types.OnPluginStartStatusOK
}

// Update the plugin context and read the config and override types.DefaultPluginContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	readAuthTypeFromPluginConfig()
	return &pluginContext{}
}

// Override types.DefaultPluginContext.
func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{}
}

/**
* Reads the auth-type from the plugin config
 */
func readAuthTypeFromPluginConfig() {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil {
		proxywasm.LogCriticalf("Error reading plugin configuration: %v", err)
	}

	proxywasm.LogCriticalf("Config: %v", string(data))

	config = parseConfigFromJson(string(data))

	proxywasm.LogCriticalf("Plugin configured: %s", config)
}

// Handle the actual request and retrieve the headers used for auth-handling
func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {

	// :authority header is set by envoy and holds the requested domain
	authorityHeader, err := proxywasm.GetHttpRequestHeader(authorityKey)
	if err != nil || authorityHeader == "" {
		proxywasm.LogCriticalf("Failed to get authority header: %v", err)
		return types.ActionContinue
	}
	// we are only interested in the host and want to ignore the port
	domain = strings.Split(authorityHeader, ":")[0]

	// :path header is set by envoy and holds the requested path
	pathHeader, err := proxywasm.GetHttpRequestHeader(pathKey)
	if err != nil || pathHeader == "" {
		proxywasm.LogCriticalf("Failed to get path header: %v", err)
		return types.ActionContinue
	}
	path = pathHeader

	return setHeader()
}

/**
* Apply the auth headers from either the cache or the auth provider
 */
func setHeader() types.Action {
	sharedDataKey := domain + path
	data, currentCas, err := proxywasm.GetSharedData(sharedDataKey)

	if err != nil || data == nil {
		return requestAuthProvider()
	}

	if data != nil {
		proxywasm.LogDebugf("Cache hit: ", string(data))
		cachedAuthInfo, err := parseCachedAuthInformation(string(data))
		if err != nil {
			proxywasm.LogCriticalf("Failed to parse cached info, request new instead. %v", err)
			cas = currentCas
			return requestAuthProvider()
		}

		proxywasm.LogDebugf("Expiry: %v, Current: %v", cachedAuthInfo.expirationTime, time.Now().Unix())
		if cachedAuthInfo.expirationTime <= time.Now().Unix() {
			proxywasm.LogDebugf("Cache expired. Request new auth.")
			cas = currentCas
			return requestAuthProvider()
		} else {
			proxywasm.LogDebugf("Cache still valid.")
			addCachedHeadersToRequest(cachedAuthInfo.cachedHeaders)
			return types.ActionContinue
		}

	}

	return types.ActionContinue
}

/**
* Apply the headers from the list to the current request
 */
func addCachedHeadersToRequest(cachedHeaders HeadersList) {
	for _, header := range cachedHeaders {
		proxywasm.LogDebugf("Add header ", fmt.Sprint(header))
		proxywasm.AddHttpRequestHeader(header.Name, header.Value)
	}
}

/**
* Request auth info at the provider. Since the call is executed asynchronous, it needs to pause the actual request handling.
 */
func requestAuthProvider() types.Action {

	proxywasm.LogCriticalf("Call to %s", config.authProviderName)
	hs, _ := proxywasm.GetHttpRequestHeaders()

	var methodIndex int
	var pathIndex int
	for i, h := range hs {
		if h[0] == ":method" {
			methodIndex = i
		}
		if h[0] == ":path" {
			pathIndex = i
		}
	}
	hs[methodIndex] = [2]string{":method", "GET"}
	hs[pathIndex] = [2]string{":path", "/" + config.authType + "/auth?domain=" + domain + "&path=" + path}

	if _, err := proxywasm.DispatchHttpCall(config.authProviderName, hs, nil, nil, config.authRequestTimeout, authCallback); err != nil {
		proxywasm.LogCriticalf("Domain " + domain + " , path: " + path + " , authType: " + config.authType)
		proxywasm.LogCriticalf("Call to auth-provider failed: %v", err)
		return types.ActionContinue
	}
	return types.ActionPause
}

/**
* Callback method to handle the authprovider response.
* It will resume the request handling before taking care of updating the cache, to reduce the latency of the request. This can lead to
* duplicate updates in edge-cases(e.g. many parallel requests), but wont lead to problems since the cache will only store the first of them, while the
* individual tokens are still valid.
 */
func authCallback(numHeaders, bodySize, numTrailers int) {
	body, err := proxywasm.GetHttpCallResponseBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("Failed to get response body for auth-request: %v", err)
		proxywasm.ResumeHttpRequest()
		return
	}
	headers, _ := proxywasm.GetHttpCallResponseHeaders()

	headersList, err := parseHeaderList(string(body))
	if err != nil {
		proxywasm.LogCriticalf("Was not able to decode header list.")
		proxywasm.ResumeHttpRequest()
		return
	}
	addCachedHeadersToRequest(headersList)
	// continue the request before handling the caching
	proxywasm.ResumeHttpRequest()

	proxywasm.LogDebugf("Handle caching.")

	// handle cachecontrol
	for _, h := range headers {
		proxywasm.LogDebugf("Parse headers: ", fmt.Sprint(h))
		if h[0] == "cache-control" {
			proxywasm.LogDebugf("Found cache-control header.")

			expiry, err := getCacheExpiry(h[1])
			if err != nil {
				proxywasm.LogCriticalf("Was not able to read cache control header. ", err)
				return
			}

			if expiry > 0 {
				proxywasm.LogDebugf("Expiry was set to: ", expiry)
				var parser fastjson.Parser
				parsedInfo, err := parser.Parse(cachedAuthInfoToJson(expiry, headersList))
				if err != nil {
					proxywasm.LogCriticalf("Was not able to parse auth info.", err)
					return
				}
				buffer := parsedInfo.Get().MarshalTo(nil)
				proxywasm.LogDebugf("Buffer is %v", string(buffer))
				proxywasm.SetSharedData(domain+path, buffer, cas)
			}
			proxywasm.LogDebugf("Cached auth info for %v / %v", domain, path)
			return
		}

	}

}

/**
* Create a json string to store in cache from the headers list
 */
func cachedAuthInfoToJson(expirationTime int64, cachedHeaders HeadersList) (jsonString string) {

	headerArray := `[`
	for i, header := range cachedHeaders {
		if i != 0 {
			headerArray = headerArray + `,`
		}
		headerArray = headerArray + `{"name":"` + header.Name + `","value":"` + header.Value + `"}`
	}
	headerArray = headerArray + `]`

	jsonString = fmt.Sprintf(`{"expiration":%d, "cachedHeaders":%s}`, expirationTime, headerArray)

	proxywasm.LogDebugf("Json string to store ", jsonString)
	return jsonString
}

/**
* Evaluate the cache-control header(e.g. max-age) to decide upon the cache-expiry. The control-header should never be ignored,
* since only the auth-provider knows about the expiry of the auth-info and therefore ignoring it will lead to invalid headers at the request.
 */
func getCacheExpiry(cacheControlHeader string) (expiry int64, err error) {
	directiveArray := strings.Split(cacheControlHeader, ",")
	for _, directive := range directiveArray {
		directiveArray := strings.Split(directive, "=")
		switch directiveArray[0] {
		case "no-cache":
			fallthrough
		case "no-store":
			fallthrough
		case "must-revalidate":
			return -1, err
		case "max-age":
			maxAge, err := strconv.Atoi(directiveArray[1])
			if err != nil {
				return -1, err
			}
			return time.Now().Unix() + int64(maxAge), err
		}
	}
	proxywasm.LogDebugf("Did not find any cache directive to be handled. ", cacheControlHeader)
	return -1, err
}

/**
* Parse the jsonstring, containing the configuration
 */
func parseConfigFromJson(jsonString string) (config Configuration) {
	var parser fastjson.Parser
	parsedJson, err := parser.Parse(jsonString)

	if err != nil {
		proxywasm.LogCriticalf("Unable to parse config: %v, will use default", err)
		config.authProviderName = defaultAuthProviderName
		config.authType = defaultAuthType
		config.authRequestTimeout = defaultAuthRequestTimeout
		return config
	}

	authRequestTimeout := parsedJson.GetInt64("authRequestTimeout")
	authType := parsedJson.GetStringBytes("authType")
	authProviderName := parsedJson.GetStringBytes("authProviderName")
	if authRequestTimeout > 0 {
		config.authRequestTimeout = authRequestTimeout
	} else {
		config.authRequestTimeout = defaultAuthRequestTimeout
	}

	if authType != nil {
		config.authType = string(authType)
	} else {
		config.authType = defaultAuthType
	}

	if authProviderName != nil {
		config.authProviderName = string(authProviderName)
	} else {
		config.authProviderName = defaultAuthProviderName
	}

	return config
}

/**
* Parse the cached json to an auth-info object
 */
func parseCachedAuthInformation(jsonString string) (cachedAuthInfo CachedAuthInformation, err error) {

	var parser fastjson.Parser
	parsedJson, err := parser.Parse(jsonString)

	expirationTime := parsedJson.GetInt64("expiration")
	cachedHeaders := parsedJson.GetArray("cachedHeaders")
	headersList, err := parseHeaderArray(cachedHeaders)
	if err != nil {
		return cachedAuthInfo, err
	}
	return CachedAuthInformation{expirationTime: expirationTime, cachedHeaders: headersList}, err

}

/**
* Helper method to parse a json-array to the headers list
 */
func parseHeaderArray(valuesArray []*fastjson.Value) (headerList HeadersList, err error) {
	if err != nil {
		return headerList, err
	}
	for _, entry := range valuesArray {
		var name string = string(entry.GetStringBytes("name"))
		var value string = string(entry.GetStringBytes("value"))
		proxywasm.LogDebugf("Header entry is %s : %s", name, value)
		headerList = append(headerList, Header{name, value})
	}
	return headerList, err
}

/**
* Helper method to parse a json-string to the headers list
 */
func parseHeaderList(jsonString string) (headerList HeadersList, err error) {

	var parser fastjson.Parser
	parsedJson, err := parser.Parse(jsonString)
	jsonArray, err := parsedJson.Array()
	return parseHeaderArray(jsonArray)

}
