package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/twinj/uuid"
)

var r = gin.Default()
var prop map[string]string

var client *redis.Client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func main() {
	prop = LoadConfig()
	r.POST("/token", Token)
	r.POST("/authenticate", Authenticate)
	log.Fatal(r.Run(":9090"))
}

//Token function
func Token(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
		return
	}
	var validUser bool = ValidateUser(u)
	if !validUser {
		c.JSON(http.StatusUnauthorized, "Please provide valid login details")
		return
	}
	ts, err := GenerateToken(u.ID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	saveErr := SaveToken(u.ID, ts)
	if saveErr != nil {
		c.JSON(http.StatusUnprocessableEntity, saveErr.Error())
	}
	tokens := map[string]string{
		"access_token": ts.AccessToken,
	}
	c.JSON(http.StatusOK, tokens)
}

//GenerateToken function generates the token based on the access_secret and the userdata
func GenerateToken(userid int64) (*TokenDetails, error) {
	var err error
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 5).Unix()
	td.AccessUUID = uuid.NewV4().String()

	//Creating Access Token
	os.Setenv(accessSecret, prop[accessSecret]) //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv(accessSecret)))
	if err != nil {
		return nil, err
	}
	return td, nil
}

//SaveToken function is used to store the token against the id passed in the request
func SaveToken(userid int64, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	now := time.Now()

	errAccess := client.Set(td.AccessUUID, strconv.Itoa(int(userid)), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	return nil
}

//Authenticate fucntion which validates the token and returns the API response
func Authenticate(c *gin.Context) {
	var td Response
	var userid int64
	if err := c.ShouldBindJSON(&td); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json")
		return
	}
	//Extract the access token metadata
	metadata, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	userid, err = AuthUserID(metadata)
	if err != nil {
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}
	td.UserID = userid
	td.Status = "success"
	c.JSON(http.StatusCreated, td)
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

//TokenValid function verifies the token
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
