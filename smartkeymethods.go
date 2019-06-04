package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

/*EncryptResponse response from SmartKey for encrypt API Call*/
type EncryptResponse struct {
	Kid    string
	Cipher string
	Iv     string
}

/*DecryptResponse response from SmartKey for decrypt API Call*/
type DecryptResponse struct {
	Kid   string
	Plain string
	Iv    string
}

/*KeyObject response from SmartKey for decrypt API Call*/
type KeyObject struct {
	// For objects which are not elliptic curves, this is the size in bits (not bytes) of the object. This field is not returned for elliptic curves.
	KeySize int32  `json:"key_size,omitempty"`
	ObjType string `json:"obj_type"`
}

/* This function calls actual SmartKey REST APIs on a SmartKey API endpoint. */
func execute(apikey string, url string, data []byte) ([]byte, error) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+apikey)

	client := &http.Client{}

	/* Call SmartKey API to perform operation */
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(req)
		log.Fatal("Error reading response. ", err)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

/* This is a method for calling encryption operation. */
func encrypt(config map[string]string, input string) (string, error) {
	/* Convert plain text to base64 */
	var base64Input = base64.StdEncoding.EncodeToString([]byte(input))
	encryptURL := config["smartkeyURL"] + "/crypto/v1/keys/" + config["encryptionKeyUuid"] + "/encrypt"
	log.Println("encrypt: map:", config)
	log.Println("encrypt: encryptURL:", encryptURL)

	/* Call SmartKey encrypt */
	var body, err = execute(config["smartkeyApiKey"], encryptURL, []byte(`{
		"alg":   "AES",
		"mode":  "CBC",
		"iv":    "`+config["iv"]+`",
		"plain": "`+base64Input+`"
	}`))

	if err != nil {
		log.Print("Error reading body. ", err)
		return "", err
	}

	var response EncryptResponse
	json.Unmarshal([]byte(body), &response)

	return response.Cipher, nil
}

/* This is a method for calling decryption operation. */
func decrypt(config map[string]string, cipher string) (string, error) {
	decryptURL := config["smartkeyURL"] + "/crypto/v1/keys/" + config["encryptionKeyUuid"] + "/decrypt"
	log.Println("decrypt: map:", config)
	log.Println("decrypt: encryptURL:", decryptURL)

	/* Call SmartKey decrypt */
	var body, err = execute(config["smartkeyApiKey"], decryptURL, []byte(`{
		"alg":   "AES",
		"mode":  "CBC",
		"iv":    "`+config["iv"]+`",
		"cipher": "`+cipher+`"
	}`))

	if err != nil {
		log.Print("Error reading body. ", err)
		return "", err
	}

	var response DecryptResponse
	json.Unmarshal([]byte(body), &response)

	var base64Input, _ = base64.StdEncoding.DecodeString(response.Plain)

	return string(base64Input), nil
}

/* This is a method for calling authentication operation. */
func auth(config map[string]string) (string, error) {
	/* Convert plain text to base64 */
	authURL := config["smartkeyURL"] + "/sys/v1/session/auth"

	/* Call SmartKey auth */
	req, err := http.NewRequest("POST", authURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+config["smartkeyApiKey"])

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil || resp.StatusCode != 200 {
		return "", errors.New("authentication failed")
	}

	return "", nil
}

/* This is a method for validating security object based on key uuid */
func validateKey(config map[string]string) (string, error) {
	/* Convert plain text to base64 */
	authURL := config["smartkeyURL"] + "/crypto/v1/keys/" + config["encryptionKeyUuid"]

	/* Call SmartKey get security object */
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+config["smartkeyApiKey"])

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		return "", errors.New("encryption key validation failed")
	}

	var keyResponse KeyObject
	if err := json.NewDecoder(resp.Body).Decode(&keyResponse); err != nil {
		return "", errors.New("encryption key validation failed")
	}

	if keyResponse.ObjType != "AES" || keyResponse.KeySize != 256 {
		return "", errors.New("encryption key validation failed")
	}

	return "", nil
}
