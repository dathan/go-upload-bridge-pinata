package upload

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/dathan/go-upload-bridge-pinata/pkg/pinata"
	log "github.com/sirupsen/logrus"
)

//UploadHandler switches logic based on the method type to either display a test upload page or accept a file upload
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error
	switch r.Method {
	case "GET":
		display(w, "test_upload", nil)

	case "POST":

		/*todo remove -- test response
		sr := NewResponse()
		sr.Status = "OK"
		sr.Payload = make(map[string]interface{})
		sr.Payload["pinata_url"] = "https://gateway.pinata.cloud/ipfs/QmNnfKdUbybj8tvSCu82ojoo8P7bgeueaZudDGCixjzUND"
		sr.WriteResponse(w)
		return
		*/
		f := fileUpload(w, r, "/tmp/upload_tmp")

		if f != nil {
			url, err = moveToPinanta(f)
			if err != nil {
				Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		new_pin, err := saveFormFile(url, r)
		if err != nil {
			Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sr := NewResponse()
		sr.Status = "OK"
		sr.Payload = make(map[string]interface{})
		sr.Payload["pinata_url"] = new_pin
		sr.WriteResponse(w)
		return
	case "OPTIONS":
		SetupCORS(&w)

	default:
		log.Warnf("Unsupported method passed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func saveFormFile(url string, r *http.Request) (string, error) {
	levelTrait := map[string]interface{}{
		"trait_type": "level",
		"value":      100,
	}

	if len(r.Form.Get("name")) == 0 || len(r.Form.Get("description")) == 0 {
		return "", errors.New("Invalid input on the form")
	}

	//https://docs.opensea.io/docs/metadata-standards
	payload := pinata.NewNFTOpenSeaFormat()
	payload.Name = r.Form.Get("name")
	payload.Description = r.Form.Get("description")
	payload.Image = url
	payload.ExternalUrl = "https://foreveraward.com/c/0x44cFE4768bB446bebA11A9D4FF8f12AE97C862c0"
	payload.Attributes = append(payload.Attributes, levelTrait)

	res, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	log.Infof("about to save :%s", res)
	f, err := os.CreateTemp("/tmp/upload_tmp", "data.json")
	if err != nil {
		return "", err
	}

	_, err = f.Write(res)
	if err != nil {
		return "", err
	}

	files := &[]string{f.Name()}

	return moveToPinanta(files)
}

//move the files (soon to be file) to pinata by uploading the file to that service
func moveToPinanta(files *[]string) (string, error) {

	err, url := pinata.Upload(files)
	if err != nil {
		return "", err
	}

	return url, nil

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

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		panic(err)
	}

	log.Infof("DUMPING DATA: %q", dump)
	log.Infof("formName - form data: [%s]", r.Form.Get("formName"))
	log.Infof("formDescription - form data: [%s]", r.Form.Get("formDescription"))
	//TODO: Ignore the name of the field and just upload it.
	f, h, err := r.FormFile("File")
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	defer f.Close()

	err = os.MkdirAll(save_dir, os.ModePerm)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	filename := save_dir + "/" + h.Filename
	files = append(files, filename)
	dst, err := os.Create(filename)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	defer func() {
		if err := dst.Close(); err != nil {
			log.Printf("Error on file close:  %s", err.Error())
		}
	}()

	_, err = io.Copy(dst, f)
	if err != nil {
		log.Printf("create: %s", err.Error())
		Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	log.Infof("We have created %d files %v", len(files), files)
	return &files

}

// TODO should make this an embed using 1.16+ new feature
func display(w http.ResponseWriter, tmpl string, data interface{}) {

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
