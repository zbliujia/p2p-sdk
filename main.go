package mypkg

import (
	"encoding/json"
	"time"
)

type Action interface {
	Do(id int, action string, t string, payload string)
	Print(s string)
}

type Printer struct {

}

func (Printer) Print(log string)  {

}

type Counter struct {
	Value   int
	printer Printer
}

func (c *Counter) notify(id int, event string) {
	c.Value++
	c.printer.Print("Hello, World!")
}

func (c *Counter) Parse(data string) {
	c.printer.Print("parse begin")
	go func() {
		time.Sleep(time.Second * 5)
		result := map[string]interface{}{}
		err := json.Unmarshal([]byte(data), &result)
		if err != nil {
			c.printer.Print(err.Error())
		} else {
			c.printer.Print("ok")
		}
	}()
}

func NewCounter(p Printer) *Counter {
	return &Counter{5, p}
}
