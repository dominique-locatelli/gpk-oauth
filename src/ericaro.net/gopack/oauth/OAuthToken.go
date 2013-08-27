package oauth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
)

type OAuthToken struct {
	privateKey       *rsa.PrivateKey
	consumerKey      string
	oauthToken       string
	oauthTokenSecret string
}

// Method of interface json/encoding/Unmarshaler
func (t *OAuthToken) UnmarshalJSON(bytes []byte) (err error) {
	//[DL] TODO Add unit test
	array := make([][]byte, 4)
	json.Unmarshal(bytes, array)

	t.privateKey, err = x509.ParsePKCS1PrivateKey(array[0])
	if err != nil {
		return err
	}

	t.consumerKey = string(array[1])
	t.oauthToken = string(array[2])
	t.oauthTokenSecret = string(array[3])

	return
}

// Method of interface json/encoding/Marshaler
func (t *OAuthToken) MarshalJSON() ([]byte, error) {
	//[DL] TODO Add unit test
	array := make([][]byte, 4)

	array[0] = x509.MarshalPKCS1PrivateKey(t.privateKey)

	array[1] = []byte(t.consumerKey)
	array[2] = []byte(t.oauthToken)
	array[3] = []byte(t.oauthTokenSecret)

	return json.Marshal(array)
}
