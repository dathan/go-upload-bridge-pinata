package upload

import "net/http"

func SetUpServer(portStr string) {

	http.HandleFunc("/", UploadHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	err := http.ListenAndServe(portStr, nil)
	if err != nil {
		panic(err)
	}
}
