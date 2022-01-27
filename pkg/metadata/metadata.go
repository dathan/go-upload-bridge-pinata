package metadata

import (
	"fmt"

	goshard "github.com/dathan/go-shard"
	"github.com/dathan/go-upload-bridge-pinata/pkg/pinata"
	"github.com/sirupsen/logrus"
)

type MetaData struct {
	*pinata.NFTOpenSeaFormat
	shardConnection *goshard.ShardConnection `json:"-"`
	PinataURL       string
	UUID            string
}

type MetaDataResult struct {
	*pinata.NFTOpenSeaFormat
	PinataURL string `json:"pinata_url,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	Address   string `json:"address,omitempty"`
}

type MetaDataResults []MetaDataResult

func New(p *pinata.NFTOpenSeaFormat) *MetaData {
	//TODO Replace this package it is to manual for constantly casting in code all over the place
	sc := ShardConfig{}
	err, sConn := sc.NewShardConnectionByShardId(1)
	if err != nil {
		panic(err)
	}

	return &MetaData{p, sConn, "", ""}

}

func (s *MetaData) List() (MetaDataResults, error) {
	logrus.Infof("Looking up the awards for: [%s]", s.Address)
	err, rows := s.shardConnection.SelectAll("SELECT *, BIN_TO_UUID(uuid) as id FROM awards WHERE address=?", s.Address)
	if err != nil {

		logrus.Infof("err %s", err)
		return nil, err
	}

	results := MetaDataResults{}

	for _, row := range rows {

		nftfomat := &pinata.NFTOpenSeaFormat{
			Name:        fmt.Sprintf("%s", row["name"]),
			Description: fmt.Sprintf("%s", row["description"]),
			ExternalUrl: fmt.Sprintf("%s", row["external_url"]),
			Address:     fmt.Sprintf("%s", row["address"]),
		}

		result := MetaDataResult{
			UUID:             fmt.Sprintf("%s", row["id"]),
			PinataURL:        fmt.Sprintf("%s", row["pinata_url"]),
			NFTOpenSeaFormat: nftfomat,
		}

		logrus.Infof("ADDING RESULT: %v", result)

		results = append(results, result)
	}
	return results, nil
}

// add the nft metadata end file
func (s *MetaData) SetPinUrl(url string) *MetaData {
	s.PinataURL = url
	return s
}

// we want to refresh the object to the db
func (s *MetaData) Update() error {
	row := &goshard.Row{"uuid": s.UUID, "pinata_url": s.PinataURL}

	var where_binds []interface{}
	where_binds = append(where_binds, s.UUID)

	err, _ := s.shardConnection.Update("awards", row, "uuid=?", where_binds)
	return err
}

//Save a record and return the uuid or error
func (s *MetaData) Save() (*MetaData, error) {

	dberr, guid := s.shardConnection.SelectRow("SELECT UUID() as id")
	if dberr != nil {
		return s, dberr
	}

	logrus.Warnf("UUID: %s", guid["id"])

	s.UUID = fmt.Sprintf("%s", guid["id"])
	row := &goshard.Row{
		"uuid":        fmt.Sprintf("UUID_TO_BIN('%s')", s.UUID),
		"name":        s.Name,
		"description": s.Description,
		"image":       s.Image,
		"pinata_url":  s.PinataURL,
		"address":     s.Address,
		"create_date": "NOW()",
	}

	// add the row to the db
	err, res := s.shardConnection.InsertIgnore("awards", row)
	if err != nil {
		return s, err
	}

	num, err := res.RowsAffected()
	if err != nil {
		return s, err
	}

	logrus.Infof("Saved Metadata %d]  id: %s", num, s.UUID)
	return s, nil
}

//Get the MetaData for an id if it exists
func (s *MetaData) Get(id string) (*MetaData, error) {

	logrus.Infof("Getting ID: %s", id)
	err, row := s.shardConnection.SelectRow("SELECT *, BIN_TO_UUID(uuid) as id FROM awards where uuid=UUID_TO_BIN(?)", id)
	if err != nil {
		return s, err
	}

	logrus.Infof("ROWINFO:%s", row)

	//uuid is defined we minted the row otherwise it doesn't exist and throw a 404
	if _, ok := row["uuid"]; !ok {
		return s, nil
	}

	s.Name = fmt.Sprintf("%s", row["name"])
	s.Description = fmt.Sprintf("%s", row["description"])
	s.Image = fmt.Sprintf("%s", row["image"])
	s.PinataURL = fmt.Sprintf("%s", row["pinata_url"])
	s.UUID = id

	return s, nil

}
