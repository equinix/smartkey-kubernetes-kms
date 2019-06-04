package main

import (
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
	configData := []byte("{\"" +
		"smartkeyApiKey\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"," +
		"\"encryptionKeyUuid\": \"123e4567-e89b-4123-9123-123456780000\"," +
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
		//"smartkeyApiKey\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"," +
		"\"encryptionKeyUuid\": \"123e4567-e89b-4123-9123-123456780000\"," +
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
		"smartkeyApiKey\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"," +
		//"\"encryptionKeyUuid\": \"123e4567-e89b-4123-9123-123456780000\"," +
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
		"smartkeyApiKey\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"," +
		"\"encryptionKeyUuid\": \"123e4567-e89b-4123-9123-123456780000\"," +
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
		"smartkeyApiKey\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"," +
		"\"encryptionKeyUuid\": \"123e4567-e89b-4123-9123-123456780000\"," +
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
		"smartkeyApiKey\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"," +
		"\"encryptionKeyUuid\": \"123e4567-e89b-4123-9123-123456780000\"," +
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
	configData := []byte("{\"" +
		"smartkeyApiKey\": \"ApiKey\"," +
		"\"encryptionKeyUuid\": \"123e4567-e89b-4123-9123-123456780000\"," +
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

func TestParseConfigFile_Negative_KeyUuiInvalid(t *testing.T) {
	configData := []byte("{\"" +
		"smartkeyApiKey\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"," +
		"\"encryptionKeyUuid\": \"uuid\"," +
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
	configData := []byte("{\"" +
		"smartkeyApiKey\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"," +
		"\"encryptionKeyUuid\": \"123e4567-e89b-4123-9123-123456780000\"," +
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
