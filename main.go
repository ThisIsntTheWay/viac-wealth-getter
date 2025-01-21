package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

type ViacCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Portfolio struct {
	Name   string `json:"name"`
	Wealth Wealth `json:"wealth"`
}

type Wealth struct {
	TotalValue       float32 `json:"totalValue"`
	TotalPerformance float32 `json:"totalPerformance"`
	TotalReturn      float32 `json:"totalReturn"`
}

var baseUrl string = "https://app.viac.ch"

// Check for and read VIAC credential env vars
func getViacCredentials() (ViacCreds, error) {
	var creds ViacCreds
	creds.Username = os.Getenv("VIAC_USER")
	creds.Password = os.Getenv("VIAC_PASSWORD")
	if creds.Username == "" || creds.Password == "" {
		return ViacCreds{}, fmt.Errorf("not all VIAC credentials passed in env")
	}

	return creds, nil
}

// Perform login
func logIn() ([]*http.Cookie, error) {
	creds, err := getViacCredentials()
	if err != nil {
		return nil, err
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	// This will get us cookies with which to POST auth later
	resp, err := client.Get(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("could not GET cookies: %v", err)
	}
	defer resp.Body.Close()

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed URL parse: %v", err)
	}
	cookies := client.Jar.Cookies(u)

	// If successful, nets us updated session cookies
	postUrl := baseUrl + "/external-login/public/authentication/password/check/"
	jsonBody, _ := json.Marshal(creds)
	req, _ := http.NewRequest(
		http.MethodPost,
		postUrl,
		bytes.NewBuffer(jsonBody),
	)

	for _, c := range cookies {
		req.AddCookie(c)
		if c.Name == "CSRFT759-S" {
			// req.Header.Add makes "X-Csrft..." which is sus to VIAC
			req.Header["X-CSRFT759"] = []string{c.Value}
		}
	}

	req.Header.Add("X-Same-Domain", "1") // Always expected

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed auth: client.Do err: %v", err)
	} else if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed auth: bad status (%d): %s", resp.StatusCode, body)
	}
	defer resp.Body.Close()

	u, _ = url.Parse(postUrl)
	return client.Jar.Cookies(u), nil
}

// Get wealth from VIAC
func GetWealth() (Wealth, error) {
	var wealth Wealth
	cookies, err := logIn()
	if err != nil {
		return wealth, err
	}

	// Get wealth
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseUrl+"/rest/web/wealth/summary", nil)
	if err != nil {
		return wealth, fmt.Errorf("failed creating wealth req: %v", err)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	req.Header.Add("X-Same-Domain", "1")
	req.Header.Add("TE", "trailers")

	resp, err := client.Do(req)
	if err != nil {
		return wealth, fmt.Errorf("could not get wealth: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return wealth, fmt.Errorf("failed wealth: bad status (%d): %s", resp.StatusCode, body)
	} else if err != nil {
		return wealth, fmt.Errorf("failed io.ReadAll on resp.Body: %v", err)
	}

	err = json.Unmarshal(body, &wealth)
	if err != nil {
		return wealth, fmt.Errorf("failed unmarshaling wealth response: %v", err)
	}

	return wealth, nil
}

func main() {
	wealth, err := GetWealth()
	if err != nil {
		panic(err)
	}

	o, _ := json.Marshal(wealth)
	fmt.Println(string(o))
}
