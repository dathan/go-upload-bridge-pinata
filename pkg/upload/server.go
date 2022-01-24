package upload

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func SetUpServer(portStr string) {

	router := httprouter.New()

	router.HandlerFunc("GET", "/", UploadHandler)
	router.HandlerFunc("POST", "/", UploadHandler)
	router.HandlerFunc("OPTIONS", "/", UploadHandler)

	router.HandlerFunc("GET", "/upload", UploadHandler)
	router.HandlerFunc("POST", "/upload", UploadHandler)
	router.HandlerFunc("OPTIONS", "/upload", UploadHandler)

	router.Handle("GET", "/award/:guid", AwardHandler)
	router.Handle("POST", "/contactus", ContactUsHandler)

	//http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	err := http.ListenAndServe(portStr, router)
	if err != nil {
		panic(err)
	}
}
