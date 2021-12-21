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
	"encoding/json"
	"strings"
	"time"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"

	cacheobject "github.com/pquerna/cachecontrol/cacheobject"
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
		var cachedAuthInfo CachedAuthInformation
		json.Unmarshal(data, &cachedAuthInfo)
		if cachedAuthInfo.expirationTime >= time.Now().Unix() {
			cas = currentCas
			return requestAuthProvider()
		} else {
			addCachedHeadersToRequest(cachedAuthInfo.cachedHeaders)
			return types.ActionContinue
		}

	}

	return types.ActionContinue
}

func addCachedHeadersToRequest(cachedHeaders HeadersList) {
	for _, header := range cachedHeaders {
		proxywasm.AddHttpRequestHeader(header.Name, header.Value)
	}
}

func requestAuthProvider() types.Action {
	proxywasm.LogInfof("Call to " + clusterName)
	hs, _ := proxywasm.GetHttpRequestHeaders()

	var methodIndex int
	var pathIndex int
	for i, h := range hs {
		proxywasm.LogInfof("++++++++++++++++ original header " + h[0] + " - " + h[1])
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

	var headersList HeadersList
	json.Unmarshal(body, &headersList)

	addCachedHeadersToRequest(headersList)
	// continue the request before handling the caching
	proxywasm.ResumeHttpRequest()

	// handle cachecontrol
	for _, h := range headers {
		proxywasm.LogInfof("Current header: " + h[0] + " - " + h[1])
		if h[0] == "Cache-Control" {
			resDirective, _ := cacheobject.ParseResponseCacheControl(h[1])
			if resDirective.NoCachePresent || resDirective.NoStore || resDirective.MustRevalidate {
				return
			}
			maxAge := resDirective.MaxAge
			if maxAge >= 0 {
				expiry := time.Now().Unix() + int64(maxAge)
				cacheAuthInfo := CachedAuthInformation{expirationTime: expiry, cachedHeaders: headersList}
				cacheObject, err := json.Marshal(cacheAuthInfo)
				if err != nil {
					proxywasm.LogCriticalf("Was not able to cache the auth information.")
					return
				}
				proxywasm.SetSharedData(domain+path, cacheObject, cas)
			}
			proxywasm.LogInfof("Cached auth info for %v / %v", domain, path)
			return
		}

	}

}
