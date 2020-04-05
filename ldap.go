package main

// cspell:words rhat ldap deref

import (
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
			return nil, fmt.Errorf("too many values (%d) for attribute (%s)", len(k.Values), k.Name)
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
		return "", fmt.Errorf("couldn't identify manager uid (%s)", managerUID)
	}

	ls := ldap.NewSearchRequest(search[1], ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, fmt.Sprintf("(%s)", search[0]), ManagerMailAttributes, nil)

	r, err := s.connection.Search(ls)
	if err != nil {
		return "", err
	}

	if len(r.Entries) == 0 {
		return "", fmt.Errorf("couldn't find manager (%s)", managerUID)
	} else if len(r.Entries) > 1 {
		return "", fmt.Errorf("too many managers found (%s)", managerUID)
	}

	m, err := getAttributesMap(r.Entries[0])
	if err != nil {
		return "", err
	}

	s.cache[managerUID] = getPreferredMail(m)

	return s.cache[managerUID], nil
}

// SearchEmployee searches and returns employees matching the search criteria provided
func (s *LDAPService) SearchEmployee(basedn, search string) ([]*Employee, error) {
	ls := ldap.NewSearchRequest(basedn, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, search, EmployeeAttributes, nil)

	r, err := s.connection.Search(ls)
	if err != nil {
		return nil, err
	}

	eg := []*Employee{}

	for _, i := range r.Entries {
		m, err := getAttributesMap(i)
		if err != nil {
			return nil, err
		}

		managerMail, err := s.fetchManagerMail(m["manager"])
		if err != nil {
			return nil, err
		}

		mail := []string{m["rhatPrimaryMail"]}

		if preferredMail, ok := m["rhatPreferredAlias"]; ok {
			mail = append([]string{preferredMail}, mail...)
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
