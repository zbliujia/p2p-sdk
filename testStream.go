package mypkg

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"
)

type innerTestStream struct {
}

func (s *innerTestStream) Close() error {
	return nil
}

func (s *innerTestStream) Write(p []byte) (n int, err error) {
	//println("begin Write")
	//println(string(p))
	//println("end Write")
	return len(p), nil
}

func (s *innerTestStream) Read(p []byte) (n int, err error) {
	println("begin Read %d", len(p))
	time.Sleep(time.Second * 2)
	println("read 1")
	resp, err := http.Get("http://10.250.24.211:8000/1.txt")
	println("read 2")
	var dump []byte
	if err != nil {
		println(err.Error())
		dump, _ = base64.StdEncoding.DecodeString("SFRUUC8xLjEgMjAwIE9LDQpBY2Nlc3MtQ29udHJvbC1BbGxvdy1PcmlnaW46ICoNCkFjY2VwdC1SYW5nZXM6IGJ5dGVzDQpDYWNoZS1Db250cm9sOiBwdWJsaWMsIG1heC1hZ2U9MA0KTGFzdC1Nb2RpZmllZDogRnJpLCAwOCBTZXAgMjAyMyAwNjozMTowMyBHTVQNCkVUYWc6IFcvIjQtMThhNzM3Y2E3NmUiDQpDb250ZW50LVR5cGU6IHRleHQvcGxhaW47IGNoYXJzZXQ9VVRGLTgNCkNvbnRlbnQtTGVuZ3RoOiA0DQpEYXRlOiBGcmksIDA4IFNlcCAyMDIzIDA5OjU1OjU5IEdNVA0KQ29ubmVjdGlvbjoga2VlcC1hbGl2ZQ0KS2VlcC1BbGl2ZTogdGltZW91dD01DQoNCjEyMwo=")
	} else {
		defer resp.Body.Close()
		dump, _ = httputil.DumpResponse(resp, true)
	}
	println("read 3")

	n = len(dump)
	copy(p, dump)
	println("read 4", n)
	println("end Read")
	fmt.Printf("%q", dump)
	err = io.EOF
	return
}
