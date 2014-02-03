package sweetpl

import ()

type TemplateData struct {
	Title string
	Path  string
	User  interface{}
	Nav   map[string]string
	Data  map[string]interface{}
}
