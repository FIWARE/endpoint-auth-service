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
var config configuration

// Default configurations

/**
* Default plugin configuration.
* The defaults targeting a plain envoy sidecar "ishare"-usecase and WILL NOT work in a mesh setup(istio, ossm)
 */
var defaultPluginConfig pluginConfiguration = pluginConfiguration{authType: "ISHARE", authProviderName: "ext-authz", authRequestTimeout: 5000}

/**
* Default overall config. Domain and path will be empty, thus the filter will not be applied to any request
 */
var defaultConfig configuration = configuration{pluginConfig: defaultPluginConfig, domainConfig: domainConfig{}, pathConfig: pathConfig{}}

/**
* Json parser for reading cache and config
 */
var parser fastjson.Parser

/**
* Full configuration for the filter
 */
type configuration struct {
	pluginConfig pluginConfiguration `json:"general"`
	domainConfig domainConfig        `json:"domains"`
	pathConfig   pathConfig          `json:"paths"`
}

/**
* Struct to hold the config for this plugin.
 */
type pluginConfiguration struct {
	authType           string `json:"authType"`
	authProviderName   string `json:"authProviderName"`
	authRequestTimeout uint32 `json:"authRequestTimeout"`
}

/**
* Array with the paths that should be handled by the plugin.
 */
type pathConfig []string

/**
* Array to hold the domains to be handled by the plugin.
 */
type domainConfig []string

/**
* Struct to represent a chache entry containing the auth information
 */
type cachedAuthInformation struct {
	expirationTime int64       `json:"expiration"`
	cachedHeaders  headersList `json:"cachedHeaders"`
}

/**
* Struct to represent a single header as defined by the auth-provider api
 */
type header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

/**
* Struct containing headers to be returned by the auth provider.
 */
type headersList []header

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
	readConfiguration()
	proxywasm.LogInfo("Successfully started plugin.")
	return types.OnPluginStartStatusOK
}

// Update the plugin context and read the config and override types.DefaultPluginContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	readConfiguration()
	return &pluginContext{}
}

// Override types.DefaultPluginContext.
func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{}
}

/**
* Reads the auth-type from the plugin config
 */
func readConfiguration() {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil {
		proxywasm.LogCriticalf("Error reading plugin configuration: %v. Using the default.", err)
		config = defaultConfig
		return
	}

	proxywasm.LogInfof("Config: %v", string(data))

	config = parseConfigFromJson(string(data))
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

	//matchPath(path)

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
func addCachedHeadersToRequest(cachedHeaders headersList) {
	for _, header := range cachedHeaders {
		proxywasm.LogDebugf("Add header ", fmt.Sprint(header))
		proxywasm.AddHttpRequestHeader(header.Name, header.Value)
	}
}

/**
* Request auth info at the provider. Since the call is executed asynchronous, it needs to pause the actual request handling.
 */
func requestAuthProvider() types.Action {

	proxywasm.LogCriticalf("Call to %s", config.pluginConfig.authProviderName)
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
	hs[pathIndex] = [2]string{":path", "/" + config.pluginConfig.authType + "/auth?domain=" + domain + "&path=" + path}

	if _, err := proxywasm.DispatchHttpCall(config.pluginConfig.authProviderName, hs, nil, nil, config.pluginConfig.authRequestTimeout, authCallback); err != nil {
		proxywasm.LogCriticalf("Domain " + domain + " , path: " + path + " , authType: " + config.pluginConfig.authType)
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

	if body == nil {
		proxywasm.LogCriticalf("Failed to get response body for auth-request, body was nil.")
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
func cachedAuthInfoToJson(expirationTime int64, cachedHeaders headersList) (jsonString string) {

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
	return
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

func parseConfigFromJson(jsonString string) (config configuration) {

	config = defaultConfig
	parsedJson, err := parser.Parse(jsonString)

	if err != nil {
		proxywasm.LogCriticalf("Unable to parse config: %v, will use default", err)
		return
	}

	generalConfigJson := parsedJson.GetStringBytes("general")
	domainsConfigJson := parsedJson.GetArray("domains")
	pathsConfigJson := parsedJson.GetArray("paths")
	if generalConfigJson != nil {
		config.pluginConfig = parsePluginConfigFromJson(string(generalConfigJson))
	}

	if domainsConfigJson != nil {
		config.domainConfig = parseDomainConfigFromJson(domainsConfigJson)
	}
	if pathsConfigJson != nil {
		config.pathConfig = parsePathConfigFromJson(pathsConfigJson)
	}
	return
}

// The following two methods contain a lot of code duplication. Thats due to the fact that tinygo does not support
// generics(yet), thus this approach is the most readable one.

/**
* Parse the json array containing the domains to be handled and return them as a config object.
 */
func parseDomainConfigFromJson(valuesArray []*fastjson.Value) (parsedConfig domainConfig) {
	parsedConfig = domainConfig{}

	if valuesArray == nil {
		proxywasm.LogWarnf("Did not receive any domain config. Return empty array.")
		return
	}

	for _, entry := range valuesArray {
		parsedConfig = append(parsedConfig, string(entry.GetStringBytes()))
	}
	return
}

/**
* Parse the json array containing the paths to be handled and return them as a config object.
 */
func parsePathConfigFromJson(valuesArray []*fastjson.Value) (parsedConfig pathConfig) {
	parsedConfig = pathConfig{}

	if valuesArray == nil {
		proxywasm.LogWarnf("Did not receive any path config. Return empty array.")
		return
	}

	for _, entry := range valuesArray {
		parsedConfig = append(parsedConfig, string(entry.GetStringBytes()))
	}
	return
}

/**
* Parse the jsonstring, containing the configuration
 */
func parsePluginConfigFromJson(jsonString string) (parsedConfig pluginConfiguration) {

	parsedConfig = defaultPluginConfig
	parsedJson, err := parser.Parse(jsonString)

	if err != nil {
		proxywasm.LogCriticalf("Unable to parse config: %v, will use default", err)
		return
	}

	authRequestTimeout := parsedJson.GetInt("authRequestTimeout")
	authType := parsedJson.GetStringBytes("authType")
	authProviderName := parsedJson.GetStringBytes("authProviderName")

	if authRequestTimeout > 0 {
		parsedConfig.authRequestTimeout = uint32(authRequestTimeout)
	} else {
		proxywasm.LogWarnf("Use default requestTimeout: %v", defaultPluginConfig.authRequestTimeout)
	}

	if authType != nil {
		parsedConfig.authType = string(authType)
	} else {
		proxywasm.LogWarnf("Use default authType: %v", defaultPluginConfig.authType)
	}

	if authProviderName != nil {
		parsedConfig.authProviderName = string(authProviderName)
	} else {
		proxywasm.LogWarnf("Use default authProvider: %v", defaultPluginConfig.authProviderName)
	}

	return
}

/**
* Parse the cached json to an auth-info object
 */
func parseCachedAuthInformation(jsonString string) (cachedAuthInfo cachedAuthInformation, err error) {

	parsedJson, err := parser.Parse(jsonString)

	expirationTime := parsedJson.GetInt64("expiration")
	cachedHeaders := parsedJson.GetArray("cachedHeaders")
	headersList := parseHeaderArray(cachedHeaders)

	return cachedAuthInformation{expirationTime: expirationTime, cachedHeaders: headersList}, err

}

/**
* Helper method to parse a json-array to the headers list
 */
func parseHeaderArray(valuesArray []*fastjson.Value) (headerList headersList) {

	if valuesArray == nil {
		proxywasm.LogWarnf("Not headers to parse. Return empty array.")
		return headersList{}
	}

	for _, entry := range valuesArray {
		var name string = string(entry.GetStringBytes("name"))
		var value string = string(entry.GetStringBytes("value"))
		proxywasm.LogDebugf("Header entry is %s : %s", name, value)
		headerList = append(headerList, header{name, value})
	}
	return
}

/**
* Helper method to parse a json-string to the headers list
 */
func parseHeaderList(jsonString string) (headerList headersList, err error) {

	parsedJson, err := parser.Parse(jsonString)
	if err != nil {
		proxywasm.LogCriticalf("Was not able to parse string %s", jsonString)
		return
	}

	jsonArray, err := parsedJson.Array()
	if err != nil {
		proxywasm.LogCriticalf("Was not able to parse json to array %s", jsonString)
		return
	}

	return parseHeaderArray(jsonArray), err

}
