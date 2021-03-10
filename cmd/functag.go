package main

import (
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"strings"
	"text/template"
)

var tpl = template.Must(template.New("").Parse(strings.TrimSpace(`
package {{ .PackageName }}

import "github.com/anttioo/functag"
{{ range .Funcs }}
var _ = functag.RegisterFunc((*{{ .Dest }}).{{ .FuncName }},{{ .Tag }}){{ end }}
`)))

type fn struct {
	Dest     string
	FuncName string
	Tag      string
}

type tplData struct {
	PackageName string
	Funcs       []*fn
}

func declRecvName(decl *ast.FuncDecl) string {
	if recv, isRecv := decl.Recv.List[0].Type.(*ast.StarExpr); isRecv {
		return recv.X.(*ast.Ident).Name
	}
	return ""
}

func filterTagComments(decl *ast.FuncDecl) []string {
	var tagComments []string
	for k := range decl.Doc.List {
		comment := decl.Doc.List[k].Text
		if strings.HasPrefix(comment, "//#") {
			tagComments = append(tagComments, strings.TrimSpace(comment[3:]))
		}
	}
	return tagComments
}

func handleDecl(decl *ast.FuncDecl) *fn {
	if decl.Doc == nil || decl.Recv == nil  {
		return nil
	}
	var tagComments = filterTagComments(decl)
	if len(tagComments) > 0 {
		return &fn{
			Dest:     declRecvName(decl),
			FuncName: decl.Name.Name,
			Tag:      "`" + strings.Join(tagComments, " ") + "`",
		}
	}

	return nil
}

func packageFuncDeclsForEach(pkg *packages.Package, fn func(*ast.FuncDecl)) {
	for i := range pkg.Syntax {
		for j := range pkg.Syntax[i].Decls {
			if decl, ok := pkg.Syntax[i].Decls[j].(*ast.FuncDecl); ok {
				fn(decl)
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expecting output filename")
	}
	fileName := os.Args[1]

	cfg := &packages.Config{
		Mode: packages.LoadSyntax ,
	}
	pkgs, err := packages.Load(cfg, ".")

	if err != nil {
		log.Fatal(err)
	}
	data := tplData{
		PackageName: pkgs[0].Name,
	}
	packageFuncDeclsForEach(pkgs[0], func(decl *ast.FuncDecl) {
		if fn := handleDecl(decl); fn != nil {
			data.Funcs = append(data.Funcs, fn)
		}
	})

	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	err = tpl.ExecuteTemplate(f, "", data)
	if err != nil {
		log.Fatal(err)
	}
}
