package server

import (
	"encoding/base64"
	"encoding/json"
)

// Encode encodes the input in base64
func encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Decode decodes the input from base64
func decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, obj)
	if err != nil {
		return err
	}
	return nil
}
