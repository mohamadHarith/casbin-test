package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	casbin "github.com/casbin/casbin"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

// Exception :
type Exception struct {
	Message string `json:"message"`
}

// Greeting :
type Greeting struct {
	Message string `json:"message"`
}

// User :
type User struct {
	Email string `json:"email"`
	Role  string `json:"sub"`
}

// authError :
type authError struct {
	header string
	prob   string
}

func (e *authError) Error() string {
	return fmt.Sprintf("%s - %s", e.header, e.prob)
}

var secretNum int

var policyEngine, _ = casbin.NewEnforcer("model.conf", "policy.csv")

// userFromHeader :
func userFromHeader(authHeader string) (User, error) {
	var user User
	if authHeader == "" {
		return user, &authError{authHeader, "Absent header"}
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 {
		return user, &authError{authHeader, "Invalid bearer token"}
	}

	token, _ := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
		log.Print("Got token ", token)
		return []byte("secret"), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		user.Email = claims["email"].(string)
		user.Role = claims["sub"].(string)
	} else {
		return user, &authError{authHeader, "Not an OIDC token"}
	}

	return user, nil
}

// Authorizer :
func Authorizer(e *casbin.Enforcer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			user, ok := userFromHeader(req.Header.Get("authorization"))
			if ok != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			res, err := e.Enforce(user.Role, req.URL.Path, req.Method)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !res {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, req)
		}
		return http.HandlerFunc(fn)
	}
}

// GetSecret :
func GetSecret(w http.ResponseWriter, req *http.Request) {
	fmt.Println("yes im here")
	json.NewEncoder(w).Encode(secretNum)
}

// PutSecret :
func PutSecret(w http.ResponseWriter, req *http.Request) {
	body, ok := ioutil.ReadAll(req.Body)
	if ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	lines := strings.Split(string(body), "\n")

	var tryNum int
	tryNum, ok = strconv.Atoi(lines[0])
	if ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	secretNum = tryNum
	json.NewEncoder(w).Encode(secretNum)
}

func main() {
	var port = flag.Int("port", 8080, "port to listen on")
	// versionFlag := flag.Bool("version", false, "Version")
	// flag.Parse()

	// if *versionFlag {
	// 	fmt.Println("Git Commit:", GitCommit)
	// 	fmt.Println("Version:", Version)
	// 	if VersionPrerelease != "" {
	// 		fmt.Println("Version PreRelease:", VersionPrerelease)
	// 	}
	// 	return
	// }

	secretNum = 42

	router := mux.NewRouter()
	fmt.Println("Starting the application...")
	fmt.Println(*port)

	router.HandleFunc("/secretNumber", GetSecret).Methods("GET")
	router.HandleFunc("/secretNumber", PutSecret).Methods("PUT")

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), Authorizer(policyEngine)(router)))
}
