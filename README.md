# Crawler
* A web crawler.


# Features
 - [x] Concurrent requesting.


# Usage
```
import (
	"github.com/a2n/crawler"
	"fmt"
)

func main() {
	c := crawler.NewCrawler(&crawler.CrawlerConfig {                               
        Email: "example@example.com",                         
        URL: "https://www.example.com",                                     
    })
    
    r, _ = http.NewRequest("GET", "https://www.google.com", nil)                   
    c.Push(r)
    
    r, _ = http.NewRequest("GET", "https://tw.yahoo.com", nil)                   
    c.Push(r)
    
    for {                                                                          
        select {                                                                   
        case r := <-c.ResponseChannel:                                             
            fmt.Printf("%s (%d).\n", r.URL, r.Response.StatusCode)                 
        }                                                                          
    }
}

```

#License
Crawler is released under the CC0 1.0 Universal.