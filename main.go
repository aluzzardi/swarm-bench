package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/codegangsta/cli"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/montanaflynn/stats"
)

const MILLIS_IN_SECOND = 1000

func worker(name string, id, requests int, image string, labels map[string]string, args []string, completeCh chan time.Duration) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	for i := 0; i < requests; i++ {
		start := time.Now()

		containerName := fmt.Sprintf("%s-%d", name, i+(requests*id))
		container, err := client.CreateContainer(docker.CreateContainerOptions{
			Name: containerName,
			Config: &docker.Config{
				Image:  image,
				Cmd:    args,
				Labels: labels,
			},
			HostConfig: &docker.HostConfig{
				PublishAllPorts: true,
			},
		})
		if err != nil {
			panic(err)
		}

		err = client.StartContainer(container.ID, nil)
		if err != nil {
			panic(err)
		}

		completeCh <- time.Since(start)
	}
}

func session(name string, requests, concurrency int, images []string, labels map[string]string, args []string, completeCh chan time.Duration) {
	var wg sync.WaitGroup
	var size = len(images)
	n := requests / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		image := images[i%size]
		go func(i int) {
			worker(name, i, n, image, labels, args, completeCh)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func bench(requests, concurrency int, images []string, labels map[string]string, args []string) {
	id, _ := rand.Int(rand.Reader, big.NewInt(0xffffff))
	name := fmt.Sprintf("swarm-bench-%X", id)

	start := time.Now()
	timings := make([]float64, requests)
	// Create a buffered channel so our display goroutine can't slow down the workers.
	completeCh := make(chan time.Duration, requests)
	doneCh := make(chan struct{})
	current := 0
	go func() {
		for timing := range completeCh {
			timings = append(timings, timing.Seconds())
			current++
			percent := float64(current) / float64(requests) * 100
			fmt.Printf("[%3.f%%] %d/%d containers started\n", percent, current, requests)
		}
		doneCh <- struct{}{}
	}()
	session(name, requests, concurrency, images, labels, args, completeCh)
	close(completeCh)
	<-doneCh

	total := time.Since(start)
	mean, _ := stats.Mean(timings)
	p90th, _ := stats.Percentile(timings, 90)
	p99th, _ := stats.Percentile(timings, 99)

	meanMillis := mean * MILLIS_IN_SECOND
	p90thMillis := p90th * MILLIS_IN_SECOND
	p99thMillis := p99th * MILLIS_IN_SECOND

	fmt.Printf("\n")
	fmt.Printf("Time taken for tests: %.3fs\n", total.Seconds())
	fmt.Printf("Time per container: %.3fms [mean] | %.3fms [90th] | %.3fms [99th]\n", meanMillis, p90thMillis, p99thMillis)
}

func main() {
	app := cli.NewApp()
	app.Name = "swarm-bench"
	app.Usage = "Swarm Benchmarking Tool"
	app.Version = "0.1.0"
	app.Author = ""
	app.Email = ""
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "concurrency, c",
			Value: 1,
			Usage: "Number of multiple requests to perform at a time. Default is one request at a time.",
		},
		cli.IntFlag{
			Name:  "requests, n",
			Value: 1,
			Usage: "Number of containers to start for the benchmarking session. The default is to just start a single container.",
		},
		cli.StringSliceFlag{
			Name:  "image, i",
			Value: &cli.StringSlice{},
			Usage: "Image(s) to use for benchmarking.",
		},
		cli.StringSliceFlag{
			Name:  "label, l",
			Value: &cli.StringSlice{},
			Usage: "Label(s) to apply to containers.",
		},
	}

	app.Action = func(c *cli.Context) {
		if !c.IsSet("image") && !c.IsSet("i") {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		var labels map[string]string
		if c.IsSet("l") || c.IsSet("label") {
			labels = map[string]string{}
			for _, l := range c.StringSlice("label") {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) != 2 {
					panic("Malformed label " + l)
				}
				labels[parts[0]] = parts[1]
			}
		}

		bench(c.Int("requests"), c.Int("concurrency"), c.StringSlice("image"), labels, c.Args())
	}

	app.Run(os.Args)
}
