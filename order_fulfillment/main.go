package main

import (
	css "challenge/client"
	orm "challenge/ordermgr"
	"flag"
	"log"
	"sync"
	"time"
)

var (
	endpoint = flag.String("endpoint", "https://api.cloudkitchens.com", "Problem server endpoint")
	auth     = flag.String("auth", "", "Authentication token (required)")
	name     = flag.String("name", "", "Problem name (optional)")
	seed     = flag.Int64("seed", 0, "Problem seed (random if zero)")

	rate = flag.Duration("rate", 500*time.Millisecond, "Inverse order rate")
	min  = flag.Duration("min", 4*time.Second, "Minimum pickup time")
	max  = flag.Duration("max", 8*time.Second, "Maximum pickup time")

	wg = sync.WaitGroup{}
)

func main() {
	flag.Parse()

	client := css.NewClient(*endpoint, *auth)
	id, orders, err := client.New(*name, *seed)
	if err != nil {
		log.Fatalf("Failed to fetch test problem: %v", err)
	}

	// ------ Simulation harness logic goes here using rate, min and max ------

	var actions []css.Action

	for _, order := range orders {
		log.Printf("Received: %+v", order)

		wg.Add(1)
		go orm.PlaceOrder(order, actions)

		time.Sleep(*rate)
	}
	wg.Wait()

	// ------------------------------------------------------------------------

	result, err := client.Solve(id, *rate, *min, *max, actions)
	if err != nil {
		log.Fatalf("Failed to submit test solution: %v", err)
	}
	log.Printf("Test result: %v", result)
}
