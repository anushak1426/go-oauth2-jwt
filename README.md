# go-oauth2-jwt
OAuth2 and JWT Authentication in Go


This tutorial illustrates the following
1. Generate token using random client credentials for client/server handshake
2. Authentication using JWT

Prerequisites:
1.GOROOT and GOPATH should be set properly to refer the packages/dependencies
2.Redis server 3.0 

OAuth2


JWT
1.Download the src from git 
2.Build the files under package jwt-auth using "go build -o jwt-auth.exe *.go"
3.Run the exe jwt-auth.exe 
4.Verify the results now

Get the access token using below curl command 
curl --location --request POST 'http://localhost:9090/token' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id":302,
    "username": "admin",
    "password": "test"
}'

output :
{
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImUyMzEyMTBmLWQyNWItNGYzYy05YzZhLWM0NzU4NDhhN2Q5MiIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMTg4NzQ3OCwidXNlcl9pZCI6MzAyfQ.Z98RKXr1bl0UwXGtmFufirimQtXDTMf0TCWRIcyy4Xk"
}

Note : the expiry of this token is 5 Min , afterwards will be automatically flushed from cache

Pass the access_token from the client program using below curl command(i.e /authenticate) 
curl --location --request POST 'http://localhost:9090/authenticate' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImUyMzEyMTBmLWQyNWItNGYzYy05YzZhLWM0NzU4NDhhN2Q5MiIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMTg4NzQ3OCwidXNlcl9pZCI6MzAyfQ.Z98RKXr1bl0UwXGtmFufirimQtXDTMf0TCWRIcyy4Xk' \
--header 'Content-Type: application/json' \
--data-raw '{
    "user_id":302,
    "title":"student"
}'

output:
{
    "user_id": 302,
    "title": "student",
    "status": "success"
}
