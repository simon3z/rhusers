package main

import "fmt"

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

func stringsRecordFormat(i []interface{}) []string {
	record := make([]string, len(i))

	for k, v := range i {
		switch v.(type) {
		case string:
			record[k] = v.(string)
		case fmt.Stringer:
			record[k] = v.(fmt.Stringer).String()
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
		}
	}

	return record
}

func sheetRecordHyperlink(text, link string) string {
	return fmt.Sprintf("=HYPERLINK(\"%s\",\"%s\")", link, text)
}
