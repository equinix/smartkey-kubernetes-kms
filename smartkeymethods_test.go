package main

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestEncrypt(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://www.smartkey.io/crypto/v1/keys/uuid1/encrypt",
		httpmock.NewStringResponder(200, `{"kid": "1", "cipher": "cipher", "iv":"iv"}`))

	config := make(map[string]string)
	config["smartkeyURL"] = "https://www.smartkey.io"
	config["encryptionKeyUuid"] = "uuid1"
	config["smartkeyApiKey"] = "api_key"

	resp, err := encrypt(config, "plain")

	if err != nil || len(resp) <= 0 {
		t.Error("Encryption test case failed")
	}
}

func TestDecrypt(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://www.smartkey.io/crypto/v1/keys/uuid1/decrypt",
		httpmock.NewStringResponder(200, `{"kid": "1", "plain": "plain", "iv":"iv"}`))

	config := make(map[string]string)
	config["smartkeyURL"] = "https://www.smartkey.io"
	config["encryptionKeyUuid"] = "uuid1"
	config["smartkeyApiKey"] = "api_key"

	resp, err := decrypt(config, "cipher")
	if err != nil || len(resp) <= 0 {
		t.Error("Decryption test case failed")
	}

}
