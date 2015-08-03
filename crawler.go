package crawler

import (
	"net/http"
	"fmt"
	"log"
	"net/url"
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
	queue []*http.Request
	user_agent string
	Response chan *Response
	config *Config
}

const BUFFER_SIZE = 16
func NewCrawler(config *Config) *Crawler {
	runtime.GOMAXPROCS(BUFFER_SIZE)
	c := &Crawler {
		queue: make([]*http.Request, 0),
		user_agent: fmt.Sprintf("%s %s", config.Email, config.URL),
		Response: make(chan *Response, config.Concurrency),
		config: config,
	}

	go c.tick()
	return c
}

type Response struct {
	URL *url.URL
	Header http.Header
	Body []byte
	StatusCode int
}

func (c *Crawler) Push(r *http.Request) {
	ch := make(chan bool)
	go func(r *http.Request) {
		r.Header.Add("User-Agent", c.user_agent)
		c.queue = append(c.queue, r)
		ch <- true
	}(r)
	<-ch
}

func (c *Crawler) tick() {
	for {
		select {
		case <-time.Tick(time.Millisecond * 250):
			for i := 0; i < c.config.Concurrency; i++ {
				if len(c.queue) > 0 {
					r := c.queue[0]
					c.queue = append(c.queue[:0], c.queue[1:]...)
					go c.fire(r)
				}
			}
		}
	}
}

func (c *Crawler) fire(r *http.Request) {
	log.Printf("%s begin requesting %s.", alu.Caller(), r.URL.String())
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Printf("%s has error, %s.", alu.Caller(), err.Error())
		return
	}
	log.Printf("%s end requesting %s.", alu.Caller(), r.URL.String())

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s has error, %s.", alu.Caller(), err.Error())
		return
	}

	err = resp.Body.Close()
	if err != nil {
		log.Printf("%s has error, %s.", alu.Caller(), err.Error())
		return
	}

	my_resp := &Response {
		URL: r.URL,
		Header: resp.Header,
		Body: b,
		StatusCode: resp.StatusCode,
	}

	c.Response <-my_resp
}
