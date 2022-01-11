package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/valyala/fastjson"
)

type testRequest struct {
	domain        string
	path          string
	cached        bool
	expectIgnored bool
}

type authResponse struct {
	body    string
	headers [][2]string
}

func TestCaching(t *testing.T) {
	type test struct {
		testName          string
		testConfig        string
		testRequests      []testRequest
		testDomain        string
		testPath          string
		expectedAction    types.Action
		expectedHeaders   [][2]string
		expectExpectCache bool
		authResponse      authResponse
	}

	tests := []test{
		{testName: "Cache headers for responses.",
			testRequests:      []testRequest{{"domain.org", "/", false, false}, {"domain.org", "/", true, false}},
			testConfig:        "{}",
			expectedAction:    types.ActionPause,
			expectExpectCache: true,
			authResponse:      authResponse{`[{"name": "Authorization", "value": "token"}]`, [][2]string{{"HTTP/1.1", "200 OK"}, {"cache-control", "max-age=60"}}},
			expectedHeaders:   [][2]string{{"Authorization", "token"}}},
		{testName: "Cache headers for responses on multiple requests.",
			testRequests:      []testRequest{{"domain.org", "/", false, false}, {"other-domain.org", "/", false, false}, {"domain.org", "/", true, false}},
			testConfig:        "{}",
			expectedAction:    types.ActionPause,
			expectExpectCache: true,
			authResponse:      authResponse{`[{"name": "Authorization", "value": "token"}]`, [][2]string{{"HTTP/1.1", "200 OK"}, {"cache-control", "max-age=60"}}},
			expectedHeaders:   [][2]string{{"Authorization", "token"}}},
		{testName: "Cache headers for responses on multiple requests with endpoint matching.",
			testRequests:      []testRequest{{"domain.org", "/", false, false}, {"other-domain.org", "/", false, true}, {"domain.org", "/", true, false}},
			testConfig:        "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/\"]}}}",
			expectedAction:    types.ActionPause,
			expectExpectCache: true,
			authResponse:      authResponse{`[{"name": "Authorization", "value": "token"}]`, [][2]string{{"HTTP/1.1", "200 OK"}, {"cache-control", "max-age=60"}}},
			expectedHeaders:   [][2]string{{"Authorization", "token"}}},
		{testName: "No cache for responses on cache-control 'no-cache'.",
			testRequests:      []testRequest{{"domain.org", "/", false, false}, {"domain.org", "/", false, false}},
			testConfig:        "{}",
			expectedAction:    types.ActionPause,
			expectExpectCache: true,
			authResponse:      authResponse{`[{"name": "Authorization", "value": "token"}]`, [][2]string{{"HTTP/1.1", "200 OK"}, {"cache-control", "no-cache"}}},
			expectedHeaders:   [][2]string{{"Authorization", "token"}}},
		{testName: "No cache for responses on cache-control 'no-store'.",
			testRequests:      []testRequest{{"domain.org", "/", false, false}, {"domain.org", "/", false, false}},
			testConfig:        "{}",
			expectedAction:    types.ActionPause,
			expectExpectCache: true,
			authResponse:      authResponse{`[{"name": "Authorization", "value": "token"}]`, [][2]string{{"HTTP/1.1", "200 OK"}, {"cache-control", "no-store"}}},
			expectedHeaders:   [][2]string{{"Authorization", "token"}}},
		{testName: "No cache for responses on cache-control 'must-revalidate.",
			testRequests:      []testRequest{{"domain.org", "/", false, false}, {"domain.org", "/", false, false}},
			testConfig:        "{}",
			expectedAction:    types.ActionPause,
			expectExpectCache: true,
			authResponse:      authResponse{`[{"name": "Authorization", "value": "token"}]`, [][2]string{{"HTTP/1.1", "200 OK"}, {"cache-control", "must-revalidate"}}},
			expectedHeaders:   [][2]string{{"Authorization", "token"}}},
	}

	for _, tc := range tests {

		t.Run(tc.testName, func(t *testing.T) {
			opt := proxytest.NewEmulatorOption().WithPluginConfiguration([]byte(tc.testConfig)).WithVMContext(&vmContext{})
			host, reset := proxytest.NewHostEmulator(opt)
			defer reset()
			// required to not fail due to logging
			log.Print("TestOnHttpRequestHeadersWithCaching +++++++++++++++++++++ Running test: " + tc.testName)

			// Initialize http context.
			id := host.InitializeHttpContext()

			for _, request := range tc.testRequests {

				log.Print("Current request: " + fmt.Sprint(request))

				hs := [][2]string{{":authority", request.domain}, {":path", request.path}}

				action := host.CallOnRequestHeaders(id, hs, true)

				if request.cached || request.expectIgnored {
					if action != types.ActionContinue {
						t.Errorf("%s: Request was expected to be served from cache, but action is %v.", tc.testName, action)
					}
				} else {
					if action != types.ActionPause {
						t.Errorf("%s: Request was not  expected to be served from cache, but action is %v.", tc.testName, action)
					}
				}
				if !request.cached && !request.expectIgnored {
					attrs := host.GetCalloutAttributesFromContext(id)
					body := []byte(tc.authResponse.body)
					host.CallOnHttpCallResponse(attrs[0].CalloutID, tc.authResponse.headers, nil, body)
				}
				// we do not verify headers for ignore case here, since its already tested extensively in the TestOnHttpRequestHeaders
				if !request.expectIgnored {
					resultHeaders := host.GetCurrentRequestHeaders(id)
					verifyHeaders(t, hs, tc.expectedHeaders, resultHeaders, tc.testName)
					endAction := host.GetCurrentHttpStreamAction(id)
					verifyEndAction(t, endAction, tc.testName)
				}
			}

		})

	}

}

func verifyHeaders(t *testing.T, generalHeaders, expectedHeaders, resultHeaders [][2]string, testName string) {
	if len(generalHeaders)+len(expectedHeaders) != len(resultHeaders) {
		t.Errorf("%s: To much headers on request. Was expected to be %v, but was %v.", testName, len(generalHeaders)+len(expectedHeaders), len(resultHeaders))
		return
	}
	for _, v := range resultHeaders {
		var contains bool
		for _, eh := range expectedHeaders {
			if eh == v {
				contains = true
			}
		}

		for _, eh := range generalHeaders {
			if eh == v {
				contains = true
			}
		}
		if !contains {
			t.Errorf("%s: Header %v was not expected.", testName, v)
		}
	}
}

func TestOnHttpRequestHeaders(t *testing.T) {

	type test struct {
		testName        string
		testConfig      string
		testDomain      string
		testPath        string
		expectedAction  types.Action
		expectedHeaders [][2]string
		expectExtCall   bool
		authResponse    string
	}

	tests := []test{
		{testName: "Do nothing for no domain.", testPath: "/", testDomain: "",
			testConfig:      "{}",
			expectedAction:  types.ActionContinue,
			expectedHeaders: [][2]string{}},
		{testName: "Do nothing for no path.", testPath: "", testDomain: "domain.org",
			testConfig:      "{}",
			expectedAction:  types.ActionContinue,
			expectedHeaders: [][2]string{}},
		{testName: "Add token for everything configured.", testPath: "/", testDomain: "domain.org",
			testConfig:      "{}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}}},
		{testName: "No-match for nothing at domain configured.", testPath: "/", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"other-domain.org\": [\"/\"]}}}",
			expectedAction:  types.ActionContinue,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{}},
		{testName: "Match for exact path and domain config.", testPath: "/", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/\"]}}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}}},
		{testName: "Match for exact path and domain at multiple auth config.", testPath: "/", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/\"]}, \"OIC\": { \"domain2.org\": [\"/\"] }}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}}},
		{testName: "Match for exact path and domain at multiple domain config.", testPath: "/", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/\"], \"domain2.org\": [\"/\"] }}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}},
		},
		{testName: "Match for exact path and domain at multiple auth and path config.", testPath: "/", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/\"]}, \"OIC\": { \"domain.org\": [\"/oic\"] }}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}},
		},
		{testName: "Match for exact sub-path.", testPath: "/sub-path", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/\", \"/sub-path\"] }}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}},
		},
		{testName: "Match for exact sub-path and multiple auth-types.", testPath: "/oic", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/\"]}, \"OIC\": { \"domain.org\": [\"/oic\"] }}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}},
		},
		{testName: "Match for sub-path of configured path.", testPath: "/sub-path", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/\"]}}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}},
		},
		{testName: "Match for sub-path of sub-path.", testPath: "/sub-path/p2", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/sub-path/\"]}}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}},
		},
		{testName: "Match for sub-path of sub-path without /.", testPath: "/sub-path/p2", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/sub-path\"]}}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}},
		},
		{testName: "Match for sub-path of in complex config.", testPath: "/sub-path/p2", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/sub-path\"]}, \"OIC\": {\"domain.org\":[\"/\"]}}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}},
		},

		{testName: "Match for sub-path of in complex config with multiple headers.", testPath: "/sub-path/p2", testDomain: "domain.org",
			testConfig:      "{\"general\":{\"enableEndpointMatching\":true},\"endpoints\":{\"ISHARE\":{\"domain.org\": [\"/sub-path\"]}, \"OIC\": {\"domain.org\":[\"/\"]}}}",
			expectExtCall:   true,
			expectedAction:  types.ActionPause,
			authResponse:    `[{"name": "Authorization", "value": "token"}, {"name": "Other-header", "value": "header-2"}]`,
			expectedHeaders: [][2]string{{"Authorization", "token"}, {"Other-header", "header-2"}},
		},
	}

	for _, tc := range tests {

		t.Run(tc.testName, func(t *testing.T) {
			opt := proxytest.NewEmulatorOption().WithPluginConfiguration([]byte(tc.testConfig)).WithVMContext(&vmContext{})
			host, reset := proxytest.NewHostEmulator(opt)
			defer reset()
			// required to not fail due to logging
			log.Print("TestOnHttpRequestHeaders +++++++++++++++++++++ Running test: " + tc.testName)

			// Initialize http context.
			id := host.InitializeHttpContext()

			hs := [][2]string{}
			if tc.testDomain != "" {
				hs = append(hs, [][2]string{{":authority", tc.testDomain}}...)
			}
			if tc.testPath != "" {
				hs = append(hs, [][2]string{{":path", tc.testPath}}...)
			}

			action := host.CallOnRequestHeaders(id, hs, true)
			if action != tc.expectedAction {
				t.Errorf("%s: Action was expected to be %v, but was %v.", tc.testName, tc.expectedAction, action)
			}
			attrs := host.GetCalloutAttributesFromContext(id)
			body := []byte(tc.authResponse)
			headers := [][2]string{
				{"HTTP/1.1", "200 OK"},
			}

			if tc.expectExtCall {
				host.CallOnHttpCallResponse(attrs[0].CalloutID, headers, nil, body)
			}

			resultHeaders := host.GetCurrentRequestHeaders(id)

			verifyHeaders(t, hs, tc.expectedHeaders, resultHeaders, tc.testName)
			endAction := host.GetCurrentHttpStreamAction(id)
			verifyEndAction(t, endAction, tc.testName)
		})

	}

}

func verifyEndAction(t *testing.T, endAction types.Action, testName string) {
	if endAction != types.ActionContinue {
		t.Errorf("%s: Request should continue in any case, but did %v.", testName, endAction)
	}
}

func TestPathMatching(t *testing.T) {
	var parser fastjson.Parser

	// required to not fail due to logging
	opt := proxytest.NewEmulatorOption().WithVMContext(&vmContext{})
	_, reset := proxytest.NewHostEmulator(opt)
	defer reset()

	type test struct {
		testName         string
		testConfig       string
		testDomain       string
		testPath         string
		expectedResult   bool
		expectedAuthType string
	}

	tests := []test{
		{testName: "No-match for no domain provided.", testConfig: "{}", testPath: "/", testDomain: "", expectedResult: false},
		{testName: "No-match for no path provided.", testConfig: "{}", testPath: "", testDomain: "domain.org", expectedResult: false},
		{testName: "No-match for nothing configured.", testConfig: "{}", testPath: "/", testDomain: "domain.org", expectedResult: false},
		{testName: "No-match for nothing at domain configured.", testPath: "/", testDomain: "domain.org", expectedResult: false,
			testConfig: "{\"ISHARE\":{\"other-domain.org\": [\"/\"]}}"},
		{testName: "Match for exact path and domain config.", testPath: "/", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/\"]}}"},
		{testName: "Match for exact path and domain at multiple auth config.", testPath: "/", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/\"]}, \"OIC\": { \"domain2.org\": [\"/\"] }}"},
		{testName: "Match for exact path and domain at multiple domain config.", testPath: "/", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/\"], \"domain2.org\": [\"/\"] }}"},
		{testName: "Match for exact path and domain at multiple auth and path config.", testPath: "/", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/\"]}, \"OIC\": { \"domain.org\": [\"/oic\"] }}"},
		{testName: "Match for exact sub-path.", testPath: "/sub-path", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/\", \"/sub-path\"] }}"},
		{testName: "Match for exact sub-path and multiple auth-types.", testPath: "/oic", testDomain: "domain.org", expectedResult: true, expectedAuthType: "OIC",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/\"]}, \"OIC\": { \"domain.org\": [\"/oic\"] }}"},
		{testName: "Match for sub-path of configured path.", testPath: "/sub-path", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/\"]}}"},
		{testName: "Match for sub-path of sub-path.", testPath: "/sub-path/p2", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/sub-path/\"]}}"},
		{testName: "Match for sub-path of sub-path without /.", testPath: "/sub-path/p2", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/sub-path\"]}}"},
		{testName: "Match for sub-path of in complex config.", testPath: "/sub-path/p2", testDomain: "domain.org", expectedResult: true, expectedAuthType: "ISHARE",
			testConfig: "{\"ISHARE\":{\"domain.org\": [\"/sub-path\"]}, \"OIC\": {\"domain.org\":[\"/\"]}}"},
	}

	for _, tc := range tests {
		log.Print("TestPathMatching +++++++++++++++++++++ Running test: " + tc.testName)

		parsedJson, _ := parser.Parse(tc.testConfig)
		parseAuthConfig(parsedJson)

		authType, match := matchEndpoint(tc.testDomain, tc.testPath)
		if match != tc.expectedResult {
			t.Errorf("%s: Match was expected to be %v, but was %v.", tc.testName, tc.expectedResult, match)
			continue
		}
		if authType != tc.expectedAuthType {
			t.Errorf("%s: Authtype was expected to be %s, but was %s.", tc.testName, tc.expectedAuthType, authType)
			continue
		}
	}
}
