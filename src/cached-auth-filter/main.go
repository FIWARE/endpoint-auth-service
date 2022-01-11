package main

import (
	"fmt"
	"path"
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
var requestDomain string
var requestPath string

/**
* Plugin configurations
 */
var config pluginConfiguration

var endpointAuthConfig endpointAuthConfiguration

// Default configurations

/**
* Default plugin configuration.
* The defaults targeting a plain envoy sidecar "ishare"-usecase and WILL NOT work in a mesh setup(istio, ossm)
 */
var defaultPluginConfig pluginConfiguration = pluginConfiguration{authProviderName: "ext-authz", authRequestTimeout: 5000, enableEndpointMatching: false, authType: "ISHARE"}

/**
* Json parser for reading cache and config
 */
var parser fastjson.Parser

/**
* Struct to hold the config for this plugin.
 */
type pluginConfiguration struct {
	authProviderName       string
	authRequestTimeout     uint32
	enableEndpointMatching bool
	authType               string
}

/**
* Tree like represenation of the endpoint-auth config
 */
type endpointAuthConfiguration map[string]map[string]string

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
	expirationTime int64
	cachedHeaders  headersList
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
		config = defaultPluginConfig
		return
	}

	proxywasm.LogInfof("Config: %v", string(data))

	parseConfigFromJson(string(data))

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
	requestDomain = strings.Split(authorityHeader, ":")[0]

	// :path header is set by envoy and holds the requested path
	pathHeader, err := proxywasm.GetHttpRequestHeader(pathKey)
	if err != nil || pathHeader == "" {
		proxywasm.LogCriticalf("Failed to get path header: %v", err)
		return types.ActionContinue
	}
	requestPath = pathHeader

	if config.enableEndpointMatching {
		proxywasm.LogDebug("Endpoint matching is enabled. Match the path")

		authType, match := matchEndpoint(requestDomain, requestPath)
		proxywasm.LogDebugf("Match result was %v - type %v", match, authType)
		if !match {
			// early exit, nothing to handle for the filter
			return types.ActionContinue
		}
		return setHeader(authType)
	} else {
		return setHeader(config.authType)
	}

}

func matchEndpoint(domainString, pathString string) (authType string, match bool) {

	proxywasm.LogDebugf("Match %s - %s.", domainString, pathString)

	if domainString == "" {
		// if no domain is provided, return immediatly.
		return
	}

	if pathString == "" {
		// if no path is provided, return immediatly.
		return
	}

	pathEntry, domainExists := endpointAuthConfig[domainString]
	if !domainExists {
		// domain not configured, return immediatly
		return
	}

	var matchLength int = 0

	for configuredPath, configuredAuthType := range pathEntry {

		match, err := path.Match(configuredPath, pathString)
		proxywasm.LogDebugf("Current match on %s - %s: %v", configuredPath, pathString, match)
		if err != nil {
			proxywasm.LogWarnf("Invalid path in configuration: %s", configuredPath)
			continue
		}
		if !match {
			//early exit, do not count length of path if no match
			continue
		}
		if cpLen := len(configuredPath); cpLen > matchLength {
			matchLength = cpLen
			authType = configuredAuthType
		}
	}

	// if something matches, the length is bigger than 0
	match = matchLength > 0

	return
}

/**
* Apply the auth headers from either the cache or the auth provider
 */
func setHeader(authType string) types.Action {
	sharedDataKey := requestDomain + requestPath
	data, currentCas, err := proxywasm.GetSharedData(sharedDataKey)

	if err != nil || data == nil {
		return requestAuthProvider(authType)
	}

	proxywasm.LogDebugf("Cache hit: %s", string(data))
	cachedAuthInfo, err := parseCachedAuthInformation(string(data))
	if err != nil {
		proxywasm.LogCriticalf("Failed to parse cached info, request new instead. %v", err)
		cas = currentCas
		return requestAuthProvider(authType)
	}

	proxywasm.LogDebugf("Expiry: %v, Current: %v", cachedAuthInfo.expirationTime, time.Now().Unix())
	if cachedAuthInfo.expirationTime <= time.Now().Unix() {
		proxywasm.LogDebugf("Cache expired. Request new auth.")
		cas = currentCas
		return requestAuthProvider(authType)
	} else {
		proxywasm.LogDebugf("Cache still valid.")
		addCachedHeadersToRequest(cachedAuthInfo.cachedHeaders)
		return types.ActionContinue
	}

}

/**
* Apply the headers from the list to the current request
 */
func addCachedHeadersToRequest(cachedHeaders headersList) {
	for _, header := range cachedHeaders {
		proxywasm.LogDebugf("Add header %s", fmt.Sprint(header))
		proxywasm.AddHttpRequestHeader(header.Name, header.Value)
	}
}

/**
* Request auth info at the provider. Since the call is executed asynchronous, it needs to pause the actual request handling.
 */
func requestAuthProvider(authType string) types.Action {

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
	hs[pathIndex] = [2]string{":path", "/" + authType + "/auth?domain=" + requestDomain + "&path=" + requestPath}

	if _, err := proxywasm.DispatchHttpCall(config.authProviderName, hs, nil, nil, config.authRequestTimeout, authCallback); err != nil {
		proxywasm.LogCriticalf("Domain " + requestDomain + " , path: " + requestPath + " , authType: " + authType)
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

	proxywasm.LogDebug("Auth callback")

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
		proxywasm.LogDebugf("Parse headers: %s", fmt.Sprint(h))
		if h[0] == "cache-control" {
			proxywasm.LogDebugf("Found cache-control header.")

			expiry, err := getCacheExpiry(h[1])
			if err != nil {
				proxywasm.LogCriticalf("Was not able to read cache control header. %v", err)
				return
			}

			if expiry > 0 {
				proxywasm.LogDebugf("Expiry was set to: %v", expiry)
				parsedInfo, err := parser.Parse(cachedAuthInfoToJson(expiry, headersList))
				if err != nil {
					proxywasm.LogCriticalf("Was not able to parse auth info: %v", err)
					return
				}
				buffer := parsedInfo.Get().MarshalTo(nil)
				proxywasm.LogDebugf("Buffer is %v", string(buffer))
				proxywasm.SetSharedData(requestDomain+requestPath, buffer, cas)
				proxywasm.LogDebugf("Cached auth info for %v / %v", requestDomain, requestPath)
			}
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

	proxywasm.LogDebugf("Json string to store: %s ", jsonString)
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
			proxywasm.LogDebugf("Do not cache, since cache-control is %s", directiveArray[0])
			return -1, err
		case "max-age":
			maxAge, err := strconv.Atoi(directiveArray[1])
			if err != nil {
				return -1, err
			}
			return time.Now().Unix() + int64(maxAge), err
		default:
			continue
		}
	}
	proxywasm.LogDebugf("Did not find any cache directive to be handled. Header: %s", cacheControlHeader)
	return -1, err
}

func parseConfigFromJson(jsonString string) {

	// initialize with defaults
	config = defaultPluginConfig
	endpointAuthConfig = endpointAuthConfiguration{}

	parsedJson, err := parser.Parse(jsonString)

	if err != nil {
		proxywasm.LogCriticalf("Unable to parse config: %v, will use default", err)
		return
	}

	generalConfig := parsedJson.Get("general")
	authConfig := parsedJson.Get("endpoints")

	if generalConfig != nil {
		config = parsePluginConfigFromJson(generalConfig)
		proxywasm.LogDebugf("Parsed config: %v", config)
	}

	if authConfig != nil {
		parseAuthConfig(authConfig)
	}

	return
}

/**
* Parse the configuration to a tree-like map of maps for fast request path checking.
 */
func parseAuthConfig(authJson *fastjson.Value) {
	endpointAuthConfig = endpointAuthConfiguration{}

	authJsonObject, err := authJson.Object()
	if err != nil {
		proxywasm.LogCriticalf("Was not able to read endpoint configuration. %v", err)
		return
	}

	authJsonObject.Visit(func(k []byte, authEntry *fastjson.Value) {

		authType := string(k)

		authEntryObject, _ := authEntry.Object()
		if err != nil {
			proxywasm.LogWarnf("Was not able to read auth configuration for %s. %v", authType, err)
			return
		}

		authEntryObject.Visit(func(k []byte, domainEntry *fastjson.Value) {

			domainName := string(k)
			if _, ok := endpointAuthConfig[domainName]; !ok {
				// initialize map for domain
				endpointAuthConfig[domainName] = make(map[string]string)
			}

			domainEntryArray, err := domainEntry.Array()
			if err != nil {
				proxywasm.LogWarnf("Was not able to read domain config for %s. %v", domainName, err)
			}
			for _, entry := range domainEntryArray {
				pathEntry := string(entry.GetStringBytes())
				// the path matcher only takes sub-paths, if the pattern ends with a `*`.
				// this will lead to:
				//                   `/path` -> two entries [`/path`, `/path/*`] for exact match and subpaths to work
				//                   `/path/` -> one entry [`/path/*`] exact match already included
				if string(pathEntry[len(pathEntry)-1]) != "/" {
					endpointAuthConfig[domainName][pathEntry] = authType
					endpointAuthConfig[domainName][pathEntry+"/*"] = authType
				} else {
					endpointAuthConfig[domainName][pathEntry+"*"] = authType
				}
			}
			if len(endpointAuthConfig[domainName]) < 1 {
				proxywasm.LogWarnf("The config for %s-%s was empty, removing it.", authType, domainName)
				delete(endpointAuthConfig, domainName)
			}
		})

	})
}

/**
* Parse the jsonstring, containing the configuration
 */
func parsePluginConfigFromJson(parsedJson *fastjson.Value) (parsedConfig pluginConfiguration) {
	proxywasm.LogDebugf("Parse the config: %v", parsedJson)

	parsedConfig = defaultPluginConfig

	authRequestTimeout := parsedJson.GetInt("authRequestTimeout")
	authProviderName := parsedJson.GetStringBytes("authProviderName")
	authType := parsedJson.GetStringBytes("authType")
	// in case of error, the boolean zero value is used
	parsedConfig.enableEndpointMatching = parsedJson.GetBool("enableEndpointMatching")

	if authRequestTimeout > 0 {
		parsedConfig.authRequestTimeout = uint32(authRequestTimeout)
	} else {
		proxywasm.LogWarnf("Use default requestTimeout: %v", defaultPluginConfig.authRequestTimeout)
	}

	if authProviderName != nil {
		parsedConfig.authProviderName = string(authProviderName)
	} else {
		proxywasm.LogWarnf("Use default authProvider: %v", defaultPluginConfig.authProviderName)
	}

	if authType != nil {
		parsedConfig.authType = string(authType)
	} else {
		proxywasm.LogWarnf("Use default authType: %v", defaultPluginConfig.authType)
	}

	proxywasm.LogDebugf("Parsed config is %v", parsedConfig)

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
		nameBytes := entry.GetStringBytes("name")
		valueBytes := entry.GetStringBytes("value")
		if nameBytes == nil || valueBytes == nil {
			proxywasm.LogWarnf("Response from the auth-povider contained invalid entries.")
			continue
		}
		name := string(nameBytes)
		value := string(valueBytes)

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
