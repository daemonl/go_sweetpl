Ssweetpl
========

Sugar for Go templates

This is a very small abstraction layer on top of the default golang template engine.

At this stage only http/template, but that's just because I'm a bit slack, and I'll fix that soon.

All existing template code will work, then the following are added:

## Extends

  {{ extends "/base.html" }}
  
  {{ define "content" }}
    Blah
  {{ end }}
  
Imports the named template.

The base golang template is the last one in the import chain.
Rendering is done on the last 'extends' base in the chain, then the default package templating takes over.



## Multiple and Optional Define Blocks
  
base.html

  <!DOCTYPE html>
  <html>
    <head>
      <title>{{ template "title" }}</title>
      {{ template "style" }}
    </head>
  </html>
  
  {{ define "title" }}Default Title{{ end }}
  
view.html

  {{ extends "/base.html" }}
  
  {{ define "title" }}Hello World{{ end }}
  
Code

  tpl.Render(w, "view.html", nil)
  
Produces

  <!DOCTYPE html>
  <html>
    <head>
      <title>Hello World</title>

    </head>
  </html>
  
Two things: 

- The 'title' block in view.html is used, it overrides the 'default' in the base. No errors.
- The 'style' call in base.html is not defined anywhere, so it defaults to an empty string

## Default data value for 'template' calls

In the standard package, the default value for "." when calling template is nil, in this 
package it defaults to the "." of the parent.

## Usage:

  tpl := sweetpl.SweeTpl{
    Loader: &sweetpl.DirLoader{
		  BasePath: templatePath,
		},
	}
	
	tpl.Render(w, "view.html", "Hello World")

  
  
