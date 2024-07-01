package ordermgr

import (
	css "challenge/client"
	"log"
	"sync"
	"time"
)

type OrderTs struct {
	Timestamp int64
	Order     css.Order
}

type OrderMgr struct {
	max_hot_orders,	max_cold_orders, max_room_orders int
	min_pickup, max_pickup *time.Duration

	cold_orders, hot_orders, room_orders []*OrderTs
	discarded_orders map[string]*OrderTs

	ordersChan chan *OrderTs 

	coldMut, hotMut, roomMut, dMut sync.Mutex
}

func (o *OrderMgr) Init(min_pickup *time.Duration, max_pickup *time.Duration, max_hot, max_cold, max_room int) {
	o.min_pickup = min_pickup
	o.max_pickup = max_pickup

	o.max_cold_orders = max_cold
	o.max_hot_orders = max_hot
	o.max_room_orders = max_room

	o.cold_orders = make([]*OrderTs, max_cold)
	o.hot_orders  = make([]*OrderTs, max_hot)
	o.room_orders = make([]*OrderTs, max_room)
	o.discarded_orders = make(map[string]*OrderTs)

	o.ordersChan = make(chan *OrderTs, 100)
}

func (o *OrderMgr) PlaceOrders(orders *[]css.Order, rate *time.Duration, wg *sync.WaitGroup, actions *[]css.Action) {
	for _, order := range *orders {
		log.Printf("Received: %+v", order)
		wg.Add(1)
		
		go o.prepareOrder(order, actions)
		o.ordersChan <- &OrderTs{Timestamp: time.Now().Unix(), Order: order}

		time.Sleep(*rate)
	}
	close(o.ordersChan)
	wg.Done()
}

func (o *OrderMgr) PickupOrders(wg *sync.WaitGroup, actions *[]css.Action) {
	for orderTs := range o.ordersChan {
		go o.pickupOrder(orderTs, wg, actions)
	}
	wg.Done()
}

func (o *OrderMgr) prepareOrder(order css.Order, actions *[]css.Action) {
	switch order.Temp {
	case "hot":
		if !o.addHotOrder(order) {
			o.addOrderToShelf(order, actions)
			log.Println("Place: Hot order", order.ID, "to shelf")
		}
	case "cold":
		if !o.addColdOrder(order) {
			o.addOrderToShelf(order, actions)
			log.Println("Moved: Cold order", order.ID, "to shelf")
		}
	case "room":
		o.addOrderToShelf(order, actions)
	}
	*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: order.ID, Action: css.Place})

}

func (o *OrderMgr) addHotOrder(order css.Order) bool {
	o.hotMut.Lock()
	defer o.hotMut.Unlock()
	for i := 0; i < o.max_hot_orders; i++ {
		if o.hot_orders[i] == nil {
			o.hot_orders[i] = &OrderTs{time.Now().Unix(), order}
			return true
		}
	}
	return false
}

func (o *OrderMgr) addColdOrder(order css.Order) bool {
	o.coldMut.Lock()
	defer o.coldMut.Unlock()
	for i := 0; i < o.max_cold_orders; i++ {
		if o.cold_orders[i] == nil {
			o.cold_orders[i] = &OrderTs{time.Now().Unix(), order}
			return true
		}
	}
	return false
}

func (o *OrderMgr) addOrderToShelf(order css.Order, actions *[]css.Action) {
	// Add order to any available slot on the shelf.
	// If shelf is full, check if any other temperature orders can be moved to thier respective storage
	// If not, figure out the best order to be discarded from the shelf.
	// We will discard the order on the shelf with least remaining freshness, i.e. the order that will go stale first

	var discardOrderIdx int
	var orderToDiscard *OrderTs
	currOrder := &OrderTs{time.Now().Unix(), order}

	o.roomMut.Lock()
	defer o.roomMut.Unlock()
	for i := 0; i < o.max_room_orders; i++ {
		if o.room_orders[i] == nil {
			o.room_orders[i] = &OrderTs{time.Now().Unix(), order}
			return
		}
	}

	// Shelf is full, try to move an order from the shelf
	for i := 0; i < o.max_room_orders; i++ {
		// Check if any orders from the shelf can be moved to their correct temp storage
		rOrder := o.room_orders[i].Order
		switch rOrder.Temp {
		case "hot":
			if o.addHotOrder(rOrder) {
				*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: rOrder.ID, Action: css.Move})
				log.Println("Moved shelf order", rOrder.ID, "to hot storage")
				o.room_orders[i] = currOrder
				return
			}
		case "cold":
			if o.addColdOrder(rOrder) {
				*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: rOrder.ID, Action: css.Move})
				log.Println("Moved cold order", rOrder.ID, "to cold storage")
				o.room_orders[i] = currOrder
				return
			}
		}

		// Also, figure out the order with the least freshness
		// This will save us another iteration in case no orders can be moved and an order needs to be discarded
		if orderToDiscard == nil {
			orderToDiscard = o.room_orders[i]
		} else {
			if freshness(o.room_orders[i]) < orderToDiscard.Order.Freshness {
				orderToDiscard = o.room_orders[i]
				discardOrderIdx = i
			}
		}
	}

	*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: orderToDiscard.Order.ID, Action: css.Discard})

	o.room_orders[discardOrderIdx] = currOrder
	o.dMut.Lock()
	o.discarded_orders[orderToDiscard.Order.ID] = orderToDiscard
	o.dMut.Unlock()

	log.Printf("Discarded: %+v", orderToDiscard.Order)
}

func freshness(rOrder *OrderTs) int {
	return rOrder.Order.Freshness - int(time.Now().Unix()-rOrder.Timestamp)
}

func (o *OrderMgr) pickupOrder(orderTs *OrderTs, wg *sync.WaitGroup, actions *[]css.Action) {
	// Calculate the pickup window
	tElapsed := time.Now().Unix() - orderTs.Timestamp
	time.Sleep(*o.min_pickup - time.Duration(tElapsed))	// Adjust for time elapsed between order place and pickup time

	if !o.processDiscardedOrder(orderTs.Order) {
		switch orderTs.Order.Temp {
		case "hot":
			if !o.pickupHotOrder(orderTs.Order) {
				// Order was not found in hot storage, check the shelf
				o.pickupShelfOrder(orderTs.Order)
			}
		case "cold":
			if !o.pickupColdOrder(orderTs.Order) {
				// Order was not found in cold storage, check the shelf
				o.pickupShelfOrder(orderTs.Order)
			}
		case "room":
			o.pickupShelfOrder(orderTs.Order)
		}
		log.Printf("Picked: %+v", orderTs.Order)
		*actions = append(*actions, css.Action{Timestamp: time.Now().UnixMicro(), ID: orderTs.Order.ID, Action: css.Pickup})
	}
	wg.Done()

}

func (o *OrderMgr) processDiscardedOrder(order css.Order) bool {
	o.dMut.Lock()
	defer o.dMut.Unlock()
	if _, ok := o.discarded_orders[order.ID]; ok {
		delete(o.discarded_orders, order.ID)
		return true
	}
	return false
}

func (o *OrderMgr) pickupHotOrder(order css.Order) bool {
	o.hotMut.Lock()
	defer o.hotMut.Unlock()
	for i := 0; i < len(o.hot_orders); i++ {
		if o.hot_orders[i] != nil {
			if o.hot_orders[i].Order.ID == order.ID {
				o.hot_orders[i] = nil
				return true
			}
		}
	}
	return false
}

func (o *OrderMgr) pickupColdOrder(order css.Order) bool {
	o.coldMut.Lock()
	defer o.coldMut.Unlock()
	for i := 0; i < len(o.cold_orders); i++ {
		if o.cold_orders[i] != nil {
			if o.cold_orders[i].Order.ID == order.ID {
				o.cold_orders[i] = nil
				return true
			}
		}
	}
	return false
}

func (o *OrderMgr) pickupShelfOrder(order css.Order) {
	o.roomMut.Lock()
	defer o.roomMut.Unlock()
	for i := 0; i < len(o.room_orders); i++ {
		if o.room_orders[i] != nil {
			if o.room_orders[i].Order.ID == order.ID {
				o.room_orders[i] = nil
				return
			}
		}
	}
	log.Println("ERROR: Shelf order not found. OrderId:", order.ID)
}
