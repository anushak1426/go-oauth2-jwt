package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type server struct{}

//Credentials to store credentials returned by server
type Credentials struct {
	ClientID     string
	ClientSecret string
}

//Token used to authenticate the client
type Token struct {
	AccessToken string `json:"access_token"`
	Expiry      int16  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var data []byte
	var credentials Credentials
	var token Token
	var message string
	response, err := http.Get("http://127.0.0.1:9096/credentials")
	if err != nil {
		fmt.Print("Failed to get credentials  , Pls verify the connectivity")
	} else {
		response, _ := ioutil.ReadAll(response.Body)
		data = response
	}
	json.Unmarshal(data, &credentials)
	response, err = http.Get("http://127.0.0.1:9096/token?grant_type=client_credentials&client_id=" + credentials.ClientID + "&client_secret=" + credentials.ClientSecret + "&scope=all")
	if err != nil {
		fmt.Print("Failed to get token  , Pls verify the credentials")
	} else {
		response, _ := ioutil.ReadAll(response.Body)
		data = response
	}
	json.Unmarshal(data, &token)
	var bearer = "Bearer " + token.AccessToken
	req, err := http.NewRequest("POST", "http://127.0.0.1:9096/authenticate", nil)
	req.Header.Add("Authorization", bearer)
	client := http.Client{}
	resp, err := client.Do(req)
	fmt.Println("Status", resp.Status)
	if resp.Status == "200 OK" {
		message = "{\"Status\":\"Success\"}"
	} else {
		message = "{\"Status\":\"Fail\"}"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

func main() {
	s := &server{}
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":9090", nil))
}
