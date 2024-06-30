package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Order is a json-friendly representation of an order.
type Order struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Temp      string `json:"temp"`
	Freshness int    `json:"freshness"`
}

// Action names
const (
	Place   = "place"
	Move    = "move"
	Pickup  = "pickup"
	Discard = "discard"
)

// Action is a json-friendly representation of an action.
type Action struct {
	Timestamp int64  `json:"timestamp"` // unix timestamp in microseconds
	ID        string `json:"id"`        // order id
	Action    string `json:"action"`    // place, move, pickup or discard
}

type options struct {
	Rate int64 `json:"rate"` // inverse rate in microseconds
	Min  int64 `json:"min"`  // min pickup in microseconds
	Max  int64 `json:"max"`  // max pickup in microseconds
}

type solution struct {
	Options options  `json:"options"`
	Actions []Action `json:"actions"`
}

// Client is a client for fetching and solving challenge test problems.
type Client struct {
	endpoint, auth string
}

func NewClient(endpoint, auth string) *Client {
	return &Client{endpoint: endpoint, auth: auth}
}

// New fetches a new test problem from the server. The URL also works in a browser for convenience.
func (c *Client) New(name string, seed int64) (string, []Order, error) {
	if seed == 0 {
		seed = rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	}

	url := fmt.Sprintf("%v/interview/challenge/new?auth=%v&name=%v&seed=%v", c.endpoint, c.auth, name, seed)

	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("%v: %v", url, resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read body: %v", err)
	}

	var orders []Order
	if err := json.Unmarshal(buf, &orders); err != nil {
		return "", nil, fmt.Errorf("failed to deserialize '%v': %v", string(buf), err)
	}
	id := resp.Header.Get("x-test-id")

	fmt.Println("Orders: ", orders)
	log.Printf("Fetched new test problem, id=%v: %v", id, url)

	// orders := []Order{
	// 	{"g9ib6", "Yoghurt", "cold", 113},
	// 	{"xp1iu", "Gas Station Sushi", "room", 3},
	// 	{"84ira", "Cookie Dough", "cold", 133},
	// 	{"qcfnz", "Danish Pastry", "room", 86},
	// 	{"5txkd", "Bacon Burger", "hot", 116},
	// 	{"c37x8", "Cookie Dough", "cold", 81},
	// 	{"53ruw", "Cheeseburger", "hot", 151},
	// 	{"qqwqr", "Dry Biscuits", "room", 12},
	// 	{"dckgp", "Cheese Pizza", "hot", 157},
	// 	{"biy6k", "Bacon Burger", "hot", 131},
	// 	//{ru5x9 Whole Wheat Bread room 80} {p5roq Strawberries room 98} {aenrf Kale Salad cold 170} {7wa86 Danish Pastry room 64} {wutw1 Vanilla Ice Cream cold 73} {69hdh Pad See Ew hot 172} {ujop4 Strawberries room 152} {m8rpd Orange Sherbet cold 171} {mtcxh Yoghurt cold 175} {8j56q French Fries hot 82} {1xzu5 Sushi cold 126} {gtmfh Danish Pastry room 65} {ztmkx Whole Wheat Bread room 115} {h9bac Hamburger hot 103} {581sf Banana room 174} {ei3mx Strawberry Ice Cream cold 128} {phfib Danish Pastry room 95} {4eomn Turkey Sandwich cold 97} {euahp Fish Tacos hot 99} {yf1sh Mixed Greens cold 103} {tp4mu Strawberry Ice Cream cold 64} {6ptbj Popsicle cold 115} {cb31s Gas Station Sushi room 130} {qwd5m Apple room 131} {qzc5f Spaghetti hot 119} {g39tr Mixed Greens cold 159} {aingb Burrito hot 83} {5cpox Italian Meatballs hot 118} {sxg5w Raspberries room 67} {tjouz Pastrami Sandwich cold 136} {xa9p4 Tuna Sandwich cold 102} {h8bcn Pressed Juice cold 140} {z466y Lukewarm Coke room 103} {9pjzb Strawberry Ice Cream cold 176} {56yuw Tomato Soup hot 93} {txffa BBQ Pizza hot 68} {8cxpd Spaghetti hot 92} {417oj Gas Station Sushi room 96}]
	// }
	return id, orders, nil
}

// Solve submits a sequence of actions and parameters as a solution to a test problem. Returns test result.
func (c *Client) Solve(id string, rate, min, max time.Duration, actions []Action) (string, error) {
	url := fmt.Sprintf("%v/interview/challenge/solve?auth=%v", c.endpoint, c.auth)

	payload := solution{
		Options: options{
			Rate: rate.Microseconds(),
			Min:  min.Microseconds(),
			Max:  max.Microseconds(),
		},
		Actions: actions,
	}

	fmt.Println("Payload: ", payload)

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Add("x-test-id", id)
	req.Header.Add("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%v: %v", url, resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(buf), nil

	// return string(body), nil
}
