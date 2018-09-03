package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type BitmexRequest struct {
}

var EXPIRES_INTERVAL int64

func init() {
	EXPIRES_INTERVAL = 10000
}

type Bitmex struct {
	APIID     string
	APISecret string
	BaseURL   string
}

func (b *Bitmex) BaseRequest(method string,
	requestBody *bytes.Buffer,
	result interface{},
	dstURL string,
	headers map[string]interface{}) error {
	log.Println("====================Request start===================")

	log.Println(method, b.BaseURL+dstURL)
	var req *http.Request
	var err error
	if requestBody == nil {
		req, err = http.NewRequest(method, b.BaseURL+dstURL, nil)
	} else {
		req, err = http.NewRequest(method, b.BaseURL+dstURL, requestBody)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v.(string))
	}

	log.Println("headers:", req.Header)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	log.Println("========================Request end==============================")
	log.Println("========================Response start==============================")
	log.Println("limit:", response.Header.Get("x-ratelimit-limit"))
	log.Println("remaining", response.Header.Get("x-ratelimit-remaining"))
	log.Println("reset", response.Header.Get("x-ratelimit-reset"))
	log.Println("========================Response end==============================")

	json.NewDecoder(response.Body).Decode(&result)

	return nil
}

func (b *Bitmex) SendRequestWithSignature(method string,
	requestBody *bytes.Buffer,
	result interface{},
	dstURL string,
	headers map[string]interface{}) error {

	expires := EXPIRES_INTERVAL + time.Now().Unix()
	var reqStr string
	if requestBody != nil {
		reqStr = fmt.Sprintf("%v%v%v%v", method, dstURL, expires, requestBody.String())
	} else {
		reqStr = fmt.Sprintf("%v%v%v", method, dstURL, expires)
	}
	log.Println(reqStr)
	if headers == nil {
		headers = map[string]interface{}{}
	}
	headers["api-expires"] = fmt.Sprintf("%v", expires)
	headers["api-key"] = b.APIID
	headers["api-signature"] = b.GetSignature(reqStr)

	log.Println("headers:", headers)
	b.BaseRequest(method, requestBody, result, dstURL, headers)
	return nil
}

// API ID: WVmbvmemWPD-NWC3ni9ObYus
// API secret: 9ZycWgT4qhIJg2ckXVA6IPIfIjmasoEo_n6jskyODqgs5QXa
func (b *Bitmex) GetOrderbook() error {
	return nil
}
func (b *Bitmex) GetSignature(reqStr string) string {
	hash := hmac.New(sha256.New, []byte(b.APISecret))
	hash.Write([]byte(reqStr))
	return hex.EncodeToString(hash.Sum(nil))
}

type APIKeyPOST struct {
	Name        string `json:"name"`
	Cidr        string `json:"cidr,omitempty"`
	Permissions string `json:"permissions"`
	Enabled     bool   `json:"enables"`
	Token       string `json:"token,omitempty"`
}

func SendRequest(bitmex Bitmex, requestBody *bytes.Buffer, method string, url string, headers map[string]interface{}) (interface{}, error) {
	var result interface{}
	err := bitmex.SendRequestWithSignature(method, requestBody, &result, url, headers)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return result, nil
}

func TestGETAnnouncement(bitmex Bitmex) (interface{}, error) {
	return SendRequest(bitmex, nil, "GET", "/api/v1/announcement", nil)
}

func TestPOSTRiskLimit(bitmex Bitmex) (interface{}, error) {
	requestBody := map[string]interface{}{
		"symbol":    "XBTUSD",
		"riskLimit": 2.0,
	}
	jsonStr, _ := json.Marshal(requestBody)
	log.Println(string(jsonStr))
	return SendRequest(bitmex, bytes.NewBuffer(jsonStr), "POST", "/api/v1/position/riskLimit", nil)
}

func main() {
	bitmex := Bitmex{
		APIID:     "WVmbvmemWPD-NWC3ni9ObYus",
		APISecret: "9ZycWgT4qhIJg2ckXVA6IPIfIjmasoEo_n6jskyODqgs5QXa",
		BaseURL:   "https://www.bitmex.com",
	}
	result, err := TestPOSTRiskLimit(bitmex)
	if err != nil {
		log.Println("ERROR:", err)
	} else {
		log.Println("RESULT:", result)
	}

}
