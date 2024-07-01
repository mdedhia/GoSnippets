# README

## Building and testing
### How to run

Install go `1.22.2` or later from [go.dev](https://go.dev/dl/) and run

```
$ go run main.go --auth=<token> --rate=500ms --pickup-min=4s --pickup-max=8s
```

Params:
1. auth (required): Auth token for communicating with the test server
2. rate (optional): Inverse order rate. Eg values: "300ms" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
3. pickup-min (optional): Minimum pickup time. Value options same as "rate" 
4. pickup-max (optional): Maximum pickup time. Value options same as "rate" 


## Order Manager

The `ordermgr` package contains the logic for placing, picking up, moving and discarding orders. 

### Discard logic
If a new order cannot be placed in either the hot/cold storage or on the shelf, a decision needs to be made to discard an existing order from the shelf.

The program uses the "Freshness" field to make this decision. The shelf order with the least remaining freshness will be discarded to make space for the new order. 

`TODO: Should we always discard an existing order from the shelf to make room for the new order, or is the new order subject for immediate discarding without being "picked" up at all? Currently the test server doesn't seem to allow for the new order to be discarded if that has the least amount of freshness remaining.` 

#### Why this approach?
We want use a greedy approach to maximize the number of orders that can be delivered while they're still fresh. That means discarding the order that's going stale the soonest, whilst giving the other orders a chance for pickup while they're still fresh

`TODO: Add logic to discard orders that are not fresh at pickup time. Doesn't look like the test server has this logic in place as of now`

## Testing

### How to test
```
$ go test challenge/ordermgr
```

The package includes a couple of cases that test the following functionality:
1. If hot/cold storage is full -> The order should be moved to the shelf if there is space
2. If the shelf is full -> Move a hot/cold order from the shelf to its corresponding temperature storage
3. If both hot/cold and shelf are full -> Discard an order from the shelf per above [Discard logic](#DiscardLogic)
