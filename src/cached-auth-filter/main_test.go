package main

import (
	"github.com/valyala/fastjson"
	"log"
	"testing"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/proxytest"
)

func TestPathMatching(t *testing.T) {
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
		var parser fastjson.Parser
		v, _ := parser.Parse(tc.testConfig)
		parseAuthConfig(v)

		authType, match := matchPath(tc.testDomain, tc.testPath)
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
