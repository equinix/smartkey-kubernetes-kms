package main

import (
	"github.com/jarcoal/httpmock"
	"io/ioutil"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	config := make(map[string]string)
	_, err := New("/path/to/sock/file", config)

	if err != nil {
		t.Error(err)
	}
}

func TestVersion(t *testing.T) {
	config := make(map[string]string)
	serv, err := New("/path/to/sock/file", config)

	val, err := serv.Version(nil, nil)

	if err != nil {
		t.Error(err)
	}
	if val.Version != "v1beta1" || val.RuntimeName != "vault" || val.RuntimeVersion != "0.1.0" {
		t.Error("Invalid version info")
	}
}

func TestCleanSocketVersion(t *testing.T) {
	config := make(map[string]string)
	serv, err := New("/path/to/sock/file", config)

	if err != nil {
		t.Error(err)
	}

	errCloseSock := serv.cleanSockFile()

	if errCloseSock != nil {
		t.Error(errCloseSock)
	}
}

func TestParseConfigFile_Positive_ValidFile(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "www.smartkey.io/sys/v1/session/auth",
		httpmock.NewStringResponder(200, `{"expires_in": 0,"access_token": "","entity_id": ""}`))

	httpmock.RegisterResponder("GET", "www.smartkey.io/crypto/v1/keys/uuid-1",
		httpmock.NewStringResponder(200, `{"key_size": 256, "obj_type": "AES"}`))

	configData := []byte("{\"" +
		"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"rFvgbU6EygpLUObqFZxITg==\"," +
		"\"socketFile\": \"unix-sockfile-path\"," +
		"\"smartkeyURL\": \"www.smartkey.io\"" +
		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err != nil {
		t.Error(err)
	}
}

func TestParseConfigFile_Negative_ApiKeyMissing(t *testing.T) {
	configData := []byte("{\"" +
		//"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"rFvgbU6EygpLUObqFZxITg==\"," +
		"\"socketFile\": \"unix-sockfile-path\"," +
		"\"smartkeyURL\": \"www.smartkey.io\"" +
		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err == nil {
		t.Error("Test case should fail as [smartkeyApiKey] is missing")
	}
}

func TestParseConfigFile_Negative_UuidMissing(t *testing.T) {
	configData := []byte("{\"" +
		"smartkeyApiKey\": \"your-api-key\"," +
		//"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"rFvgbU6EygpLUObqFZxITg==\"," +
		"\"socketFile\": \"unix-sockfile-path\"," +
		"\"smartkeyURL\": \"www.smartkey.io\"" +
		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err == nil {
		t.Error("Test case should fail as [encryptionKeyUuid] is missing")
	}
}

func TestParseConfigFile_Negative_IvMissing(t *testing.T) {
	configData := []byte("{\"" +
		"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		//"\"iv\": \"rFvgbU6EygpLUObqFZxITg==\"," +
		"\"socketFile\": \"unix-sockfile-path\"," +
		"\"smartkeyURL\": \"www.smartkey.io\"" +
		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err == nil {
		t.Error("Test case should fail as [iv] is missing")
	}
}

func TestParseConfigFile_Negative_SocketFileMissing(t *testing.T) {
	configData := []byte("{\"" +
		"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"rFvgbU6EygpLUObqFZxITg==\"," +
		//"\"socketFile\": \"unix-sockfile-path\"," +
		"\"smartkeyURL\": \"www.smartkey.io\"" +

		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err == nil {
		t.Error("Test case should fail as [socketFile] is missing")
	}
}

func TestParseConfigFile_Negative_SmartkeyURLMissing(t *testing.T) {
	configData := []byte("{\"" +
		"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"rFvgbU6EygpLUObqFZxITg==\"," +
		"\"socketFile\": \"unix-sockfile-path\"," +
		//"\"smartkeyURL\": \"www.smartkey.io\"" +
		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err == nil {
		t.Error("Test case should fail as [smartkeyURL] is missing")
	}
}

func TestParseConfigFile_Negative_ApiKeyInvalid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "www.smartkey.io/sys/v1/session/auth",
		httpmock.NewStringResponder(400, `{"expires_in": 0,"access_token": "","entity_id": ""}`))

	httpmock.RegisterResponder("GET", "www.smartkey.io/crypto/v1/keys/uuid-1",
		httpmock.NewStringResponder(200, `{"key_size": 256, "obj_type": "AES"}`))

	configData := []byte("{\"" +
		"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"iv-1\"," +
		"\"socketFile\": \"unix-sockfile-path\"," +
		"\"smartkeyURL\": \"www.smartkey.io\"" +
		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err == nil {
		t.Error("Test case should fail as [smartkeyApiKey] is invalid")
	}
}

func TestParseConfigFile_Negative_KeyInvalid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "www.smartkey.io/sys/v1/session/auth",
		httpmock.NewStringResponder(200, `{"expires_in": 0,"access_token": "","entity_id": ""}`))

	httpmock.RegisterResponder("GET", "www.smartkey.io/crypto/v1/keys/uuid-1",
		httpmock.NewStringResponder(200, `{"key_size": 128, "obj_type": "AES"}`))

	configData := []byte("{\"" +
		"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"iv-1\"," +
		"\"socketFile\": \"unix-sockfile-path\"," +
		"\"smartkeyURL\": \"www.smartkey.io\"" +
		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err == nil {
		t.Error("Test case should fail as [encryptionKeyUuid] is invalid")
	}
}

func TestParseConfigFile_Negative_IvInvalid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "www.smartkey.io/sys/v1/session/auth",
		httpmock.NewStringResponder(200, `{"expires_in": 0,"access_token": "","entity_id": ""}`))

	httpmock.RegisterResponder("GET", "www.smartkey.io/crypto/v1/keys/uuid-1",
		httpmock.NewStringResponder(200, `{"key_size": 256, "obj_type": "AES"}`))

	configData := []byte("{\"" +
		"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"iv-1\"," +
		"\"socketFile\": \"unix-sockfile-path\"," +
		"\"smartkeyURL\": \"www.smartkey.io\"" +
		"}")
	ioutil.WriteFile("smartkey-grpc_tmp.conf", configData, 0644)

	_, err := parseConfigFile("smartkey-grpc_tmp.conf")

	os.Remove("smartkey-grpc_tmp.conf")

	if err == nil {
		t.Error("Test case should fail as [encryptionKeyUuid] is invalid")
	}
}
