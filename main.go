package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

func worker(ctx context.Context, jobs <-chan int, wg *sync.WaitGroup, timeout time.Duration) {

	defer wg.Done()

	for id := range jobs {

		if ctx.Err() != nil {
			return
		}

		processMovie(ctx, id, timeout)
	}

}

func processMovie(parentContext context.Context, id int, timeout time.Duration) {

	url := fmt.Sprintf("http://homeworksite.site/%d/info.0.json", id)

	reqCtx, cancel := context.WithTimeout(parentContext, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("%d Request creating error: %v\n", id, err)
		return
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Printf("%d Response error: %v\n", id, err)
		return
	}

	defer resp.Body.Close()
	//status code check
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%d Server error: status %d\n", id, resp.StatusCode)
		return
	}

	var m Movie
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		fmt.Printf("%d JSON parsing error %v\n", id, err)
		return
	}

	fmt.Printf("%d - %s - %d - %s\n", m.ID, m.Title, m.Year, m.Director)

}

func main() {

	from := flag.Int("from", 0, "Amount of films from")
	to := flag.Int("to", 0, "Amount of films to")
	workers := flag.Int("workers", 10, "Amount of workers in worker pool")
	timeout := flag.Duration("timeout", 5*time.Second, "HTTP requirement timeout")
	flag.Parse()

	if *from > *to {
		fmt.Fprintln(os.Stderr, "Error, 'from' must be less than 'to'")
		os.Exit(1)
	}

	if *workers <= 0 {
		fmt.Fprintln(os.Stderr, "Workers amount must be more than 0")
		os.Exit(1)
	}

	if *from <= 0 || *to <= 0 {
		fmt.Fprintln(os.Stderr, "Error, flag must be more than 0")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	jobs := make(chan int, *to-*from+1)
	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(ctx, jobs, &wg, *timeout)
	}

	for id := *from; id <= *to; id++ {
		jobs <- id
	}
	close(jobs)

	wg.Wait()

}
