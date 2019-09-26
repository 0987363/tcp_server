package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
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
		defer fmt.Println("Close db")
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

func Auth() ts.HandlerFunc {
	return func(c *ts.Context) {
		logger := GetLogger(c)
		logger.Info("auth opended")
		c.Next()

		logger.Info("auth closed, err:", len(c.Errors))
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
			logger.Error("found err:", err)
		}

		logger.Info("msg index:", c.MsgIndex())

	}
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
	server.Use(Auth())

	/*
	server.OnConnectionOpen(func(c *ts.Context) {
		logger := GetLogger(c)
		logger.Info("connection opended")

		c.Set("time_begin", time.Now())

		return
	})
	*/

	server.OnNewMessage(func(c *ts.Context) {
		message := c.ReadData()
		logger := GetLogger(c)

		c.Set("device", string(message))
		logger.Info("recv:", string(message), len(message))

		logger.Info("remote: ", c.RemoteAddr())
		//		time.Sleep(time.Second)

		c.Trim(len(message))
		//		return []byte("hello world"), nil
		return 
		c.AbortWithError(errors.New("Compair failed."))
	})

	/*
	server.OnConnectionClosed(func(c *ts.Context) {
		logger := GetLogger(c)

		if t, ok := c.Get("time_begin"); ok {
			begin := t.(time.Time)
			logger.Info("spend :", time.Now().Sub(begin))
		}

		logger.Info("connection closed, err:", len(c.Errors))
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
			logger.Error("found err:", err)
		}
	})
	*/

	go server.Listen()

	time.Sleep(10 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		log.Fatal("Failed to connect to test server")
	}
	conn.Write([]byte("Test "))
	time.Sleep(10 * time.Millisecond)
	conn.Write([]byte("Message"))
	time.Sleep(10 * time.Millisecond)
	//	conn.Write([]byte("Te"))
	time.Sleep(10 * time.Millisecond)
	//	conn.Write([]byte("st Message"))
	time.Sleep(10 * time.Millisecond)
	conn.Close()

	time.Sleep(time.Second)
	time.Sleep(time.Second)

	for i := 0; i < 1; i++ {
		go func(i int) {
			conn, err = net.Dial("udp", "localhost:9999")
			if err != nil {
				log.Fatal("Failed to connect to test server")
			}
			conn.Write([]byte("i am udp:" + strconv.Itoa(i)))
			time.Sleep(10 * time.Millisecond)
			conn.Write([]byte("udp is good."))
			time.Sleep(10 * time.Millisecond)
			conn.Close()
		}(i)
	}

	time.Sleep(time.Second)

}
