package com.equinix.smartkey.kubernetes.kms;

import static org.junit.Assert.assertTrue;

import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.net.URL;
import java.net.URLDecoder;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Set;

import org.apache.http.HttpEntity;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.HttpEntityEnclosingRequestBase;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpHead;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.methods.HttpPut;
import org.apache.http.client.methods.HttpRequestBase;
import org.apache.log4j.Logger;
import org.json.JSONObject;

import com.mashape.unirest.http.HttpMethod;
import com.mashape.unirest.http.HttpResponse;
import com.mashape.unirest.http.JsonNode;
import com.mashape.unirest.http.Unirest;
import com.mashape.unirest.http.exceptions.UnirestException;
import com.mashape.unirest.http.options.Option;
import com.mashape.unirest.http.options.Options;
import com.mashape.unirest.http.utils.ClientFactory;
import com.mashape.unirest.http.utils.ResponseUtils;
import com.mashape.unirest.request.BaseRequest;
import com.mashape.unirest.request.HttpRequest;

@SuppressWarnings("serial")
public class APIServices {

	protected static final Logger log = Logger.getLogger(APIServices.class);

	public static final String AUTH_PATH = "/sys/v1/session/auth";
	public static final String KEYS_PATH = "/crypto/v1/keys";
	public static final String ENCRYPT_PATH = "/crypto/v1/keys/{{kid}}/encrypt";
	public static final String DECRYPT_PATH = "/crypto/v1/keys/{{kid}}/decrypt";
	public static final String DELETE_KID_PATH = "/crypto/v1/keys/{{kid}}";
	

	private static final String SO_NAME = "so_apiservices";
	private static final String TOKEN;

	// Authentication API
	static {
		JSONObject json = null;
		try {
			HttpResponse<InputStream> res = call(Unirest.post(SmartKeyPlugin.API_BASE_URL + AUTH_PATH).headers(new HashMap<String, String>() {
				{
					put("Authorization", "Basic " + SmartKeyPlugin.API_KEY); // Header
				}
			}).body(""), InputStream.class);
			if (res.getStatus() == 200 || res.getStatus() == 201) {
				json = new JsonNode(responseToString(res)).getObject();
			} else {
				throw new IOException("KEYS_PATH - Status Code: " + res.getStatus() + ", Body: " + responseToString(res));
			}
		} catch (Exception e) {
			e.printStackTrace();
			System.exit(1);
		}
		// Output Result
		TOKEN = "Bearer " + json.getString("access_token");
		log.debug(TOKEN);
	}
	
	public static String[] generateAES() throws UnirestException, IOException {
		String kid = null;
		String iv = null;
		String base64 = Base64.getEncoder().encodeToString("Us3rnam3/MyPa$$w0rd".getBytes()); // Text to encode
		try {
			HttpResponse<InputStream> res = call(Unirest.post(SmartKeyPlugin.API_BASE_URL + KEYS_PATH).headers(new HashMap<String, String>() {
				{
					put("Authorization", TOKEN);
				}
			}).body("{" + 
					"  \"name\":\"" + SO_NAME + "\"," + 
					"  \"description\":\"AES Key for Testing\"," + 
					"  \"key_size\": 256," + 
					"  \"obj_type\": \"AES\"" + 
					"}"), InputStream.class);
			if (res.getStatus() != 201) {
				throw new IOException("KEYS_PATH - Status Code: " + res.getStatus() + ", Body: " + responseToString(res));
			}
			JSONObject json = new JsonNode(responseToString(res)).getObject();
			
			kid = json.getString("kid");
			
			// Encrypt API
			res = call(Unirest.post(SmartKeyPlugin.API_BASE_URL + ENCRYPT_PATH.replace("{{kid}}", kid)).headers(new HashMap<String, String>() {
				{
					put("Authorization", TOKEN);
				}
			}).body("{" + 
					"	\"alg\": \"AES\"," + 
					"	\"plain\": \"" + base64 + "\"," + 
					"	\"mode\": \"CFB\"" + 
					"}"), InputStream.class);
			
			if (res.getStatus() != 200) {
				throw new IOException("ENCRYPT_PATH - Status Code: " + res.getStatus() + ", Body: " + responseToString(res));
			}
			json = new JsonNode(responseToString(res)).getObject();
			
			String cipher = json.getString("cipher");
			iv = json.getString("iv");
			
			// Decrypt API
			res = call(Unirest.post(SmartKeyPlugin.API_BASE_URL + DECRYPT_PATH.replace("{{kid}}", kid)).headers(new HashMap<String, String>() {
				{
					put("Authorization", TOKEN);
				}
			}).body("{" + 
					"  \"alg\": \"AES\"," + 
					"  \"cipher\": \"" + cipher + "\"," + 
					"  \"mode\": \"CFB\"," + 
					"  \"iv\": \"" + iv + "\"" + 
					"}"), InputStream.class);
			
			if (res.getStatus() != 200) {
				throw new IOException("DECRYPT_PATH - Status Code: " + res.getStatus() + ", Body: " + responseToString(res));
			}
			json = new JsonNode(responseToString(res)).getObject();
			
			String plain = json.getString("plain");
			assertTrue("The encryption text is not equal to decrypted text", plain.equals(base64)); // Verify correct decryption
		} catch (Exception e) {
			throw e;
		} finally {
			if (kid != null) {
				// Delete AES Key
				HttpResponse<InputStream> res = call(Unirest.delete(SmartKeyPlugin.API_BASE_URL + DELETE_KID_PATH.replace("{{kid}}", kid)).headers(new HashMap<String, String>() {
					{
						put("Authorization", TOKEN);
					}
				}), InputStream.class);
				if (res.getStatus() != 204) {
					throw new IOException("DELETE_KID_PATH - Status Code: " + res.getStatus() + ", Body: " + responseToString(res));
				}
			}
		}
		return new String[] {kid, iv};
	}

	private static <T> HttpResponse<T> call(BaseRequest rest, Class<T> responseClass) throws UnirestException {
		HttpRequest request = rest.getHttpRequest();
		HttpRequestBase requestObj = prepareRequest(request);
		HttpClient client = ClientFactory.getHttpClient(); // The
		org.apache.http.HttpResponse response;
		try {
			response = client.execute(requestObj);
			HttpResponse<T> httpResponse = new HttpResponse<T>(response, responseClass);
			requestObj.releaseConnection();
			return httpResponse;
		} catch (Exception e) {
			throw new UnirestException(e);
		} finally {
			requestObj.releaseConnection();
		}
	}

	private static HttpRequestBase prepareRequest(HttpRequest request) {

		Object defaultHeaders = Options.getOption(Option.DEFAULT_HEADERS);
		if (defaultHeaders != null) {
			@SuppressWarnings("unchecked")
			Set<Entry<String, String>> entrySet = ((Map<String, String>) defaultHeaders).entrySet();
			for (Entry<String, String> entry : entrySet) {
				request.header(entry.getKey(), entry.getValue());
			}
		}
		request.header("content-type", "application/json");

		HttpRequestBase reqObj = null;

		String urlToRequest = null;
		try {
			URL url = new URL(request.getUrl());
			URI uri = new URI(url.getProtocol(), url.getUserInfo(), url.getHost(), url.getPort(),
					URLDecoder.decode(url.getPath(), "UTF-8"), "", url.getRef());
			urlToRequest = uri.toURL().toString();
			if (url.getQuery() != null && !url.getQuery().trim().equals("")) {
				if (!urlToRequest.substring(urlToRequest.length() - 1).equals("?")) {
					urlToRequest += "?";
				}
				urlToRequest += url.getQuery();
			} else if (urlToRequest.substring(urlToRequest.length() - 1).equals("?")) {
				urlToRequest = urlToRequest.substring(0, urlToRequest.length() - 1);
			}
		} catch (Exception e) {
			throw new RuntimeException(e);
		}

		switch (request.getHttpMethod()) {
			case GET:
				reqObj = new HttpGet(urlToRequest);
				break;
			case POST:
				reqObj = new HttpPost(urlToRequest);
				break;
			case PUT:
				reqObj = new HttpPut(urlToRequest);
				break;
			case DELETE:
				reqObj = new HttpWithBody("DELETE", urlToRequest);
				break;
			case PATCH:
				reqObj = new HttpWithBody("PATCH", urlToRequest);
				break;
			case OPTIONS:
				reqObj = new HttpWithBody("OPTIONS", urlToRequest);
				break;
			case HEAD:
				reqObj = new HttpHead(urlToRequest);
				break;
		}

		Set<Entry<String, List<String>>> entrySet = request.getHeaders().entrySet();
		for (Entry<String, List<String>> entry : entrySet) {
			List<String> values = entry.getValue();
			if (values != null) {
				for (String value : values) {
					reqObj.addHeader(entry.getKey(), value);
				}
			}
		}

		// Set body
		if (!(request.getHttpMethod() == HttpMethod.GET || request.getHttpMethod() == HttpMethod.HEAD)) {
			if (request.getBody() != null) {
				HttpEntity entity = request.getBody().getEntity();
				((HttpEntityEnclosingRequestBase) reqObj).setEntity(entity);
			}
		}

		return reqObj;
	}
	
	public static String responseToString(HttpResponse<InputStream> res) throws IOException {
		return new String(ResponseUtils.getBytes(res.getBody()), "UTF-8");
	}
}

class HttpWithBody extends HttpEntityEnclosingRequestBase {

	private String method;
	
	public HttpWithBody() {
		super();
	}
	
	public HttpWithBody(final String uri) {
		super();
		setURI(URI.create(uri));
	}

	public HttpWithBody(final URI uri) {
		super();
		setURI(uri);
	}

	public HttpWithBody(String method, final String uri) {
		this(uri);
		this.method = method;
	}

	public String getMethod() {
		return method;
	}
}
