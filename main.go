package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/cxjava/gscan"
	"github.com/elazarl/goproxy"
)

var (
	cdnChan         = make(chan string, 10)
	verbose         = flag.Bool("v", false, "Should every proxy request be logged to stdout")
	addr            = flag.String("a", ":8888", "The proxy listen address")
	size            = flag.Int("s", 20, "The fastest number of IP")
	dialTimeout     = flag.Int("dt", 1, "The timeout(Second) of dial")
	deadlineTimeout = flag.Int("dlt", 1, "The timeout(Second) of deadline")
)

func init() {
	go addCDN()
}

func addCDN() {
	fastestCDN := make(map[string]string, *size)

	ips := gscan.ScanIP("./iprange.conf", gscanConf)

	t := 0
	for _, v := range ips {
		if t >= *size {
			break
		}
		t++
		fastestCDN[v] = v + ":443"
		fmt.Println(v)
	}

	go func() {
		for {
			for _, v := range fastestCDN {
				cdnChan <- v
			}
		}
	}()
}

func main() {
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.ConnectDial = func(network, addr string) (c net.Conn, err error) {
		if addr == "kyfw.12306.cn:443" {
			c, err = net.DialTimeout(network, <-cdnChan, time.Duration(*dialTimeout)*time.Second)
			if c, ok := c.(*net.TCPConn); err == nil && ok {
				c.SetKeepAlive(false)
				c.SetDeadline(time.Now().Add(time.Duration(*deadlineTimeout) * time.Second))
			}
			return
		}
		c, err = net.Dial(network, addr)
		if c, ok := c.(*net.TCPConn); err == nil && ok {
			c.SetKeepAlive(true)
		}
		return
	}
	proxy.Verbose = *verbose
	log.Println("start listen on " + *addr)
	log.Fatal(http.ListenAndServe(*addr, proxy))
}

var gscanConf = `
{
  "ScanWorker" : 40,
  "VerifyPing":  false,          
  "ScanMinPingRTT" : 5,
  "ScanMaxPingRTT" : 1000,
  "ScanMinSSLRTT":   0,
  "ScanMaxSSLRTT":   3000,
  "ScanCountPerIP" : 2,
 
  "Operation" : "ScanGoogleIP", 
  
  "ScanGoogleIP" :{
       "SSLCertVerifyHosts" : ["kyfw.12306.cn"],  
       "HTTPVerifyHosts" : ["kyfw.12306.cn"],
       "RecordLimit" :     500,
       "OutputSeparator":  "\",\r\n\"",
       "OutputFile" :      "./12306_ip.txt"
  },
  
  "ScanGoogleHosts":{
       "InputHosts":"./hosts.input",
       "OutputHosts": "./hosts.output",
       "HTTPVerifyHosts" : ["www.google.com"]
  } 
}
`
