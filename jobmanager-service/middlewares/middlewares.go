/*
Copyright 2023-2024 Bull SAS

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package middlewares

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"etsn/server/jobmanager-service/responses"
	"etsn/server/jobmanager-service/utils/logs"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	base64EncodedPublicKey = os.Getenv("KEYCLOAK_PUBLIC_KEY")
)

func SetMiddlewareJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

func SetMiddlewareLog(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logHttpCall(r.Method + " " + r.URL.String())
		next(w, r)
	}
}

func JWTValidation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if base64EncodedPublicKey == "" {
			fmt.Println("No Authentication required")
			next(w, r)
			return
		}

		tokenString := r.Header.Get("Authorization")
		splitToken := strings.Split(tokenString, "Bearer")
		if len(splitToken) < 2 {
			err := errors.New("not authorized")
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
		reqToken := splitToken[1]
		reqToken = strings.TrimSpace(reqToken)

		publicKey, err := parseKeycloakRSAPublicKey(base64EncodedPublicKey)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}

		token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// return the public key that is used to validate the token.
			return publicKey, nil
		})
		if err != nil {
			// fmt.Println("Error parsing or validating token:", err)
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}

		if !token.Valid {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		logs.Logger.Println("Claims:", claims)
		next(w, r)
	}
}

func parseKeycloakRSAPublicKey(base64Encoded string) (*rsa.PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return nil, err
	}
	parsedKey, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, err
	}
	publicKey, ok := parsedKey.(*rsa.PublicKey)
	if ok {
		return publicKey, nil
	}
	return nil, fmt.Errorf("unexpected key type %T", publicKey)
}

func logHttpCall(format string, args ...interface{}) {
	logs.Logger.Printf(time.Now().Format(time.RFC3339)+"  \x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}
