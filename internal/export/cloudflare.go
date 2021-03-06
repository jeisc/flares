package export

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Cloudflare sets up authorization to the API.
type Cloudflare struct {
	API       string
	AuthKey   string
	AuthEmail string
	Client    http.Client
}

const errDomainNotFound = "cloudflare: domain not found"

// guarantee interface compliance on build.
var _ CloudDNS = Cloudflare{}

// ExportDNS fetches the BIND DNS table for a domain.
func (cf Cloudflare) ExportDNS(domain string) ([]byte, error) {
	return cf.exportFor(domain)
}

func (cf Cloudflare) exportFor(domain string) ([]byte, error) {
	// fetch the zone for the domain
	zone, err := cf.zoneFor(domain)
	if err != nil {
		return nil, err
	}
	endpoint := cf.API + "/zones/" + zone + "/dns_records/export"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-auth-key", cf.AuthKey)
	req.Header.Add("x-auth-email", cf.AuthEmail)
	res, err := cf.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (cf Cloudflare) zoneFor(domain string) (string, error) {
	endpoint := cf.API + "/zones" + fmt.Sprintf("?name=%v", domain)
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", parsed.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("x-auth-key", cf.AuthKey)
	req.Header.Add("x-auth-email", cf.AuthEmail)
	res, err := cf.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	response := response{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}
	if !response.Success {
		return "", errors.New(errDomainNotFound)
	}
	if len(response.Result) == 0 {
		return "", errors.New(errDomainNotFound)
	}
	return response.Result[0].ID, nil
}

type response struct {
	Success bool `json:"success"`
	Result  []struct {
		ID string `json:"id"`
	} `json:"result"`
}
