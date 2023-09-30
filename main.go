package mypkg

import (
	"fmt"
	ma "github.com/multiformats/go-multiaddr"
	"io"
	"log"
)

type Action interface {
	Print(s string)
}

func relay(dst io.Writer, src io.Reader, ch chan<- error, index int) {
	println("begin relay %d", index)
	_, err := io.Copy(dst, src)
	println("end relay %d", index)
	ch <- err
}

func Init(listenPort int, proxyPort int, remoteAddr string) string {
	host := makeRandomHost("127.0.0.1", listenPort)
	destPeerID := addAddrToPeerstore(host, remoteAddr)
	println(destPeerID)
	proxyAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", proxyPort))
	if err != nil {
		log.Fatalln(err)
	}
	proxy := newProxyService(host, proxyAddr, destPeerID)
	go proxy.Serve() // serve hangs forever
	//l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", proxyPort))
	//if err != nil {
	//	return err.Error()
	//}
	//go func() {
	//	for {
	//		conn, err := l.Accept()
	//		action.Print("accept new conn")
	//		if err != nil {
	//			action.Print(fmt.Sprintf("accept error %+v\n", err))
	//			return
	//		}
	//		go func() {
	//			stream := &innerTestStream{}
	//			ch := make(chan error)
	//
	//			go relay(conn, stream, ch, 1)
	//			go relay(stream, conn, ch, 2)
	//
	//			// 只要有一个结束 就认为结束了
	//			action.Print("begin wait response")
	//			err = <-ch
	//			action.Print("end wait response")
	//			if err != nil {
	//				action.Print(fmt.Sprintf("relay error %+v", err))
	//			}
	//			if err = conn.Close(); err != nil {
	//				action.Print(fmt.Sprintf("conn Close error %+v", err))
	//			}
	//			if err = stream.Close(); err != nil {
	//				action.Print(fmt.Sprintf("stream Close error %+v", err))
	//			}
	//		}()
	//	}
	//}()
	return "ok"
}

//type Action interface {
//	Do(id int, action string, t string, payload string)
//	Print(s string)
//}
//
//type Printer struct {
//
//}
//
//func (Printer) Print(log string)  {
//
//}
//
//type Counter struct {
//	Value   int
//	printer Printer
//}
//
//func (c *Counter) notify(id int, event string) {
//	c.Value++
//	c.printer.Print("Hello, World!")
//}
//
//func (c *Counter) Parse(data string) {
//	c.printer.Print("parse begin")
//	go func() {
//		time.Sleep(time.Second * 5)
//		result := map[string]interface{}{}
//		err := json.Unmarshal([]byte(data), &result)
//		if err != nil {
//			c.printer.Print(err.Error())
//		} else {
//			c.printer.Print("ok")
//		}
//	}()
//}
//
//func NewCounter(p Printer) *Counter {
//	return &Counter{5, p}
//}
