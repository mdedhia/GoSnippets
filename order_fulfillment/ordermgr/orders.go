package ordermgr

import (
	css "challenge/client"
	"math/rand"
	"fmt"
	"sync"
	"time"
)

type orderTs struct {
	Timestamp 	int64
	Order		css.Order
}

var (
	pickup_min = 4
	pickup_max = 8

	max_hot_orders 	= 6
	max_cold_orders = 6
	max_room_orders = 12

	cold_orders = make([]*orderTs, max_cold_orders)
	hot_orders 	= make([]*orderTs, max_cold_orders)
	room_orders = make([]*orderTs, max_room_orders)

	discardedOrders = make(map[string]*orderTs)

	coldMut, hotMut, roomMut sync.Mutex
)

func PlaceOrder(order css.Order, actions *[]css.Action) {
	fmt.Println("OrderID:", order.ID, "Action: Placed", )
	prepareOrder(order, actions)
	*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Place})
}

func prepareOrder(order css.Order, actions *[]css.Action) {
	*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Pickup})
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
	fmt.Println("OrderID:", order.ID, "Action: Prepared", )
}

func addOrder(order css.Order, temp string) bool {
	switch temp {
	case "hot":
		hotMut.Lock()
		defer hotMut.Unlock()
		for i := 0; i < max_hot_orders; i++ {
			if hot_orders[i] == nil {
				hot_orders[i] = &orderTs{time.Now().Unix(), order}
				fmt.Println("Added order", order.ID, "to hot storage")
				return true
			}
		}
	case "cold":
		coldMut.Lock()
		defer coldMut.Unlock()
		for i := 0; i < max_cold_orders; i++ {
			if cold_orders[i] == nil {
				cold_orders[i] = &orderTs{time.Now().Unix(), order}
				fmt.Println("Added order", order.ID, "to cold storage")
				return true
			}
		}
	default:
		fmt.Println("Error: Invalid order temperature:", temp)
	}
	return false
}

func addOrderToShelf(order css.Order, actions *[]css.Action) {
	// Add order to any available slot on the shelf. 
	// If shelf is full, figure out the best order to be discarded from the shelf.
	// We will discard the order with least remaining freshness,
	// i.e. the order that will go stale first, and discard it
	discardOrderIdx := -1
	orderToDiscard := order		// Start with the current order, incase the curr order has the least freshness
	currOrder := &orderTs{time.Now().Unix(), order}

	roomMut.Lock()
	defer roomMut.Unlock()
	for i := 0; i < max_room_orders; i++ {
		if room_orders[i] == nil {
			room_orders[i] = &orderTs{time.Now().Unix(), order}
			fmt.Println("Added order", order.ID, "to shelf")
			return
		}
	}

	for i := 0; i < max_room_orders; i++ {
		// Check if any orders from the shelf can be moved to their correct temp storage
		switch room_orders[i].Order.Temp {
		case "hot":
			if addOrder(order, "hot") {
				*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Move})
				room_orders[i] = currOrder
				return
			}
		case "cold":
			if addOrder(order, "cold") {
				*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Move})
				room_orders[i] = currOrder
				return
			}
		}

		// While parsing the room_orders slice, also figure out the order with the least freshness
		// This will save us another iteration in case no orders can be moved and an order needs to be discarded
		if freshness(room_orders[i]) < orderToDiscard.Freshness {
			orderToDiscard = room_orders[i].Order
			discardOrderIdx = i
		}
	}

	if discardOrderIdx >= 0 {
		discardedOrder := room_orders[discardOrderIdx].Order
		discardedOrders[discardedOrder.ID] = &orderTs{time.Now().Unix(), discardedOrder}
		*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: discardedOrder.ID, Action: css.Discard})

		room_orders[discardOrderIdx] = currOrder
	}

	fmt.Println("OrderID:", order.ID, "Action: Discarded", )

	*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Pickup})
}

func freshness(rOrder *orderTs) int {
	return rOrder.Order.Freshness - int(time.Now().Unix() - rOrder.Timestamp)
}

func PickupOrder(ordersChan chan css.Order, wg *sync.WaitGroup, actions *[]css.Action) {
	time.Sleep(time.Second * time.Duration(pickup_min))		// Initial wait time of pickup_min seconds

	for order := range ordersChan {
		fmt.Println("Picking up order", order.ID)
		if orderWTs, ok := discardedOrders[order.ID]; ok {
			fmt.Println("PickuOrder(): Discarded order", order.ID)

			*actions = append(*actions, css.Action{Timestamp: orderWTs.Timestamp, ID: orderWTs.Order.ID, Action: css.Discard})
		} else {			
			switch order.Temp {
			case "hot":
				if !pickupHotOrder(order) {
					// Order was not found in hot storage, check the shelf
					pickupShelfOrder(order)
				}
			case "cold":
				if !pickupColdOrder(order) {
					// Order was not found in cold storage, check the shelf
					pickupShelfOrder(order)
				}
			case "room":
				pickupShelfOrder(order)
			}
			*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Pickup})
		}
		wg.Done()
		randNum := rand.Intn(pickup_max - pickup_min + 1) + pickup_min	// Get a random number between pickup min-max
		time.Sleep(time.Second * time.Duration(randNum))
	}
}

func pickupHotOrder(order css.Order) bool {
	hotMut.Lock()
	defer hotMut.Unlock()
	for i := 0; i < len(hot_orders); i++ {
		if hot_orders[i] != nil {
			if hot_orders[i].Order.ID == order.ID {
				hot_orders[i] = nil
				return true
			}
		}
	}
	return false
}

func pickupColdOrder(order css.Order) bool {
	coldMut.Lock()
	defer coldMut.Unlock()
	for i := 0; i < len(cold_orders); i++ {
		if cold_orders[i] != nil {
			if cold_orders[i].Order.ID == order.ID {
				cold_orders[i] = nil
				return true
			}
		}
	}
	return false
}

func pickupShelfOrder(order css.Order) {
	roomMut.Lock()
	defer roomMut.Unlock()
	for i := 0; i < len(room_orders); i++ {
		if room_orders[i] != nil {
			if room_orders[i].Order.ID == order.ID {
				room_orders[i] = nil
				return
			}
		}
	}

	fmt.Println("ERROR: Shelf order not found. OrderId:", order.ID)
}
