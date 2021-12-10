package upload

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wabarc/ipfs-pinner/pkg/pinata"
)

//UploadHandler switches logic based on the method type to either display a test upload page or accept a file upload
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		display(w, "test_upload", nil)

	case "POST":
		f := fileUpload(w, r, "/tmp/upload_tmp")
		if f != nil {
			moveToPinanta(f, w, r)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

//move the files (soon to be file) to pinata by uploading the file to that service
func moveToPinanta(files *[]string, w http.ResponseWriter, r *http.Request) {

	sr := NewResponse()
	sr.Status = "OK"
	sr.Payload = make(map[string]interface{})

	apikey := os.Getenv("IPFS_PINNER_PINATA_API_KEY")
	secret := os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
	if apikey == "" || secret == "" {
		panic("Need apikey and secret in the environment")
	}

	pnt := pinata.Pinata{Apikey: apikey, Secret: secret}

	//TODO: fix this, we do not support multiple files in the same string we just parallel upload each file
	for _, file_path := range *files {
		cid, err := pnt.PinFile(file_path)
		if err != nil {
			sr.Status = "ERROR"
			sr.Msg = err.Error()
			sr.WriteResponse(w)
			return
		}
		//TODO: once above is fixed this makes sense
		sr.Payload["pinata_url"] = fmt.Sprintf("https://gateway.pinata.cloud/ipfs/%s", cid)
	}
	sr.WriteResponse(w)

}

//upload the file to a save_dir and return the uploaded location where the file is
func fileUpload(w http.ResponseWriter, r *http.Request, save_dir string) *[]string {
	var files []string
	// left shift 32 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		panic(err)
	}

	log.Infof("Looking at form data: [%s]", r.Form.Get("testjson"))

	f, h, err := r.FormFile("myfile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	defer f.Close()

	err = os.MkdirAll(save_dir, os.ModePerm)
	if err != nil {
		log.Printf("create: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	filename := save_dir + "/" + h.Filename
	files = append(files, filename)
	dst, err := os.Create(filename)
	defer func() {
		if err := dst.Close(); err != nil {
			log.Printf("Error on file close:  %s", err.Error())
		}
	}()

	_, err = io.Copy(dst, f)
	return &files

}

// TODO should make this an embed using 1.16+ new feature
func display(w http.ResponseWriter, tmpl string, data interface{}) {
	log.Printf("executing template %s.html\n", tmpl)

	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	temp := template.Must(template.New(tmpl).Funcs(funcMap).ParseFiles(tmpl + ".html"))

	//err := templates.ExecuteTemplate(w, tmpl+".html", data)
	err := temp.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
}

func writeJsonResponse(w http.ResponseWriter, js ResponseEnvelope) {
	jbyte, err := json.Marshal(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jbyte)
	log.Infof("STRUCTURED_RESPONSE: %s", string(jbyte))

}
