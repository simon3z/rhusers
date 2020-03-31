package main

// cspell:words rhat

import (
	"fmt"
)

// Employee contains all the needed information about an employee
type Employee struct {
	UserID      string
	FirstName   string
	LastName    string
	Mail        []string
	JobTitle    string
	GeoArea     string
	Location    string
	CostCenter  string
	Component   string
	Subproduct  string
	ManagerMail string
}

// FullName returns the employee full name
func (e *Employee) FullName() string {
	if e.FirstName != "" && e.LastName != "" {
		return fmt.Sprintf("%s %s", e.FirstName, e.LastName)
	}

	return ""
}

// RoverProfileLink returns the link to the employee rover profile
func (e *Employee) RoverProfileLink() string {
	if e.UserID != "" {
		return fmt.Sprintf("https://rover.redhat.com/people/profile/%s", e.UserID)
	}

	return e.UserID
}

// PreferredMail returns the preferred email address
func (e *Employee) PreferredMail() string {
	if len(e.Mail) > 0 {
		return e.Mail[0]
	}

	return ""
}
