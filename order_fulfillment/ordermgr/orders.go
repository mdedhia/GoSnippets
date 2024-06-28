package ordermgr

import (
	css "challenge/client"
	"fmt"
	"sync"
	"time"
)

type readyOrder struct {
	Timestamp 	int64
	Order		css.Order
}

var (
	max_hot_orders 	= 6
	max_cold_orders = 6
	max_room_orders = 12

	cold_orders = make([]*readyOrder, max_cold_orders)
	hot_orders 	= make([]*readyOrder, max_cold_orders)
	room_orders = make([]*readyOrder, max_room_orders)
)

var placeOrder = make(chan css.Order, 50)
var pickupOrder = make(chan css.Order)


func PlaceOrder(order css.Order, actions []css.Action) {
	placeOrder <- order
	actions = append(actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Place})

}

func PickupOrder(order css.Order, wg *sync.WaitGroup) {
	defer wg.Done()
}

func prepareOrder(order css.Order, actions []css.Action) error {
	actions = append(actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Pickup})
	switch order.Temp {
	case "hot":
		if !addOrder(order, "hot") {
			addOrderToShelf(order, actions)
		}
	case "cold":
		if !addOrder(order, "cold") {
			addOrderToShelf(order, actions)
		}
	case "room":
		addOrderToShelf(order, actions)
	}
	return nil
}

func addOrder(order css.Order, temp string) bool {
	switch temp {
	case "hot":
		for i := 0; i < max_hot_orders; i++ {
			if hot_orders[i] == nil {
				hot_orders[i] = &readyOrder{time.Now().Unix(), order}
				return true
			}
		}
	case "cold":
		for i := 0; i < max_cold_orders; i++ {
			if cold_orders[i] == nil {
				cold_orders[i] = &readyOrder{time.Now().Unix(), order}
				return true
			}
		}
	default:
		fmt.Println("Error: Invalid order temperature:", temp)
	}
	return false
}

func addOrderToShelf(order css.Order, actions []css.Action) {
	// Add order to any available slot on the shelf. 
	// If shelf is full, figure out the best order to be discarded from the shelf.
	// We will discard the order with least remaining freshness,
	// i.e. the order that will go stale first, and discard it
	discardOrderIdx := -1
	orderToDiscard := order		// Start with the current order, incase the curr order has the least freshness
	
	for i := 0; i < max_room_orders; i++ {
		if room_orders[i] == nil {
			room_orders[i] = &readyOrder{time.Now().Unix(), order}
			return
		}
		// While parsing the room_orders slice, also figure out the order with the least freshness
		// This will save us another iteration in case the shelf is full and an order needs to be discarded
		if freshness(room_orders[i]) < orderToDiscard.Freshness {
			orderToDiscard = room_orders[i].Order
			discardOrderIdx = i
		}
	}

	if discardOrderIdx >= 0 {
		room_orders[discardOrderIdx] =  &readyOrder{time.Now().Unix(), order}
	}

	actions = append(actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Discard})
}

func freshness(rOrder *readyOrder) int {
	return rOrder.Order.Freshness - int(time.Now().Unix() - rOrder.Timestamp)
}