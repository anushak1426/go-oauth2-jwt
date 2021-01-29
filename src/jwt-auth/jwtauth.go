package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
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
	r.POST("/authenticate", Authorize(), Authenticate)
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
	c.JSON(http.StatusCreated, tokens)
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

//Authorize function provides the secure access to the API's
func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := TokenValid(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}
		c.Next()
	}
}
