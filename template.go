package sweetpl

import (
	"fmt"
	"html/template"
	"io"

	"regexp"
)

var re_extends *regexp.Regexp = regexp.MustCompile(`{{ extends ["']?([^'"}']*)["']? }}`)
var re_include *regexp.Regexp = regexp.MustCompile(`{{ include ["']?([^"]*)["']? }}`)
var re_defineTag *regexp.Regexp = regexp.MustCompile(`{{ ?define "([^"]*)" ?"?([a-zA-Z0-9]*)?"? ?}}`)
var re_templateTag *regexp.Regexp = regexp.MustCompile(`{{ ?template \"([^"]*)" ?([^ ]*)? ?}}`)

//var re_endableTag *regexp.Regexp = regexp.MustCompile(`{{ ?(define|if|range|with|end)[^}]*? ?}}`)

type SweeTpl struct {
	Loader      TemplateLoader
	FuncMap     map[string]interface{} //template.FuncMap
	cache       map[string]*template.Template
	ForceReload bool
}

type NamedTemplate struct {
	Name string
	Src  string
}

func (st *SweeTpl) ClearCache() {
	st.cache = map[string]*template.Template{}
}

func (st *SweeTpl) Render(w io.Writer, name string, data interface{}) error {
	var err error
	var tpl *template.Template

	if !st.ForceReload {
		if st.cache == nil {
			st.ClearCache()
		}
		tpl = st.cache[name]
	}
	if tpl == nil {
		tpl, err = st.assemble(name)
		if err != nil {
			return err
		}
		if !st.ForceReload {
			st.cache[name] = tpl
		}
	}

	if tpl == nil {
		return Errf("Nil template named %s", name)
	}

	err = tpl.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func (st *SweeTpl) GetTemplate(w io.Writer, name string) (*template.Template, error) {
	return st.assemble(name)
}

func (st *SweeTpl) add(stack *[]*NamedTemplate, name string) error {
	tplSrc, err := st.Loader.LoadTemplate(name)
	if err != nil {
		return err
	}

	if len(tplSrc) < 1 {
		return Errf("Empty Template named %s", name)
	}

	extendsMatches := re_extends.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 { //Did Match
		err := st.add(stack, extendsMatches[1])
		if err != nil {
			return err
		}
		tplSrc = re_extends.ReplaceAllString(tplSrc, "")
	}
	namedTemplate := &NamedTemplate{
		Name: name,
		Src:  tplSrc,
	}
	*stack = append((*stack), namedTemplate)
	// The stack is ordered 'general' to 'specific'
	return nil
}

func (st *SweeTpl) assemble(name string) (*template.Template, error) {

	stack := []*NamedTemplate{}

	err := st.add(&stack, name)
	if err != nil {
		return nil, err
	}

	// The stack is ordered 'general' to 'specific'

	blocks := map[string]string{}
	blockId := 0

	// Pick out all 'include' blocks and replace them with the raw
	// text from the requested template (using the configured template
	// directory as a base path)
	for _, namedTemplate := range stack {
		var errInReplace error = nil
		namedTemplate.Src = re_include.ReplaceAllStringFunc(namedTemplate.Src, func(raw string) string {
			parsed := re_include.FindStringSubmatch(raw)
			templatePath := parsed[1]

			subTpl, err := st.Loader.LoadTemplate(templatePath)
			if err != nil {
				errInReplace = err
				return "[error]"
			}

			return subTpl
		})
		if errInReplace != nil {
			return nil, errInReplace
		}
	}

	// Pick out all 'define' blocks and replace them with UIDs.
	// The Map should contain the 'last' definition of each, which is the most specific
	// Must be done separate from below block to make sure all are replaced first.
	for _, namedTemplate := range stack {

		namedTemplate.Src = re_defineTag.ReplaceAllStringFunc(namedTemplate.Src, func(raw string) string {
			parsed := re_defineTag.FindStringSubmatch(raw)
			blockName := fmt.Sprintf("BLOCK_%d", blockId)
			blocks[parsed[1]] = blockName

			blockId += 1
			return "{{ define \"" + blockName + "\" }}"
		})
	}

	var rootTemplate *template.Template

	// 1) Pick out all 'template' blocks, and replace with the UID from above.
	// 2) Render
	for i, namedTemplate := range stack {
		namedTemplate.Src = re_templateTag.ReplaceAllStringFunc(namedTemplate.Src, func(raw string) string {
			parsed := re_templateTag.FindStringSubmatch(raw)
			origName := parsed[1]
			replacedName, ok := blocks[origName]

			// Default the import var to . if not set
			dot := "."
			if len(parsed) == 3 && len(parsed[2]) > 0 {
				dot = parsed[2]
			}
			if ok {
				return fmt.Sprintf(`{{ template "%s" %s }}`, replacedName, dot)
			} else {
				return ""
			}
		})
		var thisTemplate *template.Template

		if i == 0 {
			thisTemplate = template.New(namedTemplate.Name)
			rootTemplate = thisTemplate
		} else {
			thisTemplate = rootTemplate.New(namedTemplate.Name)
		}
		thisTemplate.Funcs(st.FuncMap) // Must be added before Parse.
		_, err := thisTemplate.Parse(namedTemplate.Src)
		if err != nil {
			return nil, err
		}

	}

	return rootTemplate, nil
}
