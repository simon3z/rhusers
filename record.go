package main

// cspell:words jira

import (
	"fmt"
	"reflect"
)

func baseRecordBuilder(e *Employee) []interface{} {
	return []interface{}{
		e.UserID,
		e.FullName(),
		e.PreferredMail(),
		e.JobTitle,
		e.GeoArea,
		e.Location,
		e.CostCenter,
		e.Component,
		e.Subproduct,
	}
}

func regularRecordBuilder(e *Employee) []interface{} {
	return append(baseRecordBuilder(e), e.ManagerMail)
}

func jiraRecordBuilder(e *Employee) []interface{} {
	return append(baseRecordBuilder(e), e.JiraID, e.ManagerMail)
}

func stringsRecordFormat(i []interface{}) []string {
	record := make([]string, len(i))

	for k, v := range i {
		r := reflect.ValueOf(v)

		if r.Kind() == reflect.String {
			record[k] = r.String()
			continue
		}

		switch v.(type) {
		case fmt.Stringer:
			record[k] = v.(fmt.Stringer).String()
		default:
			record[k] = "<unknown>"
		}
	}

	return record
}

func sheetRecordFormat(i []interface{}) []string {
	record := stringsRecordFormat(i)

	for k, v := range i {
		switch u := v.(type) {
		case Profile:
			if u.Link != nil {
				record[k] = sheetRecordHyperlink(u.string, u.Link.String())
			}
		case EMailAddress:
			record[k] = sheetRecordHyperlink(string(u), fmt.Sprintf("mailto:%s", u))
		}
	}

	return record
}

func sheetRecordHyperlink(text, link string) string {
	return fmt.Sprintf("=HYPERLINK(\"%s\",\"%s\")", link, text)
}
