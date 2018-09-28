# go-checkr

URL crawler to check url status in a webpage. Uses breadth-first search and runs concurrently to search for links.

## Usage

```bash
# Example website with anchor links.
$ go run server.go

# Start the crawler.
$ go run main.go
```

Output:

```
2018/09/29 01:42:56 found: http://localhost:8080
2018/09/29 01:42:56 found: http://localhost:8080/test
2018/09/29 01:42:56 found: http://localhost:8080/unknown
URL                              |Frequency    |Success
---                              |---          |---
http://localhost:8080            |2            |true
http://localhost:8080/test       |1            |true
http://localhost:8080/unknown    |1            |true
```
