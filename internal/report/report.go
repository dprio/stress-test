package report

import (
	"fmt"
	"sort"
	"time"

	"github.com/dprio/stress-test/internal/runner"
)

func avg(s runner.Stat) time.Duration {
	return s.Avg().Round(time.Microsecond)
}

// PrintReport exibe o relatório final no console.
func PrintReport(r runner.Result) {
	var ok runner.Stat
	if s := r.Stats[200]; s != nil {
		ok = *s
	}

	fmt.Println()
	fmt.Println("========== Relatório do Teste de Carga ==========")
	fmt.Printf("Tempo total gasto:            %s\n", r.TotalTime)
	fmt.Printf("Total de requests realizados: %d\n", r.TotalRequests)
	fmt.Printf("Requests com status 200:      %d (tempo médio: %s)\n", ok.Count, avg(ok))

	fmt.Println("Distribuição de outros status:")
	codes := make([]int, 0, len(r.Stats))
	for code := range r.Stats {
		if code != 200 {
			codes = append(codes, code)
		}
	}
	sort.Ints(codes)

	if len(codes) == 0 && r.Errors.Count == 0 {
		fmt.Println("  (nenhum)")
	}
	for _, code := range codes {
		s := r.Stats[code]
		fmt.Printf("  %d: %d (tempo médio: %s)\n", code, s.Count, avg(*s))
	}
	if r.Errors.Count > 0 {
		fmt.Printf("  falhas sem resposta (erro de rede/timeout): %d (tempo médio até falhar: %s)\n",
			r.Errors.Count, avg(r.Errors))
	}
	fmt.Println("=================================================")
}
