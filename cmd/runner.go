package main

import (
	"io"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	TotalTime     time.Duration
	TotalRequests int
	StatusCodes   map[int]int
	Errors        int
}

// Run dispara exatamente `requests` requisições usando `concurrency` workers.
func Run(targetURL string, requests, concurrency int) Result {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        concurrency,
			MaxIdleConnsPerHost: concurrency,
		},
	}

	// O canal de jobs garante que o total seja exato: cada item é
	// consumido por um único worker.
	jobs := make(chan struct{}, requests)
	for range requests {
		jobs <- struct{}{}
	}
	close(jobs)

	// Cada worker acumula seus próprios resultados (sem lock) e
	// os devolve ao final.
	type partial struct {
		codes  map[int]int
		errors int
	}
	partials := make([]partial, concurrency)

	var wg sync.WaitGroup
	start := time.Now()

	for w := range concurrency {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p := partial{codes: make(map[int]int)}
			for range jobs {
				resp, err := client.Get(targetURL)
				if err != nil {
					p.errors++
					continue
				}
				// Consome o body para permitir reuso da conexão.
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				p.codes[resp.StatusCode]++
			}
			partials[id] = p
		}(w)
	}

	wg.Wait()
	elapsed := time.Since(start)

	result := Result{
		TotalTime:     elapsed,
		TotalRequests: requests,
		StatusCodes:   make(map[int]int),
	}
	for _, p := range partials {
		result.Errors += p.errors
		for code, n := range p.codes {
			result.StatusCodes[code] += n
		}
	}
	return result
}
