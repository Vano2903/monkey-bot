package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type jwt struct {
	IdToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
}

//given a email and password (monkeytype credentials) will return a access token
//and the refresh token for those credentials
func Login(email, password string) (string, string, error) {
	url := "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key=AIzaSyB5m_AnO575kvWriahcF1SFIWp8Fj3gQno"
	method := "POST"
	client := &http.Client{}
	payload := strings.NewReader(fmt.Sprintf(`{"email":"%s","password":"%s","returnSecureToken":true}`, email, password))
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return "", "", err
	}

	req.Header.Add("Referer", "https://monkeytype.com/")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", "", errors.New("le credenziali sono incorrette")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	var jwt jwt
	err = json.Unmarshal(body, &jwt)
	if err != nil {
		return "", "", err
	}

	return jwt.IdToken, jwt.RefreshToken, nil
}
