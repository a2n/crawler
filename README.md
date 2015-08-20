# Crawler
* A web crawler.


# Usage
```
import (
	"github.com/a2n/crawler"
	"fmt"
)

func main() {
	c := crawler.NewCrawler(&crawler.Config {                               
        Email: "example@example.com",                         
        URL: "https://www.example.com",                                     
    })
    
    r, _ = http.NewRequest("GET", "https://www.google.com", nil)                   
    c.Push(r)
    
    r, _ = http.NewRequest("GET", "https://tw.yahoo.com", nil)                   
    c.Push(r)
    
    for {                                                                          
        select {                                                                   
        case r := <-c.Response:
            fmt.Printf("%s (%d).\n", r.Request.URL.String(), r.StatusCode)                 
        }                                                                          
    }
}

```

#License
Crawler is released under the CC0 1.0 Universal.
