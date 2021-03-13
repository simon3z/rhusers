package main

// cspell:words ldap deref rhat jira

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"strings"
)

var cmdFlags = struct {
	GSheetsFormat bool
	LDAPAddress   string
	SearchBaseDN  string
	QueryString   string
	JiraUsername  string
	JiraPassword  string
	JiraAddress   string
	RecordBuilder func(*Employee) []interface{}
	RecordFormat  func([]interface{}) []string
}{}

func init() {
	flag.BoolVar(&cmdFlags.GSheetsFormat, "g", false, "google sheets format")
	flag.StringVar(&cmdFlags.LDAPAddress, "s", "ldap.corp.redhat.com:389", "ldap server address and port")
	flag.StringVar(&cmdFlags.SearchBaseDN, "b", "ou=users,dc=redhat,dc=com", "base dn for search queries")
	flag.StringVar(&cmdFlags.QueryString, "q", "(uid={})", "ldap query string")
	flag.StringVar(&cmdFlags.JiraUsername, "j", "", "jira user name")
	flag.StringVar(&cmdFlags.JiraAddress, "z", "https://issues.redhat.com", "jira server url")

	cmdFlags.RecordBuilder = regularRecordBuilder
	cmdFlags.RecordFormat = stringsRecordFormat
}

func main() {
	flag.Parse()

	log.SetFlags(0)

	r := csv.NewReader(os.Stdin)
	w := csv.NewWriter(os.Stdout)

	if cmdFlags.GSheetsFormat {
		r.Comma = '\t'
		w.Comma = '\t'
		cmdFlags.RecordFormat = sheetRecordFormat
	}

	ldap := NewLDAPService()

	err := ldap.Connect("tcp", cmdFlags.LDAPAddress)
	if err != nil {
		log.Fatal("cannot connect to ldap server:", err)
	}

	defer ldap.Disconnect()

	var jira *JiraService

	if cmdFlags.JiraUsername != "" {
		cmdFlags.JiraPassword = GetPassword("Jira Password", "JIRA_PASSWORD", true)

		jira, err = NewJiraClient(cmdFlags.JiraAddress, cmdFlags.JiraUsername, cmdFlags.JiraPassword)

		if err != nil {
			log.Fatal(err)
		}

		cmdFlags.RecordBuilder = jiraRecordBuilder
	}

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln("input parse error:", err)
		}

		uid := record[len(record)-1]

		query := strings.ReplaceAll(cmdFlags.QueryString, "{}", string(uid))
		result, err := ldap.SearchEmployee(cmdFlags.SearchBaseDN, query)

		if err != nil {
			log.Fatalln("employee search failed:", err)
		}

		if len(result) == 0 {
			log.Println("employee not found:", query)
			// empty record to maintain input and output rows alignment
			result = []*Employee{{UserID: Profile{string: uid}}}
		}

		for _, e := range result {
			if jira != nil {
				if err := jira.BuildEmployee(e); err != nil {
					log.Fatalln("unable to retrieve jira id:", err)
				}
			}

			w.Write(append(record[:len(record)-1], cmdFlags.RecordFormat(cmdFlags.RecordBuilder(e))...))
		}

		w.Flush()
	}
}
