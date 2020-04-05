package main

// cspell:words rhat ldap deref

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// EmployeeAttributes are the attributes used to fill in the employee struct
var EmployeeAttributes = []string{
	"uid",
	"displayName",
	"sn",
	"rhatPrimaryMail",
	"rhatPreferredAlias",
	"rhatJobTitle",
	"rhatGeo",
	"rhatLocation",
	"rhatCostCenterDesc",
	"rhatRnDComponent",
	"rhatSubproduct",
	"manager",
}

// ManagerMailAttributes are the attributes retrieved for the employee's managers
var ManagerMailAttributes = []string{
	"rhatPrimaryMail",
	"rhatPreferredAlias",
}

// ErrLDAPEntryNotFound is the error returned when the search produced no entries as result
var ErrLDAPEntryNotFound = errors.New("no entries found")

// ErrLDAPEntryTooMany is the error returned when the search produced too many entries as result
var ErrLDAPEntryTooMany = errors.New("too many entries found")

// ErrLDAPAttributesTooMany is the error returned when an entry has too many attributes
var ErrLDAPAttributesTooMany = errors.New("too many attributes found")

// LDAPService represents an ldap service to connect to and interact with
type LDAPService struct {
	connection *ldap.Conn
	cache      map[string]string
}

// NewLDAPService creates a new LDAPService object
func NewLDAPService() *LDAPService {
	return &LDAPService{
		cache: make(map[string]string),
	}
}

// Connect connects the relevant LDAPService to a specific ldap server
func (s *LDAPService) Connect(proto, address string) error {
	err := error(nil)

	s.connection, err = ldap.Dial(proto, address)
	if err != nil {
		return err
	}

	return err
}

// Disconnect disconnects LDAPService from the ldap server
func (s *LDAPService) Disconnect() {
	s.connection.Close()
}

func getAttributesMap(entry *ldap.Entry) (map[string]string, error) {
	m := map[string]string{}

	for _, k := range entry.Attributes {
		if len(k.Values) > 1 {
			return nil, fmt.Errorf("attribute %s: %w", k.Name, ErrLDAPAttributesTooMany)
		}

		m[k.Name] = k.Values[0]
	}

	return m, nil
}

func getPreferredMail(m map[string]string) string {
	if _, ok := m["rhatPreferredAlias"]; ok {
		return m["rhatPreferredAlias"]
	}

	return m["rhatPrimaryMail"]
}

func (s *LDAPService) fetchManagerMail(managerUID string) (string, error) {
	if managerMail, ok := s.cache[managerUID]; ok {
		return managerMail, nil
	}

	search := strings.SplitN(managerUID, ",", 2)
	if len(search) != 2 {
		return "", fmt.Errorf("manager identity %s: %w", managerUID, ErrLDAPEntryNotFound)
	}

	query := fmt.Sprintf("(%s)", search[0])
	request := ldap.NewSearchRequest(search[1], ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, query, ManagerMailAttributes, nil)

	response, err := s.connection.Search(request)
	if err != nil {
		return "", err
	}

	if err := checkOneLDAPEntry(response.Entries); err != nil {
		return "", fmt.Errorf("manager search %s: %w", search, err)
	}

	m, err := getAttributesMap(response.Entries[0])
	if err != nil {
		return "", err
	}

	s.cache[managerUID] = getPreferredMail(m)

	return s.cache[managerUID], nil
}

// SearchEmployee searches and returns employees matching the search criteria provided
func (s *LDAPService) SearchEmployee(basedn, search string) ([]*Employee, error) {
	ls := ldap.NewSearchRequest(basedn, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, search, EmployeeAttributes, nil)

	result, err := s.connection.Search(ls)
	if err != nil {
		return nil, err
	}

	eg := []*Employee{}

	for i := range result.Entries {
		m, err := getAttributesMap(result.Entries[i])
		if err != nil {
			return nil, err
		}

		managerMail, err := s.fetchManagerMail(m["manager"])
		if err != nil {
			return nil, err
		}

		mail := []string{}

		if preferredMail, ok := m["rhatPreferredAlias"]; ok {
			mail = append([]string{preferredMail}, mail...)
		}

		if primaryMail, ok := m["rhatPrimaryMail"]; ok {
			mail = append(mail, primaryMail)
		}

		eg = append(eg, &Employee{
			UserID:      m["uid"],
			FirstName:   m["displayName"],
			LastName:    m["sn"],
			Mail:        mail,
			JobTitle:    m["rhatJobTitle"],
			GeoArea:     m["rhatGeo"],
			Location:    m["rhatLocation"],
			CostCenter:  m["rhatCostCenterDesc"],
			Component:   m["rhatRnDComponent"],
			Subproduct:  m["rhatSubproduct"],
			ManagerMail: managerMail,
		})
	}

	return eg, nil
}

func checkOneLDAPEntry(entries []*ldap.Entry) error {
	switch {
	case len(entries) == 0:
		return ErrLDAPEntryNotFound
	case len(entries) > 1:
		return ErrLDAPEntryTooMany
	}
	return nil
}
