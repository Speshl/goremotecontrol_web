package server

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var sampleSecretKey = []byte("SecretYouShouldHide") //TODO: Load from env variable

func (s *Server) RegisterHTTPHandlers() {
	http.HandleFunc("/authed", s.verifyJWT(s.authedHandler))
	http.HandleFunc("/login", s.loginHandler)
	http.Handle("/", http.FileServer(http.Dir("public/")))
	http.Handle("/socket.io/", s.socketio)

}

func (s *Server) loginHandler(w http.ResponseWriter, req *http.Request) {

	_, err := s.generateJWT()
	if err != nil {
		log.Fatalln("Error generating JWT", err)
	}

	w.Header().Set("Token", "%v")
	template := template.Must(template.ParseFiles("public/login.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}

func (s *Server) authedHandler(w http.ResponseWriter, req *http.Request) {

	_, err := s.generateJWT()
	if err != nil {
		log.Fatalln("Error generating JWT", err)
	}

	w.Header().Set("Token", "%v")
	template := template.Must(template.ParseFiles("public/login.html"))
	template.Execute(w, nil) //Can pass map[string]any here and use go templates to dynamically build the html page
}

/*********************************JWT******************************/

func (s *Server) generateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodEdDSA)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(10 * time.Minute) //expiration 10 minutes
	claims["authorized"] = true
	claims["user"] = "username"

	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *Server) verifyJWT(endpointHandler func(writer http.ResponseWriter, request *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header["Token"] != nil {
			token, err := jwt.Parse(request.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				_, ok := token.Method.(*jwt.SigningMethodECDSA)
				if !ok {
					writer.WriteHeader(http.StatusUnauthorized)
					_, err := writer.Write([]byte("You're Unauthorized"))
					if err != nil {
						return nil, err
					}
				}
				return "", nil

			})
			// parsing errors result
			if err != nil {
				writer.WriteHeader(http.StatusUnauthorized)
				_, err2 := writer.Write([]byte("You're Unauthorized due to error parsing the JWT"))
				if err2 != nil {
					return
				}

			}
			// if there's a token
			if token.Valid {
				endpointHandler(writer, request)
			} else {
				writer.WriteHeader(http.StatusUnauthorized)
				_, err := writer.Write([]byte("You're Unauthorized due to invalid token"))
				if err != nil {
					return
				}
			}
		} else {
			writer.WriteHeader(http.StatusUnauthorized)
			_, err := writer.Write([]byte("You're Unauthorized due to No token in the header"))
			if err != nil {
				return
			}
		}
		// response for if there's no token header
	})
}
