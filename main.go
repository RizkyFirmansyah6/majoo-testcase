package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var mysqlDB *sql.DB

// User Struct (model)
type User struct {
	Id          int64  `jsonapi:"primary,user"`
	Username    string `jsonapi:"attr,Username"`
	Password    string `jsonapi:"attr,Password"`
	NamaLengkap string `jsonapi:"attr,Nama_Lengkap"`
	Foto        string `jsonapi:"attr,Foto"`
}

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func (user User) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("http://localhost:8080/api/user/%d", user.Id),
	}
}

func main() {
	// Init DB Connection
	mysqlDB = connect()
	defer mysqlDB.Close()

	// Init Router
	router := mux.NewRouter().StrictSlash(true)

	// Router Handlers / Endpoints
	router.HandleFunc("/api/user", getAllUser).Methods("GET")
	router.HandleFunc("/api/login", login).Methods("POST")
	router.HandleFunc("/api/add", addUser).Methods("POST")
	router.HandleFunc("/api/user/{id}", putUser).Methods("PUT")
	router.HandleFunc("/api/user/{id}", deleteUser).Methods("DELETE")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env("PORT", "8080")), router))
}

func putUser(writer http.ResponseWriter, request *http.Request) {
	userID := mux.Vars(request)["id"]
	imageName, err := FileUpload(request)
	if err != nil {
		http.Error(writer, "Invalid Data", http.StatusBadRequest)
		return
		//checking whether any error occurred retrieving image
	}
	writer.Header().Set("Content-Type", "application/json")
	var user User
	err = json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(writer, []*jsonapi.ErrorObject{{
			Title:  "ValidationError",
			Detail: "Given request is invalid",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
		}})
		return
	}
	query, err := mysqlDB.Prepare("UPDATE user SET Username = ?, Password = ?, Nama_Lengkap = ?, Foto = ? WHERE ID = ?")
	has := sha1.New()
	has.Write([]byte(user.Password))
	sum := has.Sum(nil)
	query.Exec(user.Username, sum, user.NamaLengkap, imageName, userID)
	checkError(err)

	user.Id, _ = strconv.ParseInt(userID, 10, 64)
	renderJson(writer, &user)
}

func login(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("sucess"))
}

func deleteUser(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	userID := mux.Vars(request)["id"]

	result, err := mysqlDB.Exec("DELETE FROM user WHERE ID = ?", userID)
	checkError(err)
	affected, err := result.RowsAffected()
	if affected == 0 {
		writer.WriteHeader(http.StatusNotFound)
		jsonapi.MarshalErrors(writer, []*jsonapi.ErrorObject{{
			Title:  "NotFound",
			Status: strconv.Itoa(http.StatusNotFound),
			Detail: fmt.Sprintf("User with id %s not found", userID),
		}})
	}

	writer.WriteHeader(http.StatusNoContent)
}

func addUser(writer http.ResponseWriter, request *http.Request) {
	imageName, err := FileUpload(request)
	if err != nil {
		http.Error(writer, "Invalid Data", http.StatusBadRequest)
		return
		//checking whether any error occurred retrieving image
	}
	writer.Header().Set("Content-Type", "application/json")

	var user User
	err = json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		jsonapi.MarshalErrors(writer, []*jsonapi.ErrorObject{{
			Title:  "ValidationError",
			Status: strconv.Itoa(http.StatusUnprocessableEntity),
			Detail: "Given request body was invalid",
		}})
		return
	}

	query, err := mysqlDB.Prepare("INSERT INTO user (Username, Password, Nama_Lengkap, Foto) values (?, ?, ?, ?)")
	checkError(err)
	has := sha1.New()
	has.Write([]byte(user.Password))
	sum := has.Sum(nil)
	result, err := query.Exec(user.Username, sum, user.NamaLengkap, imageName)
	checkError(err)
	userID, err := result.LastInsertId()
	checkError(err)

	user.Id = userID
	writer.WriteHeader(http.StatusCreated)
	renderJson(writer, &user)
}

func getAllUser(writer http.ResponseWriter, request *http.Request) {
	rows, err := mysqlDB.Query("SELECT ID, Username ,Password, Nama_Lengkap, Foto FROM user")
	checkError(err)

	var user []*User
	log.Print(rows)
	for rows.Next() {
		var usr User
		if err := rows.Scan(&usr.Id, &usr.Username, &usr.Password, &usr.NamaLengkap, &usr.Foto); err != nil && err != sql.ErrNoRows {
			checkError(err)
		} else {
			user = append(user, &usr)
		}
	}
	renderJson(writer, user)
}
