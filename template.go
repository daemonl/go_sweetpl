package sweetpl

import (
	"fmt"
	"html/template"
	"io"
	"regexp"
)

var re_extends *regexp.Regexp = regexp.MustCompile("{{ extends [\"']?([^'\"}']*)[\"']? }}")
var re_defineTag *regexp.Regexp = regexp.MustCompile("{{ ?define \"([^\"]*)\" ?}}")
var re_templateTag *regexp.Regexp = regexp.MustCompile("{{ ?template \"([^\"]*)\" ([^ ]*)? ?}}")

//var re_endableTag *regexp.Regexp = regexp.MustCompile(`{{ ?(define|if|range|with|end)[^}]*? ?}}`)

type SweeTpl struct {
	Loader  TemplateLoader
	FuncMap map[string]interface{} //template.FuncMap
}

type NamedTemplate struct {
	Name string
	Src  string
}

func (st *SweeTpl) Render(w io.Writer, name string, data interface{}) error {
	tpl, err := st.assemble(name)
	if err != nil {
		return err
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

	extendsMatches := re_extends.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 { //Did Match
		st.add(stack, extendsMatches[1])
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

	st.add(&stack, name)

	// The stack is ordered 'general' to 'specific'

	blocks := map[string]string{}
	blockId := 0

	var rootTemplate *template.Template

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

	// 1) Pick out all 'template' blocks, and replace with the UID from above.
	// 2) Render
	for i, namedTemplate := range stack {
		namedTemplate.Src = re_templateTag.ReplaceAllStringFunc(namedTemplate.Src, func(raw string) string {
			parsed := re_templateTag.FindStringSubmatch(raw)
			origName := parsed[1]
			replacedName, ok := blocks[origName]

			// Default the import var to . if not set
			dot := "."
			if len(parsed) == 3 {
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
