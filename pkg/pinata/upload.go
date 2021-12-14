package pinata

import (
	"fmt"
	"os"

	"github.com/wabarc/ipfs-pinner/pkg/pinata"
)

type NFTOpenSeaFormat struct {
	Name        string                   `json:"name"`
	Image       string                   `json:"image"`
	Description string                   `json:"description"`
	ExternalUrl string                   `json:"external_url"`
	Attributes  []map[string]interface{} `json:"attributes"`
}

func NewNFTOpenSeaFormat() *NFTOpenSeaFormat {
	return &NFTOpenSeaFormat{
		Attributes: make([]map[string]interface{}, 0),
	}
}

func Upload(files *[]string) (error, string) {
	apikey := os.Getenv("IPFS_PINNER_PINATA_API_KEY")
	secret := os.Getenv("IPFS_PINNER_PINATA_SECRET_API_KEY")
	//log.Info("TODO: Remove this return and call upload twice")
	//return nil, "https://gateway.pinata.cloud/ipfs/QmXRzMVnW3L75GKcJvg9TsJfRcZTXRjsYUeiviFMEfbT9j"
	if apikey == "" || secret == "" {
		panic("Need apikey and secret in the environment")
	}

	pnt := pinata.Pinata{Apikey: apikey, Secret: secret}

	//TODO: fix this, we do not support multiple files in the same string we just parallel upload each file
	for _, file_path := range *files {
		cid, err := pnt.PinFile(file_path)
		if err != nil {
			return err, ""
		}
		//TODO: once above is fixed this makes sense
		return nil, fmt.Sprintf("https://gateway.pinata.cloud/ipfs/%s", cid)
	}

	return nil, ""

}
