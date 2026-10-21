package runner

import (
	"io"
	"net/http"
	"sync"
	"time"
)

// Stat acumula contagem e tempo total de um grupo de requisições.
type Stat struct {
	Count int
	Total time.Duration
}

func (s *Stat) Add(d time.Duration) {
	s.Count++
	s.Total += d
}

func (s *Stat) Merge(o Stat) {
	s.Count += o.Count
	s.Total += o.Total
}

// Avg retorna o tempo médio; zero se não houver amostras.
func (s Stat) Avg() time.Duration {
	if s.Count == 0 {
		return 0
	}
	return s.Total / time.Duration(s.Count)
}

type Result struct {
	TotalTime     time.Duration
	TotalRequests int
	Stats         map[int]*Stat // por status HTTP
	Errors        Stat          // falhas sem resposta HTTP
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

	jobs := make(chan struct{}, requests)
	for range requests {
		jobs <- struct{}{}
	}
	close(jobs)

	type partial struct {
		stats  map[int]*Stat
		errors Stat
	}
	partials := make([]partial, concurrency)

	var wg sync.WaitGroup
	start := time.Now()

	for w := range concurrency {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p := partial{stats: make(map[int]*Stat)}
			for range jobs {
				reqStart := time.Now()
				resp, err := client.Get(targetURL)
				if err != nil {
					p.errors.Add(time.Since(reqStart))
					continue
				}
				// Lê o body inteiro: o tempo medido é o da resposta completa.
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				elapsed := time.Since(reqStart)

				st := p.stats[resp.StatusCode]
				if st == nil {
					st = &Stat{}
					p.stats[resp.StatusCode] = st
				}
				st.Add(elapsed)
			}
			partials[id] = p
		}(w)
	}

	wg.Wait()
	totalTime := time.Since(start)

	result := Result{
		TotalTime:     totalTime,
		TotalRequests: requests,
		Stats:         make(map[int]*Stat),
	}
	for _, p := range partials {
		result.Errors.Merge(p.errors)
		for code, s := range p.stats {
			dst := result.Stats[code]
			if dst == nil {
				dst = &Stat{}
				result.Stats[code] = dst
			}
			dst.Merge(*s)
		}
	}
	return result
}
