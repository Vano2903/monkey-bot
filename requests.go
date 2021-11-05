package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type JSONPB struct {
	PersonalBest PB `json:"personalBests"`
}

type jwt struct {
	IdToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
}

type jwt2 struct {
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
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
	req.Header.Set("User-Agent", conf.UserAgent)

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

//given a refresh token it will return a new access and refresh token
func GetNewAccessToken(refreshToken string) (string, string, error) {
	url := "https://securetoken.googleapis.com/v1/token?key=AIzaSyB5m_AnO575kvWriahcF1SFIWp8Fj3gQno"
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf(`grant_type=refresh_token&refresh_token=%s`, refreshToken))
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return "", "", err
	}

	req.Header.Add("Referer", "https://monkeytype.com/")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", conf.UserAgent)

	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", "", errors.New("il refresh token é incorretto o scaduto")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	var jwt2 jwt2
	err = json.Unmarshal(body, &jwt2)
	if err != nil {
		return "", "", err
	}

	return jwt2.IdToken, jwt2.RefreshToken, nil
}

//given an access token it will return a personal best struct
func GetPersonaBest(accessToken string) (PB, error) {
	url := "https://api.monkeytype.com/user"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return PB{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("User-Agent", conf.UserAgent)

	res, err := client.Do(req)
	if err != nil {
		return PB{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return PB{}, errors.New("l'access token é incorretto o scaduto")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return PB{}, err
	}

	var personalBestRaw JSONPB
	err = json.Unmarshal(body, &personalBestRaw)
	if err != nil {
		return PB{}, err
	}

	return filterPersonalBest(personalBestRaw.PersonalBest), nil
}

//filter the raw personal best by only english and italian runs
func filterPersonalBest(toFilter PB) PB {
	var personalBest PB
	//filter the times runs
	for _, time := range toFilter.Time.T15 {
		if time.Language == "english" || time.Language == "italian" {
			personalBest.Time.T15 = append(personalBest.Time.T15, time)
		}
	}

	for _, time := range toFilter.Time.T30 {
		if time.Language == "english" || time.Language == "italian" {
			personalBest.Time.T30 = append(personalBest.Time.T30, time)
		}
	}

	for _, time := range toFilter.Time.T60 {
		if time.Language == "english" || time.Language == "italian" {
			personalBest.Time.T60 = append(personalBest.Time.T60, time)
		}
	}

	for _, time := range toFilter.Time.T120 {
		if time.Language == "english" || time.Language == "italian" {
			personalBest.Time.T120 = append(personalBest.Time.T120, time)
		}
	}

	//filter the words runs
	for _, words := range toFilter.Words.W10 {
		if words.Language == "english" || words.Language == "italian" {
			personalBest.Words.W10 = append(personalBest.Words.W10, words)
		}
	}

	for _, words := range toFilter.Words.W25 {
		if words.Language == "english" || words.Language == "italian" {
			personalBest.Words.W25 = append(personalBest.Words.W25, words)
		}
	}

	for _, words := range toFilter.Words.W50 {
		if words.Language == "english" || words.Language == "italian" {
			personalBest.Words.W50 = append(personalBest.Words.W50, words)
		}
	}

	for _, words := range toFilter.Words.W100 {
		if words.Language == "english" || words.Language == "italian" {
			personalBest.Words.W100 = append(personalBest.Words.W100, words)
		}
	}
	return personalBest
}
