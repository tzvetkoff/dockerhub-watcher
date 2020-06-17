package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/tzvetkoff-go/errors"
)

// SlackAnnouncer ...
type SlackAnnouncer struct {
	WebhookURL string
}

// SlackMessage ...
type SlackMessage struct {
	Attachments []*SlackAttachment `json:"attachments"`
}

// SlackAttachment ...
type SlackAttachment struct {
	Fallback  string `json:"fallback,omitempty"`
	PreText   string `json:"pretext,omitempty"`
	Title     string `json:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty"`
	Text      string `json:"text,omitempty"`
	Color     string `json:"color,omitempty"`
}

// NewSlackAnnouncer ...
func NewSlackAnnouncer(config *SlackAnnouncerConfig) *SlackAnnouncer {
	return &SlackAnnouncer{
		WebhookURL: config.WebhookURL,
	}
}

// Announce ...
func (c *SlackAnnouncer) Announce(status string, repo string, tag *DockerTag) {
	payload := &SlackMessage{
		Attachments: []*SlackAttachment{
			{},
		},
	}

	switch status {
	case "=":
		return
	case "+":
		payload.Attachments[0].Fallback = fmt.Sprintf("Repo *%s* tag *%s* created", repo, tag.Name)
		payload.Attachments[0].PreText = "Tag created"
		payload.Attachments[0].Title = fmt.Sprintf("%s:%s", repo, tag.Name)
		payload.Attachments[0].TitleLink = fmt.Sprintf(
			"https://hub.docker.com/repository/docker/%s/tags?page=1&name=%s",
			repo,
			tag.Name,
		)
		payload.Attachments[0].Text = fmt.Sprintf(
			"Repo *<https://hub.docker.com/repository/docker/%s|%s>* tag *%s* created",
			repo,
			repo,
			tag.Name,
		)
		payload.Attachments[0].Color = "#00CC00"
	case "-":
		payload.Attachments = append(payload.Attachments)

		payload.Attachments[0].Fallback = fmt.Sprintf("Repo *%s* tag *%s* removed", repo, tag.Name)
		payload.Attachments[0].PreText = "Tag removed"
		payload.Attachments[0].Title = fmt.Sprintf("%s:%s", repo, tag.Name)
		payload.Attachments[0].Text = fmt.Sprintf(
			"Repo *<https://hub.docker.com/repository/docker/%s|%s>* tag *%s* removed",
			repo,
			repo,
			tag.Name,
		)
		payload.Attachments[0].Color = "#CC0000"
	case "~":
		payload.Attachments[0].Fallback = fmt.Sprintf("Repo *%s* tag *%s* changed", repo, tag.Name)
		payload.Attachments[0].PreText = "Tag changed"
		payload.Attachments[0].Title = fmt.Sprintf("%s:%s", repo, tag.Name)
		payload.Attachments[0].TitleLink = fmt.Sprintf(
			"https://hub.docker.com/repository/docker/%s/tags?page=1&name=%s",
			repo,
			tag.Name,
		)
		payload.Attachments[0].Text = fmt.Sprintf(
			"Repo *<https://hub.docker.com/repository/docker/%s|%s>* tag *%s* changed",
			repo,
			repo,
			tag.Name,
		)
		payload.Attachments[0].Color = "#FF9900"
	}

	b, _ := json.Marshal(payload)
	err := c.Request(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

// Request ...
func (c *SlackAnnouncer) Request(body []byte) error {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	httpRequest, err := http.NewRequest("POST", c.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return errors.Propagate(err, "could not create http request")
	}

	httpRequest.Header.Add("Content-Type", "application/json")

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return errors.Propagate(err, "error performing http request")
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode >= 400 {
		return errors.New("http server responded with %s", httpResponse.Status)
	}
	defer httpResponse.Body.Close()

	return nil
}
