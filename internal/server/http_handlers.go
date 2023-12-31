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
	http.HandleFunc("/index", s.indexHandler)
	http.HandleFunc("/login", s.loginHandler)

	//auth testing
	http.HandleFunc("/authed", s.authedHandler)
	http.HandleFunc("/signin", s.signinHandler)
	http.HandleFunc("/preauth", s.preAuthHandler)

	//serves js and static html
	http.Handle("/", http.FileServer(http.Dir("public/")))
	//sets up socket connections for video/commands
	http.Handle("/socket.io/", s.socketio)

}

func (s *Server) indexHandler(w http.ResponseWriter, req *http.Request) {
	s.buildIndex(w, IndexBuildOptions{
		includeShell: true,
		authorized:   false,
	})
}

func (s *Server) loginHandler(w http.ResponseWriter, req *http.Request) {
	var creds Credentials
	err := json.NewDecoder(req.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		log.Printf("error decoding json body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.validateCredentials(creds)
	if err != nil {
		log.Printf("error decoding credentials: %s", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenString, err := s.generateJWT()
	if err != nil {
		log.Printf("Error generating JWT: %s", err.Error())
		return
	}

	w.Header().Set("Token", tokenString)
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(5 * time.Minute),
	})

	s.buildIndex(w, IndexBuildOptions{
		includeShell: false,
		authorized:   true,
		userName:     creds.Username,
		userRank:     "69420",
	})
}

/*--------------------------Auth Testing-----------------------------*/
func (s *Server) preAuthHandler(w http.ResponseWriter, req *http.Request) {
	template := template.Must(template.ParseFiles("public/login.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}

func (s *Server) signinHandler(w http.ResponseWriter, req *http.Request) {
	var creds Credentials
	err := json.NewDecoder(req.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		log.Printf("error decoding json body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.validateCredentials(creds)
	if err != nil {
		log.Printf("error decoding credentials: %s", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenString, err := s.generateJWT()
	if err != nil {
		log.Printf("Error generating JWT: %s", err.Error())
		return
	}

	w.Header().Set("Token", tokenString)
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(5 * time.Minute),
	})
	template := template.Must(template.ParseFiles("public/welcome.html"))
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
