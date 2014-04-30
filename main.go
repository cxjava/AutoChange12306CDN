package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/elazarl/goproxy"
)

type config struct {
	ListenAddr string   `toml:"listen_addr"`
	UrlMatches []string `toml:"url_matches"`
	Timeout    int      `toml:"time_out"`
	LogLevel   int      `toml:"log_level"`
	Verbose    bool     `toml:"proxy_verbose"`
	CDN        []string `toml:"cdn"`
}

var (
	timeout = 5 * time.Second //超时时间
	index   = make(chan int)  //获取CDN下标的队列
	add     = make(chan int)  //通知CDN下标加1的队列
	Config  config            //配置文件
)

func main() {

	if err := ReadConfig(); err != nil {
		log.Println(err)
		os.Exit(0)
		return
	}
	if len(Config.CDN) < 1 {
		log.Println("淘气了,CDN也不配置～～")
		os.Exit(0)
		return
	}

	proxy := goproxy.NewProxyHttpServer()

	proxy.Verbose = Config.Verbose

	proxy.OnRequest(goproxy.ReqHostIs("kyfw.12306.cn:443")).HandleConnect(goproxy.AlwaysMitm)

	for _, matchUrl := range Config.UrlMatches {
		proxy.OnRequest(goproxy.UrlMatches(regexp.MustCompile(matchUrl))).DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			i := <-index
			add <- 0
			Info(r.URL.Host, ",", r.URL.Path)
			Info("使用第", i, "个", Config.CDN[i])

			for k, v := range r.Header {
				Info(k, "=", v)
			}

			resp, err := DoForWardRequest(Config.CDN[i], r)
			if err != nil {
				Error(Config.CDN[i], "OnRequest error:", err)
				return r, nil
			}
			Info("ddddddddd")
			for k, v := range resp.Header {
				Info(k, "=", v)
			}
			return r, resp
		})
	}

	go ChangeCDN()
	Info("监听端口:", Config.ListenAddr)
	log.Fatalln(http.ListenAndServe(Config.ListenAddr, proxy))
}

//切换CDN下标
func ChangeCDN() {
	i := 0
	index <- i
	for {
		select {
		case <-add:
			i += 1
			i = i % len(Config.CDN)
			index <- i
		}
	}
}

func DoForWardRequest(forwardAddress string, req *http.Request) (*http.Response, error) {
	if !strings.Contains(forwardAddress, ":") {
		forwardAddress = forwardAddress + ":443"
	}

	conn, err := tls.Dial("tcp", forwardAddress, &tls.Config{
		InsecureSkipVerify: true,
	})
	// conn, err := net.Dial("tcp", forwardAddress)

	if err != nil {
		fmt.Println("doForWardRequest DialTimeout error:", err)
		return nil, err
	}
	// defer conn.Close()

	//buf_forward_conn *bufio.Reader
	buf_forward_conn := bufio.NewReader(conn)
	// req.Close = true
	req.Header.Set("Connection", "close")
	errWrite := req.Write(conn)
	if errWrite != nil {
		fmt.Println("doForWardRequest Write error:", errWrite)
		return nil, err
	}

	return http.ReadResponse(buf_forward_conn, req)
}

//设置log相关
func SetLogInfo() {
	// debug 1, info 2
	if Config.LogLevel > 0 {
		SetLevel(Config.LogLevel)
	} else {
		SetLevel(2)
	}
	SetLogger("console", "")
	SetLogger("file", `{"filename":"log.log"}`)
}

//读取配置文件
func ReadConfig() error {
	if _, err := toml.DecodeFile("config.ini", &Config); err != nil {
		Error(err)
		return err
	}

	SetLogInfo()
	if Config.Timeout > 0 {
		timeout = time.Duration(Config.Timeout) * time.Second
	}

	return nil
}
