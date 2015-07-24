package crawler

import (
	"net/http"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"time"
	"sync"
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
}

type Crawler struct {
	queue []*http.Request
	user_agent string
	Response chan *Response
	lock *sync.Mutex
}

const BUFFER_SIZE = 16
func NewCrawler(config *Config) *Crawler {
	runtime.GOMAXPROCS(BUFFER_SIZE)
	c := &Crawler {
		queue: make([]*http.Request, 0),
		user_agent: fmt.Sprintf("%s %s", config.Email, config.URL),
		Response: make(chan *Response),
		lock: &sync.Mutex{},
	}
	go c.tick()
	return c
}

func (c *Crawler) Length() int {
	return len(c.queue)
}

type Response struct {
	URL *url.URL
	Header http.Header
	Body []byte
	StatusCode int
}

func (c *Crawler) Push(r *http.Request) {
	go func(r *http.Request) {
		c.lock.Lock()
		r.Header.Add("User-Agent", c.user_agent)
		c.queue = append(c.queue, r)
		c.lock.Unlock()
	}(r)
}

func (c *Crawler) tick() {
	for {
		select {
		case <-time.Tick(time.Millisecond * 250):
			if len(c.queue) > 0 {
				r := c.queue[0]
				log.Printf("%s begin requesting %s.", alu.Caller(), r.URL.String())
				resp, err := http.DefaultClient.Do(r)
				if err != nil {
					log.Printf("%s has error, %s.", alu.Caller(), err.Error())
				}
				log.Printf("%s end requesting %s.", alu.Caller(), r.URL.String())

				c.lock.Lock()
				c.queue = append(c.queue[:0], c.queue[1:]...)
				c.lock.Unlock()

				b, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()

				my_resp := &Response {
					URL: r.URL,
					Header: resp.Header,
					Body: b,
					StatusCode: resp.StatusCode,
				}

				c.Response <-my_resp
			}
		}
	}
}
