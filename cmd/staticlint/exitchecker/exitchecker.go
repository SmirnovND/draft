package exitchecker

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"strings"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "checks for direct calls to os.Exit in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Проверяем, что анализируем пакет `main`
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Проверяем, что файл действительно принадлежит пакету `main`
		if file.Name.Name != "main" {
			continue
		}

		// Исключаем файлы, не являющиеся исходным кодом
		if !strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), ".go") {
			continue
		}

		// Проверяем, есть ли в файле функция main
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				continue
			}

			// Ищем прямой вызов os.Exit внутри main()
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := sel.X.(*ast.Ident)
				if ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main is not allowed")
				}
				return true
			})
		}
	}

	return nil, nil
}
