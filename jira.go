package main

// cspell:words jira jspa

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/andygrunwald/go-jira" // cspell:disable-line
)

// JiraService represents a jira service client
type JiraService struct {
	*jira.Client
}

// NewJiraClient creates a new Jira client using basic authentication
func NewJiraClient(address, username, password string) (*JiraService, error) {
	transport := jira.BasicAuthTransport{Username: username, Password: password}

	client, err := jira.NewClient(transport.Client(), address)
	if err != nil {
		return nil, err
	}

	return &JiraService{client}, nil
}

func statusToError(code int) error {
	return errors.New(strings.ToLower(http.StatusText(code)))
}

// BuildEmployee fills in the jira information for the provided employee
func (j *JiraService) BuildEmployee(e *Employee) error {
	b := j.GetBaseURL()

	for _, a := range e.Mail {
		users, res, err := j.User.Find(a)

		if err != nil {
			if res != nil {
				return fmt.Errorf("request failed: %s", statusToError(res.StatusCode))
			}
			return err
		}

		for _, u := range users {
			if u.EmailAddress != a {
				continue
			}

			e.JiraID = Profile{
				u.Key,
				&url.URL{
					Scheme:   b.Scheme,
					Host:     b.Host,
					Path:     strings.Join([]string{b.Path, "ViewProfile.jspa"}, "/"),
					RawQuery: url.Values(map[string][]string{"name": {u.Key}}).Encode(),
				},
			}

			return nil
		}
	}

	e.JiraID = Profile{"", nil}

	return nil
}
