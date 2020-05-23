module github.com/simon3z/rhusers

go 1.15

require (
	github.com/andygrunwald/go-jira v1.11.0 // NOTE: later versions fail at user search because of the username/query parameter change
	github.com/go-ldap/ldap/v3 v3.2.4
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
)
