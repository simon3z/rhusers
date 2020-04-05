package main

// cspell:words ldap deref rhat

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var cmdFlags = struct {
	GSheetsFormat bool
	LDAPAddress   string
	SearchBaseDN  string
	QueryString   string
}{}

func init() {
	flag.BoolVar(&cmdFlags.GSheetsFormat, "g", false, "google sheets format")
	flag.StringVar(&cmdFlags.LDAPAddress, "s", "ldap.corp.redhat.com:389", "ldap server address and port")
	flag.StringVar(&cmdFlags.SearchBaseDN, "b", "ou=users,dc=redhat,dc=com", "base dn for search queries")
	flag.StringVar(&cmdFlags.QueryString, "q", "(uid={})", "ldap query string")
}

func main() {
	flag.Parse()

	log.SetFlags(0)

	w := csv.NewWriter(os.Stdout)

	if cmdFlags.GSheetsFormat {
		w.Comma = '\t'
	}

	lsv := NewLDAPService()
	err := lsv.Connect("tcp", cmdFlags.LDAPAddress)

	if err != nil {
		log.Fatal(err)
	}

	defer lsv.Disconnect()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		uid := scanner.Text()

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
				e.PreferredMail(),
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

	if err := scanner.Err(); err != nil {
		log.Fatalln("failed reading from standard input:", err)
	}
}
