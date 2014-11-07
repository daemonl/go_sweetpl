package sweetpl

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
	st := &SweeTpl{
		Loader: &MapLoader{
			"vars.html":                       `<title>{{.Title}}</title> Key={{ .Data.Key }}`,
			"Folder/Main-file_name.html.twig": tplMain,
			"sub1.html":                       tplSub1,
			"sub2.html":                       tplSub2,
			"plaintext.html":                  "<Plain>",
		},
	}

	data := &TemplateData{
		Title: "Hello World",
		Data: map[string]interface{}{
			"Key":   "Value",
			"Slice": []string{"One", "Two", "Three"},
		},
	}

	TplRun(t, st, "vars.html", data, "<title>Hello World</title>", "Key=Value")
	TplRun(t, st, "sub2.html", data, "<MAIN>", "<SUB1>", "<SUB2>", "</SUB2>", "</SUB1>", "</MAIN>")
	TplRun(t, st, "sub1.html", data, "<MAIN>", "<SUB1>", "</SUB1>", "<Plain>", "</MAIN>")

}

func TplRun(t *testing.T, st *SweeTpl, name string, data interface{}, check ...string) {
	w := &bytes.Buffer{}
	err := st.Render(w, name, data)
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}
	str := w.String()
	fmt.Println(str)
	fmt.Println("-=-=-=-=-=-")

	for _, c := range check {
		MustContain(t, &str, c)
	}
	fmt.Println("==-==-==-==")

}

func MustContain(t *testing.T, str *string, check string) {

	index := strings.Index(*str, check)
	if index < 0 {
		fmt.Printf("Template did not contain %s in the correct order\n", check)
		t.Fail()
		return
	}
	*str = (*str)[index:]
}

var tplMain string = `<MAIN>
{{ template "body" }}
{{ include "plaintext.html" }}
</MAIN>`

var tplSub1 string = `{{ extends "Folder/Main-file_name.html.twig" }}
{{ define "body" }}
<SUB1>
{{ template "content" .Data }}
</SUB1>
{{ end }}
{{ define "content" }}
	{{ if eq 1 1 }}
	  {{ range .Slice }}
	    {{ . }}
	  {{ else }}
	    NO SLICE
	  {{ end }} 
	{{ else }}
	{{ end }}
<DEFAULT CONTENT>
{{ end }}
`

var tplSub2 string = `
{{ extends 'sub1.html' }}
{{ define "content" }}
<SUB2></SUB2>
{{ end }}
`
