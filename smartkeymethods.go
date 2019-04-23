package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type EncryptResponse struct {
	Kid    string
	Cipher string
	Iv     string
}

type DecryptResponse struct {
	Kid   string
	Plain string
	Iv    string
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
func encrypt(config map[string]string, input string) string {
	/* Convert plain text to base64 */
	var base64Input = base64.StdEncoding.EncodeToString([]byte(input))
	encryptURL := config["smartkeyURL"] + "/crypto/v1/keys/" + config["encryptionKeyUuid"] + "/encrypt"
	log.Println("encrypt: map: %v", config)
	log.Println("encrypt: encryptURL:", encryptURL)
	/* Call SmartKey encrypt */
	var body, err = execute(config["smartkeyApiKey"], encryptURL, []byte(`{
		"alg":   "AES",
		"mode":  "CBC",
		"iv":    "`+config["iv"]+`",
		"plain": "`+base64Input+`"
	}`))

	if err != nil {
		log.Fatal("Error reading body. ", err)
	}

	var response EncryptResponse
	json.Unmarshal([]byte(body), &response)

	return response.Cipher
}

/* This is a method for calling decryption operation. */
func decrypt(config map[string]string, cipher string) string {
	decryptURL := config["smartkeyURL"] + "/crypto/v1/keys/" + config["encryptionKeyUuid"] + "/decryptURL"
	var body, err = execute(config["smartkeyApiKey"], decryptURL, []byte(`{
		"alg":   "AES",
		"mode":  "CBC",
		"iv":    "`+config["iv"]+`",
		"cipher": "`+cipher+`"
	}`))

	if err != nil {
		log.Fatal("Error reading body. ", err)
	}

	var response DecryptResponse
	json.Unmarshal([]byte(body), &response)

	var base64Input, _ = base64.StdEncoding.DecodeString(response.Plain)

	return string(base64Input)
}
