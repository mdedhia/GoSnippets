package ordermgr

import (
	css "challenge/client"
	"fmt"
	"sync"
	"time"
	"testing"
)

func TestHotOrders(t *testing.T) {
	orders := []css.Order{
		{"id0", "Name123", "hot", 131},
		{"id1", "Name123", "hot", 113},
		{"id2", "Name123", "hot", 143},
		{"id3", "Name123", "hot", 133},
		{"id4", "Name123", "hot", 86},
		{"id5", "Name123", "room", 151},
		{"id6", "Name123", "hot", 116},		// Hot storage is now full
		{"id7", "Name123", "hot", 81},		// This order should be moved to the shelf 
		{"id8", "Name123", "room", 12},		// Order with least "Freshness", should get discarded eventually
		{"id9", "Name123", "room", 157},	
		{"id10", "Name123", "room", 19},
		{"id11", "Name123", "room", 157},
		{"id12", "Name123", "room", 151},
		{"id13", "Name123", "room", 157},
		{"id14", "Name123", "room", 171},	// First order (id0) about to be picked up right about here
		{"id15", "Name123", "room", 157},
		{"id16", "Name123", "room", 183},
		{"id17", "Name123", "room", 157},	// Shelf is now full
		{"id18", "Name123", "room", 42},	// Try to move hot order (id7) from the shelf to hot storage 
		{"id19", "Name123", "room", 157},	
		{"id20", "Name123", "room", 157},	// Shelf and hot storage full, (id8) should get discarded as it has least "Freshness"
	}

	wantActions := []css.Action{
		{0, "id0", "place"}, {0, "id1", "place"}, {0, "id2", "place"}, {0, "id3", "place"}, {0, "id4", "place"}, {0, "id5", "place"}, {0, "id6", "place"}, 
		{0, "id7", "place"}, {0, "id8", "place"}, {0, "id9", "place"}, {0, "id10", "place"}, {0, "id11", "place"}, {0, "id12", "place"}, {0, "id13", "place"}, 
		{0, "id14", "place"}, {0, "id15", "place"}, {0, "id16", "place"}, {0, "id17", "place"}, {0, "id18", "place"}, {0, "id19", "place"}, {0, "id20", "place"},
		
		{0, "id0", "pickup"}, {0, "id1", "pickup"}, {0, "id2", "pickup"}, {0, "id3", "pickup"}, {0, "id4", "pickup"}, {0, "id5", "pickup"}, {0, "id6", "pickup"}, 
		{0, "id7", "pickup"},  {0, "id9", "pickup"}, {0, "id10", "pickup"}, {0, "id11", "pickup"}, {0, "id12", "pickup"}, {0, "id13", "pickup"}, 
		{0, "id14", "pickup"}, {0, "id15", "pickup"}, {0, "id16", "pickup"}, {0, "id17", "pickup"}, {0, "id18", "pickup"}, {0, "id19", "pickup"}, {0, "id20", "pickup"},
		
		{0, "id7", "move"}, 
		
		{0, "id8", "discard"}, 
	}
	
	rate := time.Duration(10*time.Millisecond)
	min := time.Duration(150*time.Millisecond)
	max := time.Duration(250*time.Millisecond)

	var orderMgr OrderMgr 
	orderMgr.Init(&min, &max, 6, 6, 12)

	wg := sync.WaitGroup{}
	wg.Add(2)
	var actions []css.Action
	
	go orderMgr.PlaceOrders(&orders, &rate, &wg, &actions)
	go orderMgr.PickupOrders(&wg, &actions)

	wg.Wait()

	if !verifyOutput(wantActions, actions) {
		t.Errorf("PlaceOrders(): want actions: \n %q \n not equal to got actions: \n %q", wantActions, actions)
	}
}

func TestColdOrders(t *testing.T) {
	orders := []css.Order{
		{"id0", "Name123", "cold", 131},
		{"id1", "Name123", "cold", 113},
		{"id2", "Name123", "cold", 143},
		{"id3", "Name123", "cold", 133},
		{"id4", "Name123", "cold", 86},
		{"id5", "Name123", "cold", 116},	// cold storage is now full
		{"id6", "Name123", "cold", 81},		// This order should be moved to the shelf 
		{"id7", "Name123", "room", 151},
		{"id8", "Name123", "room", 186},
		{"id9", "Name123", "room", 157},
		{"id10", "Name123", "room", 134},
		{"id11", "Name123", "room", 157},
		{"id12", "Name123", "room", 165},
		{"id13", "Name123", "room", 157},
		{"id14", "Name123", "room", 122},	// First order about to be picked up
		{"id15", "Name123", "room", 157},
		{"id16", "Name123", "room", 11},
		{"id17", "Name123", "room", 157},	// Shelf is now full
		{"id18", "Name123", "room", 121},	// Try to move one of the cold orders (id6) from the shelf to cold storage 
		{"id19", "Name123", "room", 22},	// Shelf and cold storage full, (id16) should get discarded as it has least "Freshness"
	}

	wantActions := []css.Action{
		{0, "id0", "place"}, {0, "id1", "place"}, {0, "id2", "place"}, {0, "id3", "place"}, {0, "id4", "place"}, {0, "id5", "place"}, {0, "id6", "place"}, 
		{0, "id7", "place"}, {0, "id8", "place"}, {0, "id9", "place"}, {0, "id10", "place"}, {0, "id11", "place"}, {0, "id12", "place"}, {0, "id13", "place"}, 
		{0, "id14", "place"}, {0, "id15", "place"}, {0, "id16", "place"}, {0, "id17", "place"}, {0, "id18", "place"}, {0, "id19", "place"},
		
		{0, "id0", "pickup"}, {0, "id1", "pickup"}, {0, "id2", "pickup"}, {0, "id3", "pickup"}, {0, "id4", "pickup"}, {0, "id5", "pickup"}, {0, "id6", "pickup"}, 
		{0, "id7", "pickup"},  {0, "id8", "pickup"}, {0, "id9", "pickup"}, {0, "id10", "pickup"}, {0, "id11", "pickup"}, {0, "id12", "pickup"}, {0, "id13", "pickup"}, 
		{0, "id14", "pickup"}, {0, "id15", "pickup"}, {0, "id17", "pickup"}, {0, "id18", "pickup"}, {0, "id19", "pickup"}, 
		
		{0, "id6", "move"}, 
		
		{0, "id16", "discard"}, 
	}
	
	rate := time.Duration(10*time.Millisecond)
	min := time.Duration(150*time.Millisecond)
	max := time.Duration(250*time.Millisecond)

	var orderMgr OrderMgr 
	orderMgr.Init(&min, &max, 6, 6, 12)

	wg := sync.WaitGroup{}
	wg.Add(2)
	var actions []css.Action
	
	go orderMgr.PlaceOrders(&orders, &rate, &wg, &actions)
	go orderMgr.PickupOrders(&wg, &actions)

	wg.Wait()

	if !verifyOutput(wantActions, actions) {
		t.Errorf("PlaceOrders(): want actions: \n %q \n not equal to got actions: \n %q", wantActions, actions)
	}
}

func verifyOutput(wantActions, gotActions []css.Action) bool {
	wantActionsMap := make(map[string]int)

	// Populate wantActions into orderActions map
	for _, action := range wantActions {
		wantActionsMap[genKey(action)]++
	}

	for _, action := range gotActions {
		key := genKey(action)
		if count, ok := wantActionsMap[key]; ok {
			if count == 1 {
				delete(wantActionsMap, key)
			} else {
				wantActionsMap[key]--
			}
		} else {
			fmt.Println("Error: Got action:", key, "does not exist want actions")
			return false
		}
	}

	return len(wantActionsMap) == 0
}

func genKey(action css.Action) string {
	return fmt.Sprintf("%s:%s", action.ID, action.Action)
}