package main

import (
	"fmt"
	"log"
	"net"
	"time"

	ts "github.com/0987363/tcp_server"
	"github.com/sirupsen/logrus"
)

func main() {
	server := ts.New("localhost:9999")

	server.Use(func(c *ts.Client) {
		fmt.Println("Init logger.")
		c.Set("logger", "logger....")
	})
	server.Use(func(c *ts.Client) {
		fmt.Println("Init db.")
		c.Next()
	})
	server.Use(func(c *ts.Client) {
		fmt.Println("Init micro.")
	})

	/*
		server.OnConnectionOpen(func(c *ts.Client) error {
			fmt.Println("start connection")
			return nil
		})
	*/

	server.OnNewMessage(func(c *ts.Client, message []byte) error {
		fmt.Println("recv:", string(message))
		if logger, ok := c.Get("logger").(*logrus.Logger); ok {
			logger.Info("middware logger")
		}
		return nil
	})
	server.OnConnectionClosed(func(c *ts.Client, err error) {
		fmt.Println("close, err:", err)
	})
	go server.Listen(nil)

	time.Sleep(10 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		log.Fatal("Failed to connect to test server")
	}
	conn.Write([]byte("Test message\n"))
	conn.Close()

	time.Sleep(10 * time.Millisecond)
}
