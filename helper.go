package main

import (
	"encoding/json"
	"github.com/google/jsonapi"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
)

func FileUpload(r *http.Request) (string, error) {
	//this function returns the filename(to save in database) of the saved file or an error if it occurs
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myPhoto`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myPhoto")
	//replace file with the key your sent your image with
	if err != nil {
		return "", err
	}
	defer file.Close() //close the file when we finish
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	fileLocation := filepath.Join(dir, "files", handler.Filename)
	targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, file); err != nil {
		return "", err
	}
	//here we save our file to our path
	return handler.Filename, nil
}

func env(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	return value
}

func checkError(err interface{}) {
	if err != nil {
		log.Print(err, "\nError connect database")
		return
	}
}

func renderJson(w http.ResponseWriter, buku interface{}) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	if payload, err := jsonapi.Marshal(buku); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		payloads, ok := payload.(*jsonapi.ManyPayload)
		if ok {
			val := reflect.ValueOf(buku)
			payloads.Meta = &jsonapi.Meta{
				"total": val.Len(),
			}
			json.NewEncoder(w).Encode(payloads)
		} else {
			json.NewEncoder(w).Encode(payload)
		}
	}
}
