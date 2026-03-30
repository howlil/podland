package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Same secret as .env
	secret := "b1425f24750431e96173380f85395042ba41a3793a10d5029396477172e1b1cb"

	// Test user ID from database
	userID := "a0c2764a-033c-424b-ab35-fc0862fb6a1e"

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(tokenString)
}
