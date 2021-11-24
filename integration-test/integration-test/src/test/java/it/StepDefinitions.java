package it;

import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import org.fiware.sidecar.config.ApiClient;
import org.fiware.sidecar.config.api.EndpointConfigurationApi;
import org.fiware.sidecar.config.model.AuthCredentialsVO;
import org.fiware.sidecar.config.model.AuthTypeVO;
import org.fiware.sidecar.config.model.EndpointRegistrationVO;
import org.fiware.sidecar.credentials.api.CredentialsManagementApi;
import org.fiware.sidecar.credentials.model.IShareCredentialsVO;

import java.nio.file.Files;
import java.nio.file.Path;

public class StepDefinitions {

	private static EndpointConfigurationApi endpointConfigurationApi;

	{
		ApiClient apiClient = new ApiClient();
		apiClient.setBasePath("http://localhost:9090");
		endpointConfigurationApi = new EndpointConfigurationApi(apiClient);
	}

	private static CredentialsManagementApi credentialsManagementApi;

	{
		org.fiware.sidecar.credentials.ApiClient apiClient = new org.fiware.sidecar.credentials.ApiClient();
		apiClient.setBasePath("http://localhost:7070");
		credentialsManagementApi = new CredentialsManagementApi(apiClient);
	}


	@Given("Echo-server is configured as an iShare endpoint.")
	public void echo_server_is_configured_as_an_i_share_endpoint() throws Exception {

		String signingKey =
				Files.readString(
						Path.of(
								getClass().getResource("/test-files/test-signingkey.pem").toURI()));

		String certificateChain =
				Files.readString(
						Path.of(getClass().getResource("/test-files/test-crt.pem").toURI()));

		EndpointRegistrationVO endpointRegistrationVO = new EndpointRegistrationVO()
				.domain("localhost")
				.port(6060)
				.path("/")
				.authType(AuthTypeVO.ISHARE)
				.useHttps(false)
				.authCredentials(
						new AuthCredentialsVO()
								.iShareClientId("iShareProviderClientId")
								.iShareIdpId("iShareSubscriberClientId")
								.iShareIdpAddress("http://localhost:1080/oauth2/token")
								.requestGrantType("client_credentials"));
		endpointConfigurationApi.createEndpoint(endpointRegistrationVO);

		IShareCredentialsVO iShareCredentialsVO = new IShareCredentialsVO()
				.certificateChain(certificateChain)
				.signingKey(signingKey);
		credentialsManagementApi.postCredentials("iShareProviderClientId", iShareCredentialsVO);
	}

	@When("Client sends a request to the echo-server.")
	public void client_sends_a_request_to_the_echo_server() throws Exception {

		// call 6060 since thats the intercepted path to echo-server
		Request request = new Request.Builder()
				.url("http://localhost:6060/")
				.build();
		OkHttpClient okHttpClient = new OkHttpClient();
		okHttpClient.newCall(request).execute();
	}

	@Then("Echo-server should receive a request with an authorization-header.")
	public void echo_server_should_receive_a_request_with_an_authorization_header() {
		// Write code here that turns the phrase above into concrete actions
		throw new io.cucumber.java.PendingException();
	}


}
