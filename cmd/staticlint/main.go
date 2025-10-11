package main

import (
	"fmt"

	"github.com/SmirnovND/gobase/cmd/staticlint/exitchecker"
	"github.com/gostaticanalysis/nilerr"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	fmt.Println("Enabled analyzers:")
	var analyzers []*analysis.Analyzer

	// Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	standardAnalyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		unreachable.Analyzer,
	}
	analyzers = append(analyzers, standardAnalyzers...)

	// Анализаторы SA из staticcheck.io
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer != nil && v.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Дополнительные анализаторы из staticcheck.io
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name == "ST1000" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Публичные анализаторы
	analyzers = append(analyzers, nilerr.Analyzer)    // Проверка на игнорирование ошибок
	analyzers = append(analyzers, bodyclose.Analyzer) // Проверка закрытия `http.Response.Body`

	// Собственный анализатор
	analyzers = append(analyzers, exitchecker.Analyzer)

	for _, a := range analyzers {
		fmt.Println(a.Name)
	}
	// Запуск multichecker
	multichecker.Main(analyzers...)
}
