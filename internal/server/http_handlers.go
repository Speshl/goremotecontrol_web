package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var tempSecretKey = []byte("TempSecretKey") //TODO: Load from env variable
var tempUser = "username"
var tempPass = "password"

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (s *Server) RegisterHTTPHandlers() {
	http.HandleFunc("/authed", s.authedHandler)
	http.HandleFunc("/login", s.loginHandler)
	http.HandleFunc("/preauth", s.preAuthHandler)
	http.Handle("/", http.FileServer(http.Dir("public/")))
	http.Handle("/socket.io/", s.socketio)

}

func (s *Server) preAuthHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("public/login.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}

func (s *Server) loginHandler(w http.ResponseWriter, req *http.Request) {
	var creds Credentials
	err := json.NewDecoder(req.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.validateCredentials(creds)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenString, err := s.generateJWT()
	if err != nil {
		log.Println("Error generating JWT", err)
		return
	}

	w.Header().Set("Token", tokenString)
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(5 * time.Minute),
	})
	template := template.Must(template.ParseFiles("public/login.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}

func (s *Server) authedHandler(w http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenString := cookie.Value
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return tempSecretKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	template := template.Must(template.ParseFiles("public/welcome.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}

func (s *Server) validateCredentials(creds Credentials) error {
	if creds.Username != tempUser && creds.Password != tempPass {
		return fmt.Errorf("invalid username and password")
	}
	return nil
}

/*********************************JWT******************************/

func (s *Server) generateJWT() (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: tempUser,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(tempSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed using secret key: %w", err)
	}
	return tokenString, nil
}

// func (s *Server) verifyJWT(endpointHandler func(writer http.ResponseWriter, request *http.Request)) http.HandlerFunc {
// 	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
// 		if request.Header["Token"] != nil {
// 			token, err := jwt.Parse(request.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
// 				_, ok := token.Method.(*jwt.SigningMethodECDSA)
// 				if !ok {
// 					writer.WriteHeader(http.StatusUnauthorized)
// 					_, err := writer.Write([]byte("You're Unauthorized"))
// 					if err != nil {
// 						return nil, err
// 					}
// 				}
// 				return "", nil

// 			})
// 			// parsing errors result
// 			if err != nil {
// 				writer.WriteHeader(http.StatusUnauthorized)
// 				_, err2 := writer.Write([]byte("You're Unauthorized due to error parsing the JWT"))
// 				if err2 != nil {
// 					return
// 				}

// 			}
// 			// if there's a token
// 			if token.Valid {
// 				endpointHandler(writer, request)
// 			} else {
// 				writer.WriteHeader(http.StatusUnauthorized)
// 				_, err := writer.Write([]byte("You're Unauthorized due to invalid token"))
// 				if err != nil {
// 					return
// 				}
// 			}
// 		} else {
// 			writer.WriteHeader(http.StatusUnauthorized)
// 			_, err := writer.Write([]byte("You're Unauthorized due to No token in the header"))
// 			if err != nil {
// 				return
// 			}
// 		}
// 		// response for if there's no token header
// 	})
// }
