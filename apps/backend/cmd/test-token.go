package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Same secret as .env
	secret := "b1425f24750431e96173380f85395042ba41a3793a10d5029396477172e1b1cb"

	// Test user data
	userID := "a0c2764a-033c-424b-ab35-fc0862fb6a1e"
	email := "test@student.unand.ac.id"

	// Match the exact Claims structure from internal/auth/jwt.go
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		"iat":     jwt.NewNumericDate(time.Now()),
		"iss":     "podland",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(tokenString)
	fmt.Println()
	fmt.Println("Usage: curl -H \"Authorization: Bearer", tokenString, "\" http://localhost:8080/api/users/me")
}
