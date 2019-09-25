package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	ts "github.com/0987363/tcp_server"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const loggerKey = "Logger"
const RequestIDKey = "RequestID"

func DBConnector() ts.HandlerFunc {
	return func(c *ts.Context) {
		fmt.Println("Init db.")
		c.Next()
	}
}

func Logger() ts.HandlerFunc {
	return func(c *ts.Context) {
		logger := logrus.New().WithField(RequestIDKey, uuid.NewV4().String())
		c.Set(loggerKey, logger)
		c.Next()
	}
}

func GetLogger(c *ts.Context) *logrus.Entry {
	if logger, ok := c.Get(loggerKey); ok {
		return logger.(*logrus.Entry)
	}

	return nil
}

func main() {
	server := ts.New("localhost:9999")
	server.SetUdpProc(1)

	/*
		server.Use(func(c *ts.Context)  {
			logger := logrus.New()
			logger.Info("Init global logger.")
			c.Set("logger-global", logger)
			 c.Next()
		})
	*/
	server.Use(DBConnector())
	server.Use(Logger())

	server.OnConnectionOpen(func(c *ts.Context) {
		logger := GetLogger(c)
		logger.Info("connection opended")

		c.Set("time_begin", time.Now())

		return
	})

	server.OnNewMessage(func(c *ts.Context) ([]byte, error) {
		message := c.GetData()
		logger := GetLogger(c)

		if res := strings.Compare(string(message), "Test Message"); res != 0 {
			fmt.Println("failed msg:", string(message), res, len(message))
			//			return errors.New("Compair failed:" + string(message))
		}
		c.Set("device", string(message))
		logger.Debug("event keys:", c.Keys)
		logger.Debug("recv:", string(message), len(message))

		logger.Info("middware logger found.")
		logger.Info("remote: ", c.RemoteAddr())
//		time.Sleep(time.Second)

		c.Trim(len(message))
		return []byte("hello world"), nil
		//		return errors.New("Compair failed.")
	})

	server.OnConnectionClosed(func(c *ts.Context) {
		logger := GetLogger(c)

		if t, ok := c.Get("time_begin"); ok {
			begin := t.(time.Time)
			logger.Info("spend :", time.Now().Sub(begin))
		}

		for _, err := range c.Errors {
			if err == io.EOF {
				logger.Info("connection normal closed")
				continue
			}
			if err, ok := err.(net.Error); ok {
				if err.Timeout() {
					logger.Warning("connection timeout")
					continue
				}
				if ts.IsErrConnReset(err) {
					logger.Info("connection normal closed by remote ")
					continue
				}
			}
			logger.Error("connection err:", err)
		}
	})

	go server.Listen()

	time.Sleep(10 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		log.Fatal("Failed to connect to test server")
	}
	conn.Write([]byte("Test "))
	time.Sleep(10 * time.Millisecond)
	//	conn.Write([]byte("Message"))
	time.Sleep(10 * time.Millisecond)
	//	conn.Write([]byte("Te"))
	time.Sleep(10 * time.Millisecond)
	//	conn.Write([]byte("st Message"))
	time.Sleep(10 * time.Millisecond)
	conn.Close()

	time.Sleep(time.Second)
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		go func(i int) {
			conn, err = net.Dial("udp", "localhost:9999")
			if err != nil {
				log.Fatal("Failed to connect to test server")
			}
			conn.Write([]byte("i am udp:" + strconv.Itoa(i)))
			time.Sleep(10 * time.Millisecond)
			conn.Close()
		}(i)
	}

	time.Sleep(time.Second)

}
