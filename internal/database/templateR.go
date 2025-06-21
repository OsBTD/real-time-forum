package db

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

var templates *template.Template

// Custom 'in' function to check if a value exists in a slice
func in(slice interface{}, value interface{}) bool {
	switch slice := slice.(type) {
	case []string:
		for _, v := range slice {
			if v == value {
				return true
			}
		}
	case []int:
		for _, v := range slice {
			if v == value {
				return true
			}
		}
	}
	return false
}

func InitTemplates() error {
	var err error
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"in": in,
	})
	templates, err = tmpl.New("").ParseFiles(
		"templates/home.html",
		"templates/login.html",
		"templates/register.html",
		"templates/add_post.html",
		"templates/add_comment.html",
		"templates/error.html",
	)
	if err != nil {
		return fmt.Errorf("template initialization error: %v", err)
	}

	// Log loaded templates
	var templateNames []string
	for _, t := range templates.Templates() {
		templateNames = append(templateNames, t.Name())
	}
	log.Printf("Loaded templates: %v", templateNames)
	return nil
}

func RenderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// Convert data to map if it isn't already
	var dataMap map[string]interface{}
	if data == nil {
		dataMap = make(map[string]interface{})
	} else if m, ok := data.(map[string]interface{}); ok {
		dataMap = m
	} else {
		// If data is not a map, create a new map and add the data as "Data"
		dataMap = map[string]interface{}{
			"Data": data,
		}
	}

	// Ensure we have a title
	if _, hasTitle := dataMap["Title"]; !hasTitle {
		dataMap["Title"] = name
	}

	// Execute template
	err := templates.ExecuteTemplate(w, name+".html", dataMap)
	if err != nil {
		log.Printf("Template execution: %v, Error: %v", name, err)
		HandleError(w, http.StatusInternalServerError, "Internal server error")
	}
}
