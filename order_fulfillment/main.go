package main

import (
	css "challenge/client"
	omg "challenge/ordermgr"
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
	pickupMin  = flag.Duration("pickup-min", 4*time.Second, "Minimum pickup time")
	pickupMax  = flag.Duration("pickup-max", 8*time.Second, "Maximum pickup time")

	max_cold_orders = 6 
	max_hot_orders = 6
	max_room_orders = 12
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
	var orderMgr omg.OrderMgr 
	orderMgr.Init(pickupMin, pickupMax, max_cold_orders, max_hot_orders, max_room_orders)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go orderMgr.PlaceOrders(&orders, rate, &wg, &actions)
	go orderMgr.PickupOrders(&wg, &actions)
	wg.Wait()

	// ------------------------------------------------------------------------

	result, err := client.Solve(id, *rate, *pickupMin, *pickupMax, actions)
	if err != nil {
		log.Fatalf("Failed to submit test solution: %v", err)
	}
	log.Printf("Test result: %v", result)
}
