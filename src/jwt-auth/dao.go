package main

//User struct
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

//Users struct
type Users struct {
	Users []User `json:"users"`
}

// AccessDetails struct
type AccessDetails struct {
	AccessUUID string
	UserID     int64
}

//TokenDetails struct
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

//Response details struct
type Response struct {
	UserID int64  `json:"user_id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}
