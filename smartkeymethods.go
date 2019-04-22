package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

/* XXX: Pleae enter apikey here before building the code. */
var apikey = "TBD"
var kid = "36c2a6d9-d222-4b45-90f8-fb58fa8ca3f3"
var iv = "tDdpYqC2YIoFNzwc3TfSeQ=="
var baseURL = "https://smartkey.io/crypto/v1/keys/" + kid
var encryptURL = baseURL + "/encrypt"
var decryptURL = baseURL + "/decrypt"

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
func execute(url string, data []byte) ([]byte, error) {
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
func encrypt(input string) string {
	/* Convert plain text to base64 */
	var base64Input = base64.StdEncoding.EncodeToString([]byte(input))

	/* Call SmartKey encrypt */
	var body, err = execute(encryptURL, []byte(`{
		"alg":   "AES",
		"mode":  "CBC",
		"iv":    "`+iv+`",
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
func decrypt(cipher string) string {
	var body, err = execute(decryptURL, []byte(`{
		"alg":   "AES",
		"mode":  "CBC",
		"iv":    "`+iv+`",
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

/*
func main() {
 	var plainText = "hello world"
 	fmt.Println("Test plainText used >> " + plainText)
 	var cipherText = encrypt(plainText)
 	fmt.Println("Cipher text generate from plainText [" + plainText + "] >> " + cipherText)
 	var convertedPlainText = decrypt(cipherText)
 	fmt.Println("Cipher text [" + cipherText + "] converted back to plainText >> " + convertedPlainText)
}
*/
