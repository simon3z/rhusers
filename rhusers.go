package main

// cspell:words ldap deref rhat

import (
	"bufio"
	"encoding/csv"
	"flag"
	"log"
	"os"
	"strings"
)

var cmdFlags = struct {
	GSheetsFormat bool
	LDAPAddress   string
	SearchBaseDN  string
	QueryString   string
	RecordBuilder func(*Employee) []interface{}
	RecordFormat  func([]interface{}) []string
}{}

func init() {
	flag.BoolVar(&cmdFlags.GSheetsFormat, "g", false, "google sheets format")
	flag.StringVar(&cmdFlags.LDAPAddress, "s", "ldap.corp.redhat.com:389", "ldap server address and port")
	flag.StringVar(&cmdFlags.SearchBaseDN, "b", "ou=users,dc=redhat,dc=com", "base dn for search queries")
	flag.StringVar(&cmdFlags.QueryString, "q", "(uid={})", "ldap query string")

	cmdFlags.RecordBuilder = regularRecordBuilder
	cmdFlags.RecordFormat = stringsRecordFormat
}

func main() {
	flag.Parse()

	log.SetFlags(0)

	w := csv.NewWriter(os.Stdout)

	if cmdFlags.GSheetsFormat {
		w.Comma = '\t'
		cmdFlags.RecordFormat = sheetRecordFormat
	}

	ldap := NewLDAPService()

	err := ldap.Connect("tcp", cmdFlags.LDAPAddress)
	if err != nil {
		log.Fatal("cannot connect to ldap server:", err)
	}

	defer ldap.Disconnect()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		uid := scanner.Text()

		query := strings.ReplaceAll(cmdFlags.QueryString, "{}", string(uid))
		result, err := ldap.SearchEmployee(cmdFlags.SearchBaseDN, query)

		if err != nil {
			log.Fatalln("employee search failed:", err)
		}

		if len(result) == 0 {
			log.Println("employee not found:", query)
			// empty record to maintain input and output rows alignment
			result = []*Employee{{}}
		}

		for _, e := range result {
			w.Write(cmdFlags.RecordFormat(cmdFlags.RecordBuilder(e)))
		}

		w.Flush()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln("failed reading from standard input:", err)
	}
}
