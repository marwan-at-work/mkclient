package mkclient

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/iancoleman/strcase"
	"marwan.io/swag/swagger"
)

//go:generate esc -o template.go -pkg mkclient client.template

// MakeClient creates a client package based on the swagger schema
func MakeClient(s *swagger.API) error {
	bts, err := makeClient(s)
	if err != nil {
		return err
	}
	os.MkdirAll("./client", 0700)
	return ioutil.WriteFile("./client/client.go", bts, 0660)
}

func makeClient(s *swagger.API) ([]byte, error) {
	var srv service
	imps := map[string]struct{}{}
	s.Walk(func(path string, ep *swagger.Endpoint) {
		var m method
		m.Name = strcase.ToCamel(ep.OperationID)
		m.Method = ep.Method
		m.Path = path
		responseSchema := ep.Responses["200"].Schema
		if responseSchema != nil {
			proto := responseSchema.Prototype
			m.Returns = proto.String()
			m.ReturnsComma = m.Returns + ", "
			m.ReturnVar = "response, "
			pkgPath := proto.PkgPath()
			if pkgPath != "" {
				imps[pkgPath] = struct{}{}
			}
		} else {
			m.Returns = ""
		}
		var a string
		for idx, param := range ep.Parameters {
			paramName := strcase.ToLowerCamel(param.Name)
			if param.In == "query" {
				m.Queries = append(m.Queries, query{param.Name, paramName})
				if idx == 0 {
					a = paramName + " " + param.Type
				} else {
					a += ", " + paramName + " " + param.Type
				}
			} else { // in body, TODO: check what else things can be in
				proto := param.Schema.Prototype
				pkgPath := proto.PkgPath()
				if pkgPath != "" {
					imps[pkgPath] = struct{}{}
				}

				paramName = strcase.ToLowerCamel(proto.Name())
				paramVal := proto.String()
				if idx == 0 {
					a = paramName + " " + paramVal
				} else {
					a += ", " + paramName + " " + paramVal
				}
				m.Body = &body{paramName, paramVal}
			}
		}
		m.Args = a
		srv.Methods = append(srv.Methods, m)
	})
	for i := range imps {
		srv.Imports = append(srv.Imports, i)
	}
	var bts bytes.Buffer
	tmpl := FSMustString(false, "/client.template")
	t := template.Must(template.New("tmpl").Parse(tmpl))
	err := t.Execute(&bts, &srv)
	if err != nil {
		return nil, err
	}
	res, err := format.Source(bts.Bytes())
	return res, err
}

type service struct {
	Imports []string
	Methods []method
}

type method struct {
	Name         string
	Method       string
	Path         string
	Args         string
	Returns      string
	ReturnsComma string
	ReturnVar    string
	Queries      []query
	Body         *body
}

type body struct {
	Name, Value string
}

// if the query was like ?hello=there
// name would be hello, and value would be
// the paramater name
type query struct {
	Name, Value string
}

type arg struct {
	Query bool
	Name  string
	Value string
}
