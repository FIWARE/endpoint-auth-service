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
	authType           = "ISHARE"
	clusterName        = "ext-authz"
	authRequestTimeout = 5000
	authorityKey       = ":authority"
	pathKey            = ":path"
)

/**
* Global compare & set value for cache control
 */
var cas uint32 = 0
var domain string
var path string

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

// Override types.VMContext.
func (*vmContext) OnVMStart(vmConfigurationSize int) types.OnVMStartStatus {

	proxywasm.LogInfo("Successfully started.")
	return types.OnVMStartStatusOK
}

// Override types.DefaultPluginContext.
func (ctx pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
	}

	proxywasm.LogInfof("plugin config: %s", string(data))
	proxywasm.LogInfo("Successfully read config and started.")
	return types.OnVMStartStatusOK
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

// Override types.DefaultPluginContext.
func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{}
}

// Override types.DefaultHttpContext.
func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {

	authorityHeader, err := proxywasm.GetHttpRequestHeader(authorityKey)
	if err != nil || authorityHeader == "" {
		proxywasm.LogCriticalf("Failed to get authority header: %v", err)
		return types.ActionContinue
	}
	domain = strings.Split(authorityHeader, ":")[0]

	pathHeader, err := proxywasm.GetHttpRequestHeader(pathKey)
	if err != nil || pathHeader == "" {
		proxywasm.LogCriticalf("Failed to get path header: %v", err)
		return types.ActionContinue
	}
	path = pathHeader

	return setHeader()
}

func setHeader() types.Action {
	sharedDataKey := domain + path
	data, currentCas, err := proxywasm.GetSharedData(sharedDataKey)

	if err != nil || data == nil {
		return requestAuthProvider()
	}

	if data != nil {
		proxywasm.LogDebugf("Cache hit: ", string(data))
		cachedAuthInfo, err := parsCachedAuthInformation(string(data))
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

func addCachedHeadersToRequest(cachedHeaders HeadersList) {
	for _, header := range cachedHeaders {
		proxywasm.LogDebugf("Add header ", fmt.Sprint(header))
		proxywasm.AddHttpRequestHeader(header.Name, header.Value)
	}
}

func requestAuthProvider() types.Action {

	proxywasm.LogDebugf("Call to ", clusterName)
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
	hs[pathIndex] = [2]string{":path", "/" + authType + "/auth?domain=" + domain + "&path=" + path}

	if _, err := proxywasm.DispatchHttpCall(clusterName, hs, nil, nil, authRequestTimeout, authCallback); err != nil {
		proxywasm.LogDebugf("Domain " + domain + " , path: " + path + " , authType: " + authType)
		proxywasm.LogCriticalf("Call to auth-provider failed: %v", err)
		return types.ActionContinue
	}
	return types.ActionPause
}

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

func cachedAuthInfoToJson(expirationTime int64, cachedHeaders HeadersList) (jsonString string) {

	headerArray := `[`
	for i, header := range cachedHeaders {
		if i != 0 {
			headerArray = headerArray + `,`
		}
		headerArray = headerArray + `{"name":"` + header.Name + `","value":"` + header.Value + `"}`
	}
	headerArray = headerArray + `]`

	// TODO: use parser.MarshalTo
	jsonString = fmt.Sprintf(`{"expiration":%d, "cachedHeaders":%s}`, expirationTime, headerArray)

	proxywasm.LogDebugf("Json string to store ", jsonString)
	return jsonString
}

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

func parsCachedAuthInformation(jsonString string) (cachedAuthInfo CachedAuthInformation, err error) {

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

func parseHeaderList(jsonString string) (headerList HeadersList, err error) {

	var parser fastjson.Parser
	parsedJson, err := parser.Parse(jsonString)
	jsonArray, err := parsedJson.Array()
	return parseHeaderArray(jsonArray)

}
