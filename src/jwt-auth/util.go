package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
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

//ExtractTokenMetadata function from the request and cache
func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUUID, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &AccessDetails{
			AccessUUID: accessUUID,
			UserID:     int64(userID),
		}, nil
	}
	return nil, err
}

//AuthUserID function validates the id used while generating the token and the one passed while invoking the API
func AuthUserID(authD *AccessDetails) (int64, error) {
	userid, err := client.Get(authD.AccessUUID).Result()
	if err != nil {
		return 0, err
	}
	userID, _ := strconv.ParseUint(userid, 10, 64)
	if uint64(authD.UserID) != userID {
		return 0, errors.New("unauthorized")
	}
	return int64(userID), nil
}

//ExtractToken function extracts the token passed in the request header
func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// VerifyToken function Parse, validate, and return a token.keyFunc will receive the parsed token and should return the key for validating.
func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv(accessSecret)), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

//TokenValid function verifies the toke
func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return err
	}
	return nil
}
