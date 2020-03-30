package main

// cspell:words ldap deref rhat

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	ldap "gopkg.in/ldap.v2"
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

// Employee contains all the needed information about an employee
type Employee struct {
	UserID      string
	FirstName   string
	LastName    string
	Mail        string
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

// LDAPService represents an ldap service to connect to and interact with
type LDAPService struct {
	connection *ldap.Conn
}

// NewLDAPService creates a new LDAPService object
func NewLDAPService() (*LDAPService, error) {
	return &LDAPService{}, nil
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

	return getPreferredMail(m), nil
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

		eg = append(eg, &Employee{
			UserID:      m["uid"],
			FirstName:   m["displayName"],
			LastName:    m["sn"],
			Mail:        getPreferredMail(m),
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

var cmdFlags = struct {
	TabSeparated  bool
	GSheetsFormat bool
	LDAPAddress   string
	SearchBaseDN  string
	QueryString   string
}{}

func init() {
	flag.BoolVar(&cmdFlags.TabSeparated, "t", false, "tab-separated output format")
	flag.BoolVar(&cmdFlags.GSheetsFormat, "g", false, "google sheets format")
	flag.StringVar(&cmdFlags.LDAPAddress, "s", "ldap.corp.redhat.com:389", "ldap server address and port")
	flag.StringVar(&cmdFlags.SearchBaseDN, "b", "ou=users,dc=redhat,dc=com", "base dn for search queries")
	flag.StringVar(&cmdFlags.QueryString, "q", "(uid={})", "ldap query string")
}

func main() {
	flag.Parse()

	log.SetFlags(0)

	w := csv.NewWriter(os.Stdout)

	if cmdFlags.TabSeparated {
		w.Comma = '\t'
	}

	lsv, err := NewLDAPService()
	if err != nil {
		log.Fatal(err)
	}

	err = lsv.Connect("tcp", cmdFlags.LDAPAddress)
	if err != nil {
		log.Fatal(err)
	}

	defer lsv.Disconnect()

	r := bufio.NewReader(os.Stdin)

	for {
		uid, _, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		r, err := lsv.SearchEmployee(cmdFlags.SearchBaseDN, strings.ReplaceAll(cmdFlags.QueryString, "{}", string(uid)))
		if err != nil {
			log.Fatal(err)
		}

		if len(r) == 0 {
			log.Printf("couldn't find employee (uid=%s)", uid)
			r = []*Employee{{}}
		}

		for _, e := range r {
			var userID = ""

			if cmdFlags.GSheetsFormat && e.UserID != "" {
				userID = fmt.Sprintf("=HYPERLINK(\"%s\",\"%s\")", e.RoverProfileLink(), e.UserID)
			} else {
				userID = e.UserID
			}

			w.Write([]string{
				userID,
				e.FullName(),
				e.Mail,
				e.JobTitle,
				e.GeoArea,
				e.Location,
				e.CostCenter,
				e.Component,
				e.Subproduct,
				e.ManagerMail,
			})
		}

		w.Flush()
	}
}
