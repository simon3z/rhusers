package main

// cspell:words rhat jira

import (
	"fmt"
	"net/url"
)

// Profile represents an user profile
type Profile struct {
	string
	Link *url.URL
}

func (p Profile) String() string {
	return p.string
}

// EMailAddress represents an email address
type EMailAddress string

// Employee contains all the needed information about an employee
type Employee struct {
	UserID      Profile
	JiraID      Profile
	FirstName   string
	LastName    string
	Mail        []EMailAddress
	JobTitle    string
	GeoArea     string
	Location    string
	CostCenter  string
	Component   string
	Subproduct  string
	ManagerMail EMailAddress
}

// FullName returns the employee full name
func (e *Employee) FullName() string {
	if e.FirstName != "" && e.LastName != "" {
		return fmt.Sprintf("%s %s", e.FirstName, e.LastName)
	}

	return ""
}

// PreferredMail returns the preferred email address
func (e *Employee) PreferredMail() EMailAddress {
	if len(e.Mail) > 0 {
		return e.Mail[0]
	}

	return ""
}
