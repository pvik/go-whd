package whd

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/go-retryablehttp"
)

func GetAsset(uri string, user User, assetNumber string, asset *[]Asset, sslVerify bool) error {
	req, err := retryablehttp.NewRequest("GET", uri+urn+"Assets", nil)
	if err != nil {
		return err
	}

	WrapAuth(req, user)

	q := req.URL.Query()
	q.Add("assetNumber", assetNumber)
	req.URL.RawQuery = q.Encode()

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = 10
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)

	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(data, &asset); err != nil {
		log.Printf("error unmarshalling: %s | %s", err, data)
		return err
	}

	return nil
}

func GetAssets(uri string, user User, qualifier string, limit uint, page uint, asset *[]Asset, sslVerify bool) error {
	req, err := retryablehttp.NewRequest("GET", uri+urn+"Assets", nil)
	if err != nil {
		return err
	}

	WrapAuth(req, user)

	if limit == 0 {
		limit = 25
	} else if limit > 100 {
		limit = 100
	}

	if page == 0 {
		page = 1
	}

	q := req.URL.Query()
	q.Add("qualifier", qualifier)
	q.Add("limit", strconv.FormatUint(uint64(limit), 10))
	q.Add("page", strconv.FormatUint(uint64(page), 10))
	req.URL.RawQuery = q.Encode()

	var client *http.Client
	if !sslVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	retryclient := retryablehttp.NewClient()
	retryclient.RetryMax = 10
	retryclient.HTTPClient = client

	resp, err := retryclient.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(data, &asset); err != nil {
		log.Printf("error unmarshalling: %s | %s", err, data)
		return err
	}

	return nil
}
