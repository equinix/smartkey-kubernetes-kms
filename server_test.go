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
		"smartkeyApiKey\": \"your-api-key\"," +
		"\"encryptionKeyUuid\": \"uuid-1\"," +
		"\"iv\": \"iv-1\"," +
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
		"\"iv\": \"iv-1\"," +
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
		"\"iv\": \"iv-1\"," +
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
		//"\"iv\": \"iv-1\"," +
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
		"\"iv\": \"iv-1\"," +
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
		"\"iv\": \"iv-1\"," +
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

// func TestEncrypt(t *testing.T) {

// 	config := make(map[string]string)
// 	serv, err := New("/path/to/sock/file", config)

// 	serv.Encrypt(nil, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestParseCmd(t *testing.T) {
// 	typeFlag = "text"
// 	flag.String("socketFile1", "/sock/file/path", "socket file that gRpc server listens to")
// 	flag.String("config1", "config file path", "config file location")
// 	//flag.Var(&typeFlag, "f", usage+" (shorthand)")
// 	_, err := parseCmd()
// 	log.Fatal(err)
// }
