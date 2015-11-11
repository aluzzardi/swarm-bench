package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/codegangsta/cli"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/montanaflynn/stats"
)

func worker(requests int, image string, completeCh chan time.Duration) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	for i := 0; i < requests; i++ {
		start := time.Now()

		container, err := client.CreateContainer(docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: image,
			}})
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

func session(requests, concurrency int, image string, completeCh chan time.Duration) {
	var wg sync.WaitGroup

	n := requests / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			worker(n, image, completeCh)
			wg.Done()
		}()
	}
	wg.Wait()

}

func bench(requests, concurrency int, image string) {
	start := time.Now()

	timings := make([]float64, requests)
	completeCh := make(chan time.Duration)
	current := 0
	prevPercent := -1
	go func() {
		for timing := range completeCh {
			timings = append(timings, timing.Seconds())
			current++
			percent := int(float64(current) / float64(requests) * 100)
			if percent > prevPercent {
				log.Printf("[%3.0d%%] %d/%d containers started (%vms)", percent, current, requests, int(timing.Seconds()*1000))
				prevPercent = percent
			}
		}
	}()
	session(requests, concurrency, image, completeCh)
	close(completeCh)

	total := time.Since(start)
	p50th, _ := stats.Median(timings)
	p90th, _ := stats.Percentile(timings, 90)
	p99th, _ := stats.Percentile(timings, 99)

	log.Printf("---")
	log.Printf("Time taken for tests: %s", total.String())
	log.Printf("Time per container: %vms [50th] | %vms [90th] | %vms [99th]", int(p50th*1000), int(p90th*1000), int(p99th*1000))
}

func main() {
	app := cli.NewApp()
	app.Name = "docker-bench"
	app.Usage = "docker benchmarking tool"
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
		cli.StringFlag{
			Name:  "image, i",
			Value: "helloworld",
			Usage: "Image to use for benchmarking.",
		},
	}

	app.Action = func(c *cli.Context) {
		bench(c.Int("requests"), c.Int("concurrency"), c.String("image"))
	}

	app.Run(os.Args)
}
