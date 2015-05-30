package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {

	var show bool

	flag.BoolVar(&show, "show", false, "Display the response content")
	flag.Parse()

	url := flag.Args()[0]
	fmt.Println("URL:", url)

	tp := newTransport()
	client := &http.Client{Transport: tp}

	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("get error: %s: %s", err, url)
	}
	defer resp.Body.Close()

	output := ioutil.Discard
	if show {
		output = os.Stdout
	}
	io.Copy(output, resp.Body)

	log.Println("Duration:", tp.Duration())
	log.Println("Request duration:", tp.ReqDuration())
	log.Println("Connection duration:", tp.ConnDuration())

}

type customTransport struct {
	rtp       http.RoundTripper
	dialer    *net.Dialer
	connStart time.Time
	connEnd   time.Time
	reqStart  time.Time
	reqEnd    time.Time
}

func newTransport() *customTransport {

	tr := &customTransport{
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}
	tr.rtp = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                tr.dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return tr
}

func (tr *customTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tr.reqStart = time.Now()
	resp, err := tr.rtp.RoundTrip(r)
	tr.reqEnd = time.Now()
	return resp, err
}

func (tr *customTransport) dial(network, addr string) (net.Conn, error) {
	tr.connStart = time.Now()
	cn, err := tr.dialer.Dial(network, addr)
	tr.connEnd = time.Now()
	return cn, err
}

func (tr *customTransport) ReqDuration() time.Duration {
	return tr.Duration() - tr.ConnDuration()
}

func (tr *customTransport) ConnDuration() time.Duration {
	return tr.connEnd.Sub(tr.connStart)
}

func (tr *customTransport) Duration() time.Duration {
	return tr.reqEnd.Sub(tr.reqStart)
}
