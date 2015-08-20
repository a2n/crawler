package crawler

import (
	"net/http"
	"fmt"
	"log"
	"runtime"
	"time"
	"io/ioutil"

	"github.com/a2n/alu"
)

/*
 * TODO
 *
 * - Make the concurrent requesting depends on bandwidth.
 *
 */

type Config struct {
	Email string
	URL string
	Concurrency int
}

type Crawler struct {
	logger *log.Logger

	config *Config
	user_agent string
	Response chan *Response
	queue []*http.Request

	// Clients
	clients []*client
	currentClient *client
}

func NewCrawler(config *Config) *Crawler {
	runtime.GOMAXPROCS(runtime.NumCPU())

	c := &Crawler {
		user_agent: fmt.Sprintf("%s %s", config.Email, config.URL),
		Response: make(chan *Response, config.Concurrency),
		queue: make([]*http.Request, 0),
		config: config,
		logger: alu.NewLogger("crawler.log"),
	}

	/*
	 * TODO
	 * The transport of client supports MaxIdleConnsPerHost property.
	 * It controls the concurrent connections,use few transports instead of 
	 * clients.
	 *
	 */
	// Init clients
	c.clients = make([]*client, 0)
	for i := 0; i < config.Concurrency; i++ {
		client := NewClient(c.logger, uint8(i), c.Response)
		c.clients = append(c.clients, client)
	}

	currentClient := c.clients[0]
	for i := 0; i < len(c.clients); i++ {
		if i != len(c.clients) - 1{
			c.clients[i].Next = c.clients[i + 1]
		} else {
			c.clients[i].Next = c.clients[0]
		}
	}
	c.currentClient = currentClient

	return c
}

func (c *Crawler) Push(r *http.Request) {
	r.Header.Add("User-Agent", c.user_agent)
	c.currentClient.Push(r)
	c.currentClient = c.currentClient.Next
}

type  client struct {
	logger *log.Logger
	index uint8
	client *http.Client
	queue []*http.Request
	Response chan *Response
	Next *client
}

func NewClient(l *log.Logger, idx uint8, ch chan *Response) *client {
	c := &client {
		logger: l,
		index: idx,
		client: &http.Client{},
		queue: make([]*http.Request, 0),
		Response: ch,
	}

	go func() {
		for {
			select {
			case <- time.Tick(250 * time.Microsecond):
				if len(c.queue) > 0 {
					c.fire()
				}
			}
		}
	}()

	return c
}

func (c *client) Push(r *http.Request) {
	c.queue = append(c.queue, r)
}

type Response struct {
	Header http.Header
	Body []byte
	Request *http.Request
	StatusCode int
}

func (c *client) fire() {
	req := c.queue[0]

	c.logger.Printf("client %d, begin to request %s.", c.index, req.URL.String())
	resp, err := c.client.Do(req)
	c.queue = append(c.queue[:0], c.queue[1:]...)
	if err != nil {
		c.logger.Printf("client %d has error, %s.", c.index, err.Error())
		return
	}
	c.logger.Printf("client %d, finished requesting %s.", c.index, req.URL.String())

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		c.logger.Printf("client %d has error, %s.", c.index, err.Error())
		return
	}

	my_resp := &Response {
		Header: resp.Header,
		Body: b,
		Request: resp.Request,
		StatusCode: resp.StatusCode,
	}

	c.Response <- my_resp
}
