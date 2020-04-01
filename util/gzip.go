package util

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io/ioutil"
	"net/http"
)

func decode(data []byte) ([]byte, error) {
	if isGzip(data) {
		return parseGzip(data)
	}
	return data, nil
}

// isGzip from ContentType
func isGzip(data []byte) bool {
	mime := http.DetectContentType(data)
	return mime == "application/x-gzip"
}

func parseGzip(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	_ = binary.Write(b, binary.LittleEndian, data)
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	unGzip, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return unGzip, nil
}
