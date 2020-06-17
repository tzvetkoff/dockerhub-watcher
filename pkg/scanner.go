package pkg

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/tzvetkoff-go/errors"
	"gopkg.in/yaml.v2"
)

// FindTagResult ...
type FindTagResult int

// FindTagResult values ...
const (
	Found     FindTagResult = 0
	NotFound  FindTagResult = 1
	Different FindTagResult = 2
)

// Scanner ...
type Scanner struct {
	Db           string
	Repos        []string
	DockerClient *DockerClient
	Announcers   []Announcer
}

// NewScanner ...
func NewScanner(config *DockerConfig, dockerClient *DockerClient, announcers []Announcer) *Scanner {
	return &Scanner{
		Db:           config.DB,
		Repos:        config.Repos,
		DockerClient: dockerClient,
		Announcers:   announcers,
	}
}

// Scan ...
func (s *Scanner) Scan() error {
	for _, repo := range s.Repos {
		localTags, err := s.LoadLocalTags(repo)
		if err != nil {
			return errors.Propagate(err, "could not load local tags for repo %s", repo)
		}

		remoteTags, err := s.DockerClient.ListTags(repo)
		if err != nil {
			return errors.Propagate(err, "could not list remote tags")
		}

		for _, remoteTag := range remoteTags {
			switch s.FindTag(localTags, remoteTag) {
			case Found:
				// Old tag.
				s.Announce("=", repo, remoteTag)
			case NotFound:
				// New tag.
				s.Announce("+", repo, remoteTag)
			case Different:
				// Changed tag.
				s.Announce("~", repo, remoteTag)
			}
		}

		for _, localTag := range localTags {
			if s.FindTag(remoteTags, localTag) == NotFound {
				// Deleted tag.
				s.Announce("-", repo, localTag)
			}
		}

		err = s.SaveLocalTags(repo, remoteTags)
		if err != nil {
			return errors.Propagate(err, "could not save local tags for repo %s", repo)
		}
	}

	return nil
}

// LoadLocalTags ...
func (s *Scanner) LoadLocalTags(repo string) ([]*DockerTag, error) {
	b, err := ioutil.ReadFile(path.Join(s.Db, repo) + ".yml")
	if err != nil {
		if os.IsNotExist(err) {
			return []*DockerTag{}, nil
		}

		return nil, errors.Propagate(err, "could not read local tags for repo %s", repo)
	}

	result := []*DockerTag{}
	err = yaml.Unmarshal(b, &result)
	if err != nil {
		return nil, errors.Propagate(err, "could not unmarshal local tags for repo %s", repo)
	}

	return result, nil
}

// SaveLocalTags ...
func (s *Scanner) SaveLocalTags(repo string, db []*DockerTag) error {
	err := os.MkdirAll(path.Dir(path.Join(s.Db, repo)), 0755)
	if err != nil {
		return errors.Propagate(err, "could not create local tags directory for repo %s", repo)
	}

	b, err := yaml.Marshal(db)
	if err != nil {
		return errors.Propagate(err, "could not marshal local tags for repo %s", repo)
	}

	err = ioutil.WriteFile(path.Join(s.Db, repo)+".yml", b, 0644)
	if err != nil {
		return errors.Propagate(err, "could not write local tags for repo %s", repo)
	}

	return nil
}

// FindTag ...
func (s *Scanner) FindTag(haystack []*DockerTag, needle *DockerTag) FindTagResult {
	for _, tag := range haystack {
		if tag.Name == needle.Name {
			if tag.Digest == needle.Digest {
				return Found
			}

			return Different
		}
	}

	return NotFound
}

// Announce ...
func (s *Scanner) Announce(status string, repo string, tag *DockerTag) {
	for _, announcer := range s.Announcers {
		announcer.Announce(status, repo, tag)
	}
}
