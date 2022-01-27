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

	// use awards here
	router.Handle("GET", "/awards/:address", AwardsHandler)

	// allow to record the fact that this award is actually minted and what the token_id will be
	router.Handle("POST", "/award/:guid", AwardHandler)

	// print out an award
	router.Handle("GET", "/award/:guid", AwardHandler)

	router.Handle("POST", "/contactus", ContactUsHandler)

	//http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	err := http.ListenAndServe(portStr, router)
	if err != nil {
		panic(err)
	}
}
