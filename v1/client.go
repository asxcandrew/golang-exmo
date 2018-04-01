package exmo

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Client manages all the communication with the API.
type Client struct {
	// Base URL for API requests.
	BaseURL *url.URL

	APIKey    string
	APISecret string

	Trades *TradesService
}

// NewClient creates new API client.
func NewClient() *Client {
	baseURL, _ := url.Parse(BaseURL)

	c := &Client{BaseURL: baseURL}

	return c
}

// newAuthenticatedRequest creates new http request for authenticated routes.
// func (c *Client) newAuthenticatedRequest(m string, refURL string, data map[string]interface{}) (*http.Request, error) {
// 	req, err := c.newRequest(m, refURL, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	nonce := utils.GetNonce()
// 	payload := map[string]interface{}{
// 		"request": "/v1/" + refURL,
// 		"nonce":   nonce,
// 	}

// 	for k, v := range data {
// 		payload[k] = v
// 	}

// 	p, err := json.Marshal(payload)
// 	if err != nil {
// 		return nil, err
// 	}

// 	encoded := base64.StdEncoding.EncodeToString(p)

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Accept", "application/json")
// 	req.Header.Add("X-BFX-APIKEY", c.APIKey)
// 	req.Header.Add("X-BFX-PAYLOAD", encoded)
// 	req.Header.Add("X-BFX-SIGNATURE", c.signPayload(encoded))

// 	return req, nil
// }

func (c *Client) signPayload(payload string) string {
	sig := hmac.New(sha512.New384, []byte(c.APISecret))
	sig.Write([]byte(payload))
	return hex.EncodeToString(sig.Sum(nil))
}

// Auth sets api key and secret for usage is requests that requires authentication.
func (c *Client) Auth(key string, secret string) *Client {
	c.APIKey = key
	c.APISecret = secret

	return c
}

func (c *Client) performRequest(req *http.Request, v interface{}) (*Response, error) {
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		body = []byte(`Error reading body:` + err.Error())
	}

	response := &Response{resp, body}

	err = checkResponse(response)
	if err != nil {
		// Return response in case caller need to debug it.
		return response, err
	}

	if v != nil {
		err = json.Unmarshal(response.Body, v)
		if err != nil {
			return response, err
		}
	}

	return response, nil
}

// checkResponse checks response status code and response
// for errors.
func checkResponse(r *Response) error {
	if c := r.Response.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	// Try to decode error message
	errorResponse := &ErrorResponse{Response: r}
	err := json.Unmarshal(r.Body, errorResponse)
	if err != nil {
		errorResponse.Message = "Error decoding response error message. " +
			"Please see response body for more information."
	}

	return errorResponse
}