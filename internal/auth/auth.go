package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "mini-url",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			Subject:   userID.String(),
		},
	)

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		},
	)

	if err != nil {
		return uuid.UUID{}, err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	return uuid.Parse(id)
}

func GetBearerToken(store *sessions.CookieStore, r *http.Request) (string, error) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return "", err
	}

	if session.IsNew {
		return "", errors.New("no session found")
	}

	token, ok := session.Values["jwt"].(string)
	if !ok {
		return "", errors.New("no token found")
	}

	return token, nil
}

func SetBearerToken(store *sessions.CookieStore, w http.ResponseWriter, r *http.Request, token string) error {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return err
	}

	session.Values["jwt"] = token
	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}
