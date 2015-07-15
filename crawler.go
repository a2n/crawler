package crawler

import (
	"net/http"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"time"
	"sync"

	"github.com/a2n/alu"
)

/*
 * TODO
 *
 * - Make the presistent connection if http 1.1 supported.
 *   Supported by default.
 * 
 * - Make the concurrent requesting depends on bandwidth.
 * - Register some channels for reading responses.
 *
 *
 * References
 *
 * - https://talks.golang.org/2013/advconc.slide#39
 *
 */

type CrawlerConfig struct {
	Email string
	URL string
}

type Crawler struct {
	queue []*http.Request
	client *http.Client
	user_agent string
	ResponseChannel chan *Response
	ltime map[string]time.Time
	lock *sync.Mutex
}

const BUFFER_SIZE = 16
func NewCrawler(config *CrawlerConfig) *Crawler {
	runtime.GOMAXPROCS(BUFFER_SIZE)
	c := &Crawler {
		queue: make([]*http.Request, 0),
		client: &http.Client{ Transport: &http.Transport{} },
		user_agent: fmt.Sprintf("%s %s", config.Email, config.URL),
		ResponseChannel: make(chan *Response),
		ltime: make(map[string]time.Time, 0),
		lock: &sync.Mutex{},
	}
	return c
}

func (c *Crawler) Length() int {
	return len(c.queue)
}

type Response struct {
	URL *url.URL
	Response *http.Response
}

func (c *Crawler) Push(r *http.Request) {
	go func(r *http.Request) {
		c.lock.Lock()
		c.queue = append(c.queue, r)
		c.lock.Unlock()

		r.Header.Add("User-Agent", c.user_agent)

		// Last access time.
		if !c.ltime[r.URL.Host].IsZero() {
			ltime := c.ltime[r.URL.Host]
			if ltime.Sub(time.Now()) <= (time.Microsecond * 200) {
				time.Sleep(ltime.Sub(time.Now()))
			}
		}
		c.ltime[r.URL.Host] = time.Now()

		log.Printf("%s begin requesting %s.", alu.Caller(), r.URL.String())
		resp, err := c.client.Do(r)
		if err != nil {
			log.Printf("%s has error, %s.", alu.Caller(), err.Error())
		}
		log.Printf("%s end requesting %s.", alu.Caller(), r.URL.String())

		// Delete the successfully requesting
		if resp.StatusCode == 200 {
			c.lock.Lock()
			c.queue = append(c.queue[:0], c.queue[1:]...)
			c.lock.Unlock()

			c.ResponseChannel <- &Response {
				URL: r.URL,
				Response: resp,
			}
		}
	}(r)
}
