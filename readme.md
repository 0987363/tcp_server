# TS
Package tcp_server created to help build TCP/UDP servers faster. support tcp/udp/middleware.

### Install package

``` bash
go get github.com/0987363/tcp_server@master
```

### Usage:

``` go
package main

import ts "github.com/0987363/tcp_server"

func main() {
	server := ts.New("localhost:9999")
	server.SetUdpProc(1)

    server.Use(func(c *ts.Context)  {
        c.Set("logger", "logggg")
        c.Next()
    })

	server.OnNewMessage(func(c *ts.Context) {
		message := c.ReadData()

		v, ok := c.Get("logger")

		c.Trim(len(message))
		c.AbortWithError(errors.New("read failed."))
	})

    server.Listen()
}
```
