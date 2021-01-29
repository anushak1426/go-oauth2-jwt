package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/magiconair/properties"
)

const accessSecret string = "ACCESS_SECRET"
const refershSecret string = "REFRESH_SECRET"

//ValidateUser parse the users.json to return  a boolean value if authorized
func ValidateUser(u User) bool {
	var users Users
	file, _ := ioutil.ReadFile("users.json")
	_ = json.Unmarshal([]byte(file), &users)
	for _, element := range users.Users {
		if element.Username == u.Username && element.Password == u.Password {
			return true
		}
	}
	return false
}

//LoadConfig function loads the properties and convert to map
func LoadConfig() map[string]string {
	var values = []string{"${GOPATH}\\src\\jwt-auth\\config.properties"}
	var P, _ = properties.LoadFiles(values, properties.UTF8, true)
	accessSecretValue := P.MustGet(accessSecret)
	refreshSecretValue := P.MustGet(refershSecret)
	return map[string]string{
		accessSecret:  accessSecretValue,
		refershSecret: refreshSecretValue,
	}
}
