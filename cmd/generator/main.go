package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultUrl         string = "http://localhost:8080/api/v1/calc"
	defaultMinNum      int    = -100
	defaultMaxNum      int    = 100
	defaultWorkerCount int    = 10

	httpClientTimeout   = 10 * time.Second
	maxIdleConnsPerHost = 100
	maxConnsPerHost     = 100
	idleConnTimeout     = 90 * time.Second
)

type flags struct {
	baseUrl      *string
	minNum       *int
	maxNum       *int
	workersCount *int
}

type response struct {
	status int
	err    error
}

func main() {
	l := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fls := setFlags()

	l.Info().Msgf("Generator started: %d workers -> %s", *fls.workersCount, *fls.baseUrl)

	client := &http.Client{
		Timeout: httpClientTimeout,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			MaxConnsPerHost:     maxConnsPerHost,
			IdleConnTimeout:     idleConnTimeout,
		},
	}

	randNums := generator(ctx, *fls.minNum, *fls.maxNum)

	resp := workerPool(
		ctx,
		*fls.workersCount,
		randNums,
		sendRequest(client, *fls.baseUrl),
	)

	logSuccess(l, resp)

	<-ctx.Done()
	l.Info().Msg("SIGINT received, stopping generator...")
}

func setFlags() flags {
	fls := flags{}

	fls.baseUrl = flag.String("url", defaultUrl, "Base url to calculator server")
	fls.minNum = flag.Int("min", defaultMinNum, "Min random num")
	fls.maxNum = flag.Int("max", defaultMaxNum, "Max random num")
	fls.workersCount = flag.Int("workers", defaultWorkerCount, "Count workers")
	flag.Parse()

	return fls
}

func logSuccess(l zerolog.Logger, responses <-chan response) {
	var statusOK int
	var statusErr int
	for v := range responses {
		if v.err != nil {
			if errors.Is(v.err, context.Canceled) {
				continue
			}

			statusErr++
			l.Error().Err(v.err).Msg("failed to send request")
			continue
		}
		if v.status == http.StatusOK {
			statusOK++
			continue
		}
		statusErr++
		l.Error().Int("status", v.status).Msg("server returned non-200 status")
	}
	l.Info().Int("ok", statusOK).Int("errors", statusErr).Msg("Total requests")
}

func generator(ctx context.Context, min, max int) <-chan int {
	numbers := make(chan int)

	go func() {
		defer close(numbers)

		for {
			select {
			case <-ctx.Done():
				return
			case numbers <- rand.IntN(max-min+1) + min:
			}
		}
	}()

	return numbers
}

func sendRequest(client *http.Client, baseURL string) func(context.Context, int) response {
	return func(ctx context.Context, num int) response {
		url := fmt.Sprintf("%s?num=%d", baseURL, num)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			return response{
				status: http.StatusInternalServerError,
				err:    fmt.Errorf("creating request: %w", err),
			}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return response{
				status: http.StatusInternalServerError,
				err:    fmt.Errorf("sending request: %w", err),
			}
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)

		return response{
			status: resp.StatusCode,
			err:    nil,
		}
	}
}

func workerPool(
	ctx context.Context,
	workerCount int,
	input <-chan int,
	transform func(ctx context.Context, e int) response,
) <-chan response {
	result := make(chan response)
	var wg sync.WaitGroup

	for range workerCount {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-input:
					if !ok {
						return
					}

					r := transform(ctx, v)

					select {
					case <-ctx.Done():
						return
					case result <- r:
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}
