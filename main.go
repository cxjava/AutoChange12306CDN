package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cxjava/gscan"
)

var (
	logger   *log.Logger
	cdnChan  = make(chan string, 10)
	fastest  = ""
	addr     = flag.String("a", ":8888", "The proxy listen address")
	queryURL = flag.String("q", "/otn/leftTicket/query", "The prefix of 12306 query URL")
	size     = flag.Int("s", 20, "The fastest number of IP")
)

func init() {
	go addCDN()
	logger = log.New(os.Stdout, "[AutoChange12306CDN]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func addCDN() {
	fastestCDN := make(map[string]string, *size)

	ips := gscan.ScanIP("./iprange.conf", gscanConf)
	fastest = ips[0] + ":443"
	t := 0
	for _, v := range ips {
		if t >= *size {
			break
		}
		t++
		fastestCDN[v] = v + ":443"
		fmt.Println(v)
	}
	logger.Println("CDN query is done! fastest one is:" + fastest)
	logger.Println("Server is listening at ", *addr)

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

	ch := make(chan bool)
	Gomitmproxy(*addr, ch)
	<-ch
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
