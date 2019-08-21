package tilepix

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
  ___       _
 |   \ __ _| |_ __ _
 | |) / _` |  _/ _` |
 |___/\__,_|\__\__,_|

*/

// Data is a TMX file structure holding data.
type Data struct {
	Encoding    string `xml:"encoding,attr"`
	Compression string `xml:"compression,attr"`
	RawData     []byte `xml:",innerxml"`
	// DataTiles is only used when layer encoding is XML.
	DataTiles []*DataTile `xml:"tile"`
}

func (d *Data) String() string {
	return fmt.Sprintf("Data{Compression: %s, DataTiles count: %d}", d.Compression, len(d.DataTiles))
}

func (d *Data) decodeBase64() (data []byte, err error) {
	rawData := bytes.TrimSpace(d.RawData)
	r := bytes.NewReader(rawData)

	encr := base64.NewDecoder(base64.StdEncoding, r)

	var comr io.Reader
	switch d.Compression {
	case "gzip":
		log.Debug("decodeBase64: compression is gzip")

		comr, err = gzip.NewReader(encr)
		if err != nil {
			return
		}
	case "zlib":
		log.Debug("decodeBase64: compression is zlib")

		comr, err = zlib.NewReader(encr)
		if err != nil {
			return
		}
	case "":
		log.Debug("decodeBase64: no compression")

		comr = encr
	default:
		err = ErrUnknownCompression
		log.WithError(ErrUnknownCompression).WithField("Compression", d.Compression).Error("decodeBase64: unable to handle this compression type")
		return
	}

	return ioutil.ReadAll(comr)
}

func (d *Data) decodeCSV() ([]GID, error) {
	cleaner := func(r rune) rune {
		if (r >= '0' && r <= '9') || r == ',' {
			return r
		}
		return -1
	}

	rawDataClean := strings.Map(cleaner, string(d.RawData))

	str := strings.Split(string(rawDataClean), ",")

	gids := make([]GID, len(str))
	for i, s := range str {
		d, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			log.WithError(err).WithField("String to convert", s).Error("decodeCSV: could not parse UInt")
			return nil, err
		}
		gids[i] = GID(d)
	}
	return gids, nil
}
