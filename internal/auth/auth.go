package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func GetBearerToken(headers http.Header) (string, error){
	authHeader:= headers.Get("Authorization")
	if strings.HasPrefix(authHeader,"Bearer "){
		token := authHeader[len("Bearer "):]
		return token, nil
	}else{
		return "", fmt.Errorf("invalid authorization header")
	}
}

func HashPassword(password string) (string, error){
	hash, err := bcrypt.GenerateFromPassword([]byte(password,),bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password, hash string) error{
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error)  {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
	})
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil{
		return "", err
	}

	return signedToken, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error){
	var userid uuid.UUID
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userid,err = uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid user ID format: %v", err)
		}
	} else {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	return userid, nil
}