package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/tzvetkoff-go/errors"
	"github.com/tzvetkoff-go/xd"
)

// DockerClient ...
type DockerClient struct {
	Db       string
	Username string
	Password string
	Token    string
}

// NewDockerClient ...
func NewDockerClient(config *DockerConfig) *DockerClient {
	return &DockerClient{
		Db:       config.DB,
		Username: config.Username,
		Password: config.Password,
	}
}

// Authenticate ...
func (c *DockerClient) Authenticate() error {
	if c.Token == "" {
		b, err := ioutil.ReadFile(path.Join(c.Db, "token.txt"))
		if err == nil {
			c.Token = string(b)
			return nil
		}

		payload, _ := json.Marshal(struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: c.Username,
			Password: c.Password,
		})
		response, err := c.Request("POST", "https://hub.docker.com/v2/users/login/", payload, false)
		if err != nil {
			return errors.Propagate(err, "could not acquire hub token")
		}

		token, err := xd.DigE(response, "token")
		if err != nil {
			return errors.Propagate(err, "could not extract token from hub response")
		}

		c.Token = token.(string)
		ioutil.WriteFile(path.Join(c.Db, "token.txt"), []byte(token.(string)), 0644)
	}

	return nil
}

// Request ...
func (c *DockerClient) Request(
	verb string,
	url string,
	body []byte,
	authenticated bool,
) (map[string]interface{}, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	httpRequest, err := http.NewRequest(verb, url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Propagate(err, "could not create http request")
	}

	httpRequest.Header.Add("Content-Type", "application/json")

	if authenticated {
		err = c.Authenticate()
		if err != nil {
			return nil, errors.Propagate(err, "could not authenticate")
		}

		httpRequest.Header.Add("Authorization", "JWT "+c.Token)
	}

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, errors.Propagate(err, "error performing http request")
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode >= 400 {
		return nil, errors.New("http server responded with %s", httpResponse.Status)
	}

	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, errors.Propagate(err, "could not read http response body")
	}

	result := map[string]interface{}{}
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return nil, errors.Propagate(err, "could not unmarshal response body")
	}

	return result, nil
}

// ListTags ...
func (c *DockerClient) ListTags(repo string) ([]*DockerTag, error) {
	result := []*DockerTag{}

	for url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags?page_size=1024", repo); true; {
		response, err := c.Request("GET", url, nil, true)
		if err != nil {
			return nil, err
		}

		responseResults, err := xd.DigE(response, "results")
		if err != nil {
			return nil, errors.Propagate(err, "response has no results")
		}

		for _, responseResult := range responseResults.([]interface{}) {
			name, err := xd.DigE(responseResult, "name")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", errors.Propagate(err, "result has no name"))
				continue
			}

			timestamp, err := xd.DigE(responseResult, "last_updated")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", errors.Propagate(err, "result has no last updated timestamp"))
				continue
			}

			digest, err := xd.DigE(responseResult, "images[0].digest")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", errors.Propagate(err, "result has no digest"))
				continue
			}

			result = append(result, &DockerTag{
				Name:      name.(string),
				Timestamp: timestamp.(string),
				Digest:    digest.(string),
			})
		}

		next, err := xd.DigE(response, "next")
		if err != nil {
			return nil, errors.Propagate(err, "response has no next")
		}

		if next == nil {
			break
		}
		url = next.(string)
	}

	return result, nil
}
