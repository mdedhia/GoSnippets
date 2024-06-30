package ordermgr

import (
	css "challenge/client"
	"fmt"
	"log"
	"sync"
	"time"
)

type OrderTs struct {
	Timestamp int64
	Order     css.Order
}

var (
	max_hot_orders  = 6
	max_cold_orders = 6
	max_room_orders = 12

	cold_orders = make([]*OrderTs, max_cold_orders)
	hot_orders  = make([]*OrderTs, max_cold_orders)
	room_orders = make([]*OrderTs, max_room_orders)
	discardedOrders = make(map[string]*OrderTs)

	ordersChan = make(chan *OrderTs, 100)

	coldMut, hotMut, roomMut, dMut sync.Mutex
)

func PlaceOrders(orders *[]css.Order, rate *time.Duration, wg *sync.WaitGroup, actions *[]css.Action) {
	for _, order := range *orders {
		log.Printf("Received: %+v", order)
		wg.Add(1)
		go prepareOrder(order, actions)
		ordersChan <- &OrderTs{Timestamp: time.Now().Unix(), Order: order}
		*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Place})

		time.Sleep(*rate)
	}
	close(ordersChan)
	wg.Done()
}

func PickupOrders(min *time.Duration, max *time.Duration, wg *sync.WaitGroup, actions *[]css.Action) {
	for orderTs := range ordersChan {
		go pickupOrder(orderTs, min, max, wg, actions)
	}
	wg.Done()
}

func prepareOrder(order css.Order, actions *[]css.Action) {
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
}

func addOrder(order css.Order, temp string) bool {
	switch temp {
	case "hot":
		hotMut.Lock()
		defer hotMut.Unlock()
		for i := 0; i < max_hot_orders; i++ {
			if hot_orders[i] == nil {
				hot_orders[i] = &OrderTs{time.Now().Unix(), order}
				fmt.Println("Added order", order.ID, "to hot storage")
				return true
			}
		}
	case "cold":
		coldMut.Lock()
		defer coldMut.Unlock()
		for i := 0; i < max_cold_orders; i++ {
			if cold_orders[i] == nil {
				cold_orders[i] = &OrderTs{time.Now().Unix(), order}
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
	orderToDiscard := &OrderTs{time.Now().Unix(), order}
	currOrder := &OrderTs{time.Now().Unix(), order}

	roomMut.Lock()
	defer roomMut.Unlock()
	for i := 0; i < max_room_orders; i++ {
		if room_orders[i] == nil {
			room_orders[i] = &OrderTs{time.Now().Unix(), order}
			fmt.Println("Added order", order.ID, "to shelf")
			return
		}
	}

	for i := 0; i < max_room_orders; i++ {
		// Check if any orders from the shelf can be moved to their correct temp storage
		rOrder := room_orders[i].Order
		switch rOrder.Temp {
		case "hot":
			if addOrder(rOrder, "hot") {
				*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: rOrder.ID, Action: css.Move})
				fmt.Println("Moved order", rOrder.ID, "to hot storage")
				room_orders[i] = currOrder
				return
			}
		case "cold":
			if addOrder(rOrder, "cold") {
				*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: rOrder.ID, Action: css.Move})
				fmt.Println("Moved order", rOrder.ID, "to cold storage")
				room_orders[i] = currOrder
				return
			}
		}

		// While parsing the room_orders slice, also figure out the order with the least freshness
		// This will save us another iteration in case no orders can be moved and an order needs to be discarded
		if room_orders[i].Timestamp < orderToDiscard.Timestamp {
			orderToDiscard = room_orders[i]
			discardOrderIdx = i
		}
	}

	if discardOrderIdx >= 0 {
		room_orders[discardOrderIdx] = currOrder
	}
	dMut.Lock()
	discardedOrders[orderToDiscard.Order.ID] = orderToDiscard
	dMut.Unlock()
	*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: orderToDiscard.Order.ID, Action: css.Discard})

	fmt.Println("OrderID:", orderToDiscard.Order.ID, "Action: Discarded")
}

// func freshness(rOrder *orderTs) int {
// 	return rOrder.Order.Freshness - int(time.Now().Unix()-rOrder.Timestamp)
// }

func pickupOrder(orderTs *OrderTs, min *time.Duration, max *time.Duration, wg *sync.WaitGroup, actions *[]css.Action) {
	// Calculate the pickup window
	tElapsed := time.Now().Unix() - orderTs.Timestamp
	time.Sleep(*min - time.Duration(tElapsed))	// Adjust for time elapsed between order place and pickup time

	fmt.Println("Picking up order", orderTs.Order.ID)

	if processDiscardedOrder(orderTs.Order) {
		fmt.Println("PickupOrder(): Discarded order", orderTs.Order.ID)
		*actions = append(*actions, css.Action{Timestamp: orderTs.Timestamp, ID: orderTs.Order.ID, Action: css.Discard})
	} else {
		switch orderTs.Order.Temp {
		case "hot":
			if !pickupHotOrder(orderTs.Order) {
				// Order was not found in hot storage, check the shelf
				pickupShelfOrder(orderTs.Order)
			}
		case "cold":
			if !pickupColdOrder(orderTs.Order) {
				// Order was not found in cold storage, check the shelf
				pickupShelfOrder(orderTs.Order)
			}
		case "room":
			pickupShelfOrder(orderTs.Order)
		}

		if time.Now().Unix() - orderTs.Timestamp >= int64(max.Seconds()) {
			*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: orderTs.Order.ID, Action: css.Discard})
		} else {
			*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: orderTs.Order.ID, Action: css.Pickup})
		}
	}
	wg.Done()

}

func processDiscardedOrder(order css.Order) bool {
	dMut.Lock()
	defer dMut.Unlock()
	if _, ok := discardedOrders[order.ID]; ok {
		delete(discardedOrders, order.ID)
		return true
	}
	return false
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
