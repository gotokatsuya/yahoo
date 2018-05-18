package attribute

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"
)

const (
	endpoint = "https://userinfo.yahooapis.jp/yconnect/v2/attribute"
)

type Client struct {
	client *http.Client
}

type RequestBody struct {
	AccessToken string `url:"access_token"`
	Callback    string `url:"callback,omitempty"`
}

type ResponseBody struct {
	Sub           string `json:"sub"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Gender        string `json:"gender"`
	Zoneinfo      string `json:"zoneinfo"`
	Locale        string `json:"locale"`
	Birthdate     string `json:"birthdate"`
	Nickname      string `json:"nickname"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		client: httpClient,
	}
}

func (c *Client) queryStringify(urlStr string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return urlStr, nil
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr, err
	}
	qs, err := query.Values(opt)
	if err != nil {
		return urlStr, err
	}
	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func (c *Client) NewRequest(body *RequestBody) (*http.Request, error) {
	urlStr, err := c.queryStringify(endpoint, body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*ResponseBody, error) {
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()
	respBody := new(ResponseBody)
	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return respBody, nil
}
