package report

import (
	"fmt"
	"sort"

	"github.com/dprio/stress-test/internal/runner"
)

// PrintReport exibe o relatório final no console.
func PrintReport(r runner.Result) {
	fmt.Println()
	fmt.Println("========== Relatório do Teste de Carga ==========")
	fmt.Printf("Tempo total gasto:            %s\n", r.TotalTime)
	fmt.Printf("Total de requests realizados: %d\n", r.TotalRequests)
	fmt.Printf("Requests com status 200:      %d\n", r.StatusCodes[200])

	fmt.Println("Distribuição de outros status:")
	codes := make([]int, 0, len(r.StatusCodes))
	for code := range r.StatusCodes {
		if code != 200 {
			codes = append(codes, code)
		}
	}
	sort.Ints(codes)

	if len(codes) == 0 && r.Errors == 0 {
		fmt.Println("  (nenhum)")
	}
	for _, code := range codes {
		fmt.Printf("  %d: %d\n", code, r.StatusCodes[code])
	}
	if r.Errors > 0 {
		fmt.Printf("  falhas sem resposta (erro de rede/timeout): %d\n", r.Errors)
	}
	fmt.Println("=================================================")
}
