package sweetpl

import (
	"io/ioutil"
	"os"
	"path"
)

type TemplateLoader interface {
	LoadTemplate(string) (string, error)
}

type DirLoader struct {
	BasePath string
}

// DirLoader.LoadTemplate gets BaseAddress + name. No safety checking yet.
func (l *DirLoader) LoadTemplate(name string) (string, error) {
	file, err := os.Open(path.Join(l.BasePath, name))
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(file)
	return string(b), err
}

type MapLoader map[string]string

// MapLoader.LoadTemplate gets name from a map.
func (l *MapLoader) LoadTemplate(name string) (string, error) {
	src, ok := (*l)[name]
	if !ok {
		return "", Errf("Could not find template " + name)
	}
	return src, nil
	//buff := bytes.NewBufferString(src)
	//return buff, nil
}
