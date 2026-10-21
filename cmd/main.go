package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/dprio/stress-test/internal/report"
	"github.com/dprio/stress-test/internal/runner"
)

func main() {
	targetURL := flag.String("url", "", "URL do serviço a ser testado (obrigatório)")
	requests := flag.Int("requests", 0, "Número total de requisições (obrigatório, > 0)")
	concurrency := flag.Int("concurrency", 0, "Número de chamadas simultâneas (obrigatório, > 0)")
	flag.Parse()

	if err := validate(*targetURL, *requests, *concurrency); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if *concurrency > *requests {
		*concurrency = *requests
	}

	fmt.Printf("Iniciando teste de carga: url=%s requests=%d concurrency=%d\n", *targetURL, *requests, *concurrency)

	result := runner.Run(*targetURL, *requests, *concurrency)
	report.PrintReport(result)
}

func validate(rawURL string, requests, concurrency int) error {
	if rawURL == "" {
		return fmt.Errorf("--url é obrigatório")
	}
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("--url inválida (use http:// ou https://): %q", rawURL)
	}
	if requests <= 0 {
		return fmt.Errorf("--requests deve ser maior que 0")
	}
	if concurrency <= 0 {
		return fmt.Errorf("--concurrency deve ser maior que 0")
	}
	return nil
}
