package it;

import com.fasterxml.jackson.databind.ObjectMapper;
import io.cucumber.java.After;
import io.cucumber.java.Before;
import io.cucumber.java.an.E;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import org.awaitility.Awaitility;
import org.fiware.sidecar.config.ApiClient;
import org.fiware.sidecar.config.ApiException;
import org.fiware.sidecar.config.api.EndpointConfigurationApi;
import org.fiware.sidecar.config.model.AuthCredentialsVO;
import org.fiware.sidecar.config.model.AuthTypeVO;
import org.fiware.sidecar.config.model.EndpointRegistrationVO;
import org.fiware.sidecar.credentials.api.CredentialsManagementApi;
import org.fiware.sidecar.credentials.model.IShareCredentialsVO;
import org.junit.jupiter.api.Assertions;

import java.io.IOException;
import java.net.URISyntaxException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.time.temporal.ChronoUnit;
import java.util.List;
import java.util.Map;

import static java.lang.Thread.sleep;

public class StepDefinitions {

	private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

	private static final String CONFIG_HOST = "10.5.0.5";
	private static final String AUTH_HOST = "10.5.0.6";
	private static final String ECHO_HOST = "10.5.0.2";
	private static final String ECHO_2_HOST = "10.5.0.8";
	private static final String IDP_HOST = "10.5.0.7";

	public static final int WAIT_TIMEOUT = 3000;
	private static EndpointConfigurationApi endpointConfigurationApi;

	{
		ApiClient apiClient = new ApiClient();
		apiClient.setBasePath(String.format("http://%s:9090", CONFIG_HOST));
		endpointConfigurationApi = new EndpointConfigurationApi(apiClient);
	}

	private static CredentialsManagementApi credentialsManagementApi;

	{
		org.fiware.sidecar.credentials.ApiClient apiClient = new org.fiware.sidecar.credentials.ApiClient();
		apiClient.setBasePath(String.format("http://%s:7070", AUTH_HOST));
		credentialsManagementApi = new CredentialsManagementApi(apiClient);
	}

	@Given("The Data-provider is running with the endpoint-authentication-service as a sidecar-proxy.")
	public void setup_sidecar_in_docker() throws Exception {
		// since the setup is relatively complex and does require root-permission, we only check that it is running here.
		Request request = new Request.Builder()
				.url(String.format("http://%s:9090/health", CONFIG_HOST))
				.build();
		OkHttpClient okHttpClient = new OkHttpClient();

		Awaitility
				.await()
				.atMost(Duration.of(10, ChronoUnit.SECONDS))
				.untilAsserted(() -> Assertions.assertEquals(200, okHttpClient.newCall(request).execute().code(), "We expect the setup to run before starting."));
	}

	@Given("Data-Consumer's root path is configured as an iShare endpoint.")
	public void echo_server_is_configured_as_an_i_share_endpoint() throws Exception {

		configureDataConsumer(ECHO_HOST, 6060, "/");
	}

	@Given("Data-Consumer subpath is configured as an iShare endpoint.")
	public void echo_server_sub_path_is_configured_as_an_i_share_endpoint() throws Exception {

		configureDataConsumer(ECHO_HOST, 6060, "/subpath");
	}


	@Given("Data-Consumer anotherpath is configured as an iShare endpoint.")
	public void echo_server_another_path_is_configured_as_an_i_share_endpoint() throws Exception {

		configureDataConsumer(ECHO_HOST, 6060, "/anotherpath");
	}


	@Before
	public void beforeScenario() throws Exception {
		cleanupConfiguration();
	}


	@After
	public void afterScenario() throws Exception {
		cleanupConfiguration();
	}

	private void cleanupConfiguration() throws Exception {
		endpointConfigurationApi.getEndpoints().forEach(endpointInfoVO -> {
			try {
				endpointConfigurationApi.deleteEndpoint(endpointInfoVO.getId());
			} catch (ApiException a) {
				// swallow exceptions to get it clean
			}
		});

		credentialsManagementApi.getCredentialsList().forEach(credential -> {
			try {
				credentialsManagementApi.deleteCredentials(credential);
			} catch (Exception a) {
				// swallow exceptions to get it clean
			}
		});
		// cleanup echo server
		Request request = new Request.Builder()
				.method("DELETE", null)
				.url(String.format("http://%s:6061/last", ECHO_HOST))
				.build();
		OkHttpClient okHttpClient = new OkHttpClient();
		okHttpClient.newCall(request).execute();


		Request requestEcho2 = new Request.Builder()
				.method("DELETE", null)
				.url(String.format("http://%s:6061/last", ECHO_2_HOST))
				.build();
		okHttpClient.newCall(request).execute();

		// wait a second, so that envoy has time to update its config
		sleep(WAIT_TIMEOUT);
	}

	@Given("No other endpoint is configured.")
	public void no_other_endpoint_configured() throws Exception {
		Assertions.assertEquals(1, endpointConfigurationApi.getEndpoints().size(), "Exactly one endpoint should be configured.");
	}

	@Given("No endpoint is configured.")
	public void no_endpoint_is_configured() throws Exception {
		Assertions.assertEquals(0, endpointConfigurationApi.getEndpoints().size(), "No endpoint should be configured.");

	}

	@When("Data-Provider sends a request to the data-consumer's root path.")
	public void client_sends_a_request_to_the_echo_server() throws Exception {
		sendRequestToEchoServer(ECHO_HOST, "http://%s:6060/");
	}

	@When("Data-Provider sends a request to a sub-path of the data-consumer.")
	public void client_sends_a_request_to_a_sub_path_of_the_echo_server() throws Exception {

		// call 6060 since that is the intercepted path to echo-server
		sendRequestToEchoServer(ECHO_HOST, "http://%s:6060/subpath");
	}

	@When("Data-Provider sends a request to the data-consumer-2's root path.")
	public void client_sends_a_request_to_the_echo_server_2() throws Exception {

		// call 6060 since that is the intercepted path to echo-server
		sendRequestToEchoServer(ECHO_2_HOST, "http://%s:6060/");
	}

	@When("Data-Provider sends a request to a sub-path of the data-consumer-2.")
	public void client_sends_a_request_to_a_sub_path_of_the_echo_server_2() throws Exception {

		// call 6060 since that is the intercepted path to echo-server
		sendRequestToEchoServer(ECHO_2_HOST, "http://%s:6060/subpath");
	}

	@Then("Data-Consumer should receive a request with an authorization-header.")
	public void echo_server_should_receive_a_request_with_an_authorization_header() throws Exception {

		assertLastRequestHasToken(ECHO_HOST, "myIShareToken", "The auth-token as it is provided by the mock-idp should have been sent.");
	}

	@Then("Data-Consumer-2 should receive a request with an authorization-header.")
	public void echo_server_2_should_receive_a_request_with_an_authorization_header() throws Exception {

		assertLastRequestHasToken(ECHO_2_HOST, "myIShareToken", "The auth-token as it is provided by the mock-idp should have been sent.");
	}

	@Then("Data-Consumer should receive requests with an authorization-header at the subpath.")
	public void echo_server_should_receive_a_request_with_authorization_header_at_the_subpath() throws Exception {
		assertTokenAtSubpath("/subpath", "myIShareToken", ECHO_HOST, "Authorization should have been added.");
	}


	@Then("Data-Consumer should receive requests without an authorization-header at root path.")
	public void consumer_receives_request_without_header_at_root() throws Exception {
		assertTokenAtSubpath("/", null, ECHO_HOST, "Authorization should not have been added.");
	}

	@Then("Data-Consumer should receive requests with an authorization-header.")
	public void echo_server_should_receive_a_request_with_authorization_header() throws Exception {
		assertAllRequestsHaveToken(ECHO_HOST, "myIShareToken", "A token should have been added.");
	}

	@Then("Data-Consumer-2 should receive requests with an authorization-header.")
	public void echo_server_2_should_receive_requests_with_authorization_header() throws Exception {
		assertAllRequestsHaveToken(ECHO_2_HOST, "myIShareToken", "A token should have been added.");
	}

	@Then("Data-Consumer should receive a request without an authorization-header.")
	public void echo_server_should_receive_a_request_without_an_authorization_header() throws Exception {

		assertLastRequestHasToken(ECHO_HOST, "", "No auth token should have been added.");
	}

	@Then("Data-Consumer-2 should receive requests without an authorization-header.")
	public void echo_server_2_should_receive_a_request_without_an_authorization_header() throws Exception {
		assertThatNoHeaderWasAdded("authorization", ECHO_2_HOST, "No authorization should have been added.");

	}


	@Then("Data-Consumer should receive requests without an authorization-header.")
	public void echo_server_should_receive_requests_without_an_authorization_header() throws Exception {
		assertThatNoHeaderWasAdded("authorization", ECHO_HOST, "No authorization should have been added.");
	}

	private void assertAllRequestsHaveToken(String echoHost, String token, String message) throws Exception {
		Request request = new Request.Builder()
				.url(String.format("http://%s:6061/last/all", echoHost))
				.build();
		OkHttpClient okHttpClient = new OkHttpClient();
		Response response = okHttpClient.newCall(request).execute();
		List<Map> responseList = OBJECT_MAPPER.readValue(response.body().string(), List.class);

		responseList.stream().forEach(m -> Assertions.assertEquals(token, ((Map<Object, Object>) m.get("headers")).get("authorization"), message));

	}


	private void assertTokenAtSubpath(String path, String token, String echoHost, String message) throws Exception {
		Request request = new Request.Builder()
				.url(String.format("http://%s:6061/last/all", echoHost))
				.build();
		OkHttpClient okHttpClient = new OkHttpClient();
		Response response = okHttpClient.newCall(request).execute();
		List<Map> responseList = OBJECT_MAPPER.readValue(response.body().string(), List.class);

		responseList.stream()
				.filter(m -> m.get("path").equals(path))
				.map(m -> m.get("headers"))
				.forEach(m -> Assertions.assertEquals(token, ((Map) m).get("authorization"), message));
	}

	private void assertThatNoHeaderWasAdded(String header, String echoHost, String message) throws Exception {
		Request request = new Request.Builder()
				.url(String.format("http://%s:6061/last/all", echoHost))
				.build();
		OkHttpClient okHttpClient = new OkHttpClient();
		Response response = okHttpClient.newCall(request).execute();
		List<Map> responseList = OBJECT_MAPPER.readValue(response.body().string(), List.class);

		responseList.stream().forEach(m -> Assertions.assertFalse(((Map<Object, Object>) m.get("headers")).containsKey(header), message));

	}

	private void configureDataConsumer(String domain, int port, String path) throws IOException, URISyntaxException, org.fiware.sidecar.credentials.ApiException, ApiException, InterruptedException {
		String signingKey =
				Files.readString(
						Path.of(
								getClass().getResource("/test-files/test-signingkey.pem").toURI()));

		String certificateChain =
				Files.readString(
						Path.of(getClass().getResource("/test-files/test-crt.pem").toURI()));
		try {
			EndpointRegistrationVO endpointRegistrationVO = new EndpointRegistrationVO()
					.domain(domain)
					.port(port)
					.path(path)
					.authType(AuthTypeVO.ISHARE)
					.useHttps(false)
					.authCredentials(
							new AuthCredentialsVO()
									.iShareClientId("iShareProviderClientId")
									.iShareIdpId("iShareSubscriberClientId")
									.iShareIdpAddress(String.format("http://%s:1080/oauth2/token", IDP_HOST))
									.requestGrantType("client_credentials"));
			endpointConfigurationApi.createEndpoint(endpointRegistrationVO);

			if (!credentialsManagementApi.getCredentialsList().contains("iShareProviderClientId")) {
				IShareCredentialsVO iShareCredentialsVO = new IShareCredentialsVO()
						.certificateChain(certificateChain)
						.signingKey(signingKey);
				credentialsManagementApi.postCredentials("iShareProviderClientId", iShareCredentialsVO);
			}
		} catch (ApiException a) {
			if (a.getCode() == 409) {
				return;
			}
			throw a;
		}

		// wait a second, so that envoy has time to update its config
		sleep(WAIT_TIMEOUT);
	}

	private void assertLastRequestHasToken(String echo2Host, String myIShareToken, String s) throws IOException {
		Request request = new Request.Builder()
				.url(String.format("http://%s:6061/last/headers/authorization", echo2Host))
				.build();
		OkHttpClient okHttpClient = new OkHttpClient();
		Response response = okHttpClient.newCall(request).execute();
		Assertions.assertEquals(myIShareToken, response.body().string(), s);
	}

	private void sendRequestToEchoServer(String echoHost, String s) throws IOException {
		// call 6060 since that is the intercepted path to echo-server
		Request request = new Request.Builder()
				// currently required in local setups
				.header("x-envoy-original-dst-host", String.format("%s:6060", echoHost))
				.url(String.format(s, echoHost))
				.build();
		OkHttpClient okHttpClient = new OkHttpClient();
		okHttpClient.newCall(request).execute();
	}

}
