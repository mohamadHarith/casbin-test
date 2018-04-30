package main

import (
    "strings"
    "strconv"
    "fmt"
    "flag"
    "log"
    "io/ioutil"
    "net/http"
    "encoding/json"

    "github.com/dgrijalva/jwt-go"
    "github.com/casbin/casbin"
    "github.com/gorilla/mux"
)

type Exception struct {
    Message string `json:"message"`
}

type Greeting struct {
    Message string `json:"message"`
}

type User struct {
    email string `json:"email"`
    source string `json:"iss"`
}

type authError struct {
    header string
    prob string
}

func (e *authError) Error() string {
    return fmt.Sprintf("%s - s", e.header, e.prob)
}

var secret_num int
var policyEngine = casbin.NewEnforcer("model.conf", "policy.csv")

func userFromHeader(auth_header string) (User, error) {
    var user User
    if auth_header == "" {
        return user, &authError{auth_header, "Absent header"}
    }

    bearerToken := strings.Split(auth_header, " ")
    if len(bearerToken) != 2 {
        return user, &authError{auth_header, "Invalid bearer token"}
    }

    token, _ := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
        log.Print("Got token ", token)
        return []byte("secret"), nil
    })

    if claims, ok := token.Claims.(jwt.MapClaims); ok {
        user.email = claims["email"].(string)
        user.source = claims["iss"].(string)
    } else {
        return user, &authError{auth_header, "Not an OIDC token"}
    }

    return user, nil
}

func Authorizer(e *casbin.Enforcer) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        fn := func(w http.ResponseWriter, req *http.Request) {
            user, ok := userFromHeader(req.Header.Get("authorization"))
            if ok != nil {
                w.WriteHeader(http.StatusUnauthorized)
                return
            }

            res, err := e.EnforceSafe(user.source, req.URL.Path, req.Method)
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

func GetSecret(w http.ResponseWriter, req *http.Request) {
    json.NewEncoder(w).Encode(secret_num)
}

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

    secret_num = tryNum
    json.NewEncoder(w).Encode(secret_num)
}

func main() {
    var port = flag.Int("port", 8080, "port to listen on")
	versionFlag := flag.Bool("version", false, "Version")
	flag.Parse()

	if *versionFlag {
		fmt.Println("Git Commit:", GitCommit)
		fmt.Println("Version:", Version)
		if VersionPrerelease != "" {
			fmt.Println("Version PreRelease:", VersionPrerelease)
		}
		return
	}

    secret_num = 42;

    router := mux.NewRouter()
    fmt.Println("Starting the application...")
    fmt.Println(*port)

    router.HandleFunc("/secret_number", GetSecret).Methods("GET")
    router.HandleFunc("/secret_number", PutSecret).Methods("PUT")

    log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), Authorizer(policyEngine)(router)))
}
