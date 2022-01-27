package upload

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/dathan/go-upload-bridge-pinata/pkg/metadata"
	"github.com/dathan/go-upload-bridge-pinata/pkg/pinata"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

//UploadHandler switches logic based on the method type to either display a test upload page or accept a file upload
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	var url string

	switch r.Method {
	case "GET":
		display(w, "test_upload.htm", nil)

	case "POST":

		/*TODO: remove -- test response
		sr1 := NewResponse()
		sr1.Status = "OK"
		sr1.Payload = make(map[string]interface{})
		sr1.Payload["pinata_url"] = "https://gateway.pinata.cloud/ipfs/QmYPRg2fTXUMi2AsP9E3Gm43bVoPBtaLXB7CZzze4RTfod"
		sr1.WriteResponse(w)
		return
		*/
		f, err := fileUpload(r, "/tmp/upload_tmp")

		if err != nil {
			Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if f != nil {
			url, err = moveToPinanta(f)
			if err != nil {
				Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		//TODO: have a better definition that can be altered
		levelTrait := map[string]interface{}{
			"trait_type": "level",
			"value":      100,
		}

		//https://docs.opensea.io/docs/metadata-standards
		// build the nft struct and set the fields
		NFTDataJson := pinata.NewNFTOpenSeaFormat()
		NFTDataJson.Name = r.Form.Get("name")
		NFTDataJson.Description = r.Form.Get("description")
		NFTDataJson.Image = url
		NFTDataJson.Address = r.Form.Get("address")

		// save the Metadata to the database and get the metadata uuid
		md, err := metadata.New(NFTDataJson).Save()
		if err != nil {
			Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//this award uri allows og tags to render
		NFTDataJson.ExternalUrl = "https://foreveraward.com/award/" + md.UUID

		//TODO: should we save this?
		NFTDataJson.Attributes = append(NFTDataJson.Attributes, levelTrait)

		//save the file to pinata
		new_pin, err := saveFormFile(NFTDataJson, r)
		if err != nil {
			Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// update the metadata with the pin uri
		md.PinataURL = new_pin
		errUpdate := md.Update()
		if errUpdate != nil {
			Error(w, errUpdate.Error(), http.StatusInternalServerError)
			return
		}

		// send a valid response
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

//go:embed assets/*
var assetData embed.FS

//AwardHandler takes in a writer and request and does the following, return meta og tags for award
//TODO: move this into its own service??
func AwardHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// grab data from the database - this implies that data is added on file upload.
	// fill out template data
	// render
	templateData := struct {
		Title       string
		Image       string
		Description string
		GUID        string
	}{
		Title:       "File inside Go",
		Image:       "https://gateway.pinata.cloud/ipfs/QmRPGRxkKtf3Vs3ngEefSGVbXbnZyHTRcs2BjFFSR8GUs3",
		Description: "This is a description of the award",
		GUID:        "13",
	}

	guid := ps.ByName("guid")
	if len(guid) == 0 {
		Error(w, "Invalid input for url", http.StatusInternalServerError)
		return
	}

	p := &pinata.NFTOpenSeaFormat{}

	md := metadata.New(p)
	md, err := md.Get(guid)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// it might be clearer that the struct is not set
	if md.UUID != guid {
		Error(w, "UNKNOWN GUID", http.StatusNotFound)
		return
	}

	templateData.Title = md.Name
	templateData.Description = md.Description
	templateData.Image = md.Image
	templateData.GUID = guid

	log.Infof("METHOD: %s] PARAM: (%s)", r.Method, ps.ByName("guid"))
	display(w, "award.htm", templateData)

}

//TODO: move this into its own service??
func AwardsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p := &pinata.NFTOpenSeaFormat{Address: ps.ByName("address")}
	results, err := metadata.New(p).List()
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {

	case "GET":

		sr := NewResponse()
		sr.Status = "OK"
		sr.Payload = make(map[string]interface{})
		sr.Payload["awards"] = results
		sr.WriteResponse(w)
		return
	case "OPTIONS":
		SetupCORS(&w)
	}

}

func ContactUsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	log.Warnf("CONTACTUS: [%s] [%s] - %s", r.PostFormValue("name"), r.PostFormValue("email"), r.PostFormValue("message"))

	sr := NewResponse()
	sr.Status = "OK"
	sr.Payload = make(map[string]interface{})
	sr.WriteResponse(w)
	return
}

func saveFormFile(payload *pinata.NFTOpenSeaFormat, r *http.Request) (string, error) {

	if len(r.Form.Get("name")) == 0 || len(r.Form.Get("description")) == 0 {
		return "", errors.New("Invalid input on the form")
	}

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
func fileUpload(r *http.Request, save_dir string) (*[]string, error) {
	var files []string
	// left shift 32 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return nil, err
	}

	log.Infof("Form Data - NAME: [%s]", r.Form.Get("name"))
	log.Infof("Form Data - DESC: [%s]", r.Form.Get("description"))
	log.Infof("Form Data - ADDR: [%s]", r.Form.Get("address"))

	if r.Form.Get("name") == "" || r.Form.Get("description") == "" {
		return nil, errors.New("Invalid Input")
	}

	//TODO: Ignore the name of the field and just upload it.
	f, h, err := r.FormFile("File")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = os.MkdirAll(save_dir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	filename := save_dir + "/" + h.Filename
	files = append(files, filename)
	dst, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := dst.Close(); err != nil {
			log.Printf("Error on file close:  %s", err.Error())
		}
	}()

	_, err = io.Copy(dst, f)
	if err != nil {
		log.Errorf("create: %s", err.Error())
		return nil, err
	}

	log.Infof("We have created %d files %v", len(files), files)
	return &files, nil

}

// TODO should make this an embed using 1.16+ new feature
func display(w http.ResponseWriter, tmpl string, data interface{}) {

	t, err := template.ParseFS(assetData, "assets/"+tmpl)
	if err != nil {
		log.Errorf("display: %s", err.Error())
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, data)

}
