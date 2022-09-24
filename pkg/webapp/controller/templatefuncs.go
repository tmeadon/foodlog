package controller

import (
	"html/template"
	"time"
)

func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"FormatEntryDate": formatEntryDate,
	}
}

func formatEntryDate(d time.Time) string {
	return d.Format("Monday, 2 January 2006 at 15:04")
}
