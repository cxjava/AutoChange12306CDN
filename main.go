package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/elazarl/goproxy"
	"github.com/koding/multiconfig"
)

type myResp struct {
	Resp *http.Response
	IP   string
}
type Config struct {
	ListenAddr   string `default:":8080"`
	URLMatches   []string
	FromStation  []string
	ToStation    []string
	ProxyVerbose bool `default:"false"`
	VerifyBody   bool `default:"false"`
	CDN          []string
	ConnTimeout  int `default:"1500"`
	ReadTimeout  int `default:"1000"`
}

var (
	cdns           = make(map[string]string)
	timeout        = 5 * time.Second       //超时时间
	fromStationStr = ""                    //起点站
	fromStation    []string                //起点站数组
	toStationStr   = ""                    //终点站
	toStation      []string                //终点站数组
	cdnChan        = make(chan string, 20) //CDN暂存的地方
	index          = make(chan int)        //获取CDN下标的队列
	add            = make(chan int)        //通知CDN下标加1的队列
	config         = new(Config)           //配置文件
	r              = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func main() {
	defer log.Flush()
	defer func() {
		if err := recover(); err != nil {
			log.Critical(err)
			log.Critical(string(debug.Stack()))
		}
	}()
	readConfig()
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	proxy := goproxy.NewProxyHttpServer()

	proxy.Verbose = config.ProxyVerbose
	proxy.OnRequest(goproxy.ReqHostIs("kyfw.12306.cn:443")).HandleConnect(goproxy.AlwaysMitm)

	addUrlMatches(proxy)

	AddCDN()

	log.Info("监听端口:", config.ListenAddr)
	http.ListenAndServe(config.ListenAddr, proxy)
}

func addUrlMatches(proxy *goproxy.ProxyHttpServer) {
	for _, matchUrl := range config.URLMatches {
		proxy.OnRequest(goproxy.UrlMatches(regexp.MustCompile(matchUrl))).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			req.Header.Set("If-Modified-Since", time.Now().Local().Format(time.RFC1123Z))
			req.Header.Set("If-None-Match", strconv.FormatInt(time.Now().UnixNano(), 10))
			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Del("Content-Length")

			from := req.FormValue("leftTicketDTO.from_station")
			to := req.FormValue("leftTicketDTO.to_station")

			log.Debug("from:", from, " to:", to)
			log.Debug(req.URL.RawQuery)
			//随机读取提供的查询地址，不在是单一的武汉，可以随机：武汉，武昌，汉口。免得缓存
			if len(fromStation) > 0 && strings.Contains(fromStationStr, from) {
				req.URL.RawQuery = strings.Replace(req.URL.RawQuery, from, fromStation[r.Intn(len(fromStation))], -1)
			}

			if len(toStation) > 0 && strings.Contains(toStationStr, to) {
				req.URL.RawQuery = strings.Replace(req.URL.RawQuery, to, toStation[r.Intn(len(toStation))], -1)
			}

			log.Info("request URL:", req.URL.RawQuery)

			resp := GetResponse(req).Resp

			return req, resp
		})
	}

	proxy.OnRequest(goproxy.UrlMatches(regexp.MustCompile(".*/otn/leftTicket/log.*"))).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		return req, goproxy.NewResponse(req,
			goproxy.ContentTypeText, http.StatusForbidden,
			"Don't log !")
	})
}

//	clientConn, _ := newForwardClientConn("www.google.com","https")
//	resp, err := clientConn.Do(req)
func NewForwardClientConn(forwardAddress, scheme string) (*httputil.ClientConn, error) {
	if "http" == scheme {
		conn, err := net.Dial("tcp", forwardAddress+":80")
		if err != nil {
			log.Error(forwardAddress, "newForwardClientConn net.Dial error:", err)
			return nil, err
		}
		return httputil.NewProxyClientConn(conn, nil), nil
	}
	ipConn, err := net.DialTimeout("tcp", forwardAddress+":443", time.Duration(config.ConnTimeout)*time.Millisecond)
	if err != nil {
		log.Error("NewForwardClientConn DialTimeout error:", err)
		return nil, err
	}
	go func(ipConn net.Conn) {
		time.Sleep(5 * time.Second)
		log.Debug("ipConn closed!")
		ipConn.Close()
	}(ipConn)
	conn := tls.Client(ipConn, &tls.Config{
		InsecureSkipVerify: true,
	})
	conn.SetReadDeadline(time.Now().Add(time.Duration(config.ReadTimeout) * time.Millisecond))
	go func(conn *tls.Conn) {
		time.Sleep(5 * time.Second)
		log.Debug("conn closed!")
		conn.Close()
	}(conn)

	return httputil.NewProxyClientConn(conn, nil), nil
}

func DoForWardRequest(forwardAddress string, req *http.Request) (*http.Response, error) {
	clientConn, err := NewForwardClientConn(forwardAddress, "https")
	if err != nil {
		log.Error("DoForWardRequest newForwardClientConn error:", err)
		return nil, err
	}
	return clientConn.Do(req)
}

func DoForWardRequest2(forwardAddress string, req *http.Request) (*http.Response, error) {
	if !strings.Contains(forwardAddress, ":") {
		forwardAddress = forwardAddress + ":443"
	}
	ipConn, err := net.DialTimeout("tcp", forwardAddress, time.Duration(config.ConnTimeout)*time.Millisecond)
	if err != nil {
		log.Error("doForWardRequest2 DialTimeout error:", err)
		return nil, err
	}
	go func(ipConn net.Conn) {
		time.Sleep(5 * time.Second)
		log.Debug("ipConn closed!")
		ipConn.Close()
	}(ipConn)

	conn := tls.Client(ipConn, &tls.Config{
		InsecureSkipVerify: true,
	})
	conn.SetReadDeadline(time.Now().Add(time.Duration(config.ReadTimeout) * time.Millisecond))
	go func(conn *tls.Conn) {
		time.Sleep(5 * time.Second)
		log.Debug("conn closed!")
		conn.Close()
	}(conn)
	errWrite := req.Write(conn)
	// errWrite := req.WriteProxy(conn)
	if errWrite != nil {
		log.Error("doForWardRequest Write error:", errWrite)
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(conn), req)
}

//读取配置文件
func readConfig() {
	m := multiconfig.NewWithPath("config.toml") // supports TOML and JSON
	// Populated the serverConf struct
	m.MustLoad(config) // Check for error

	if len(config.CDN) < 1 {
		log.Error("淘气了,CDN也不配置～～")
		return
	}
	//解析地址
	parseStationNames()
	//解析起点站
	if len(config.FromStation) > 0 {
		fromStation = make([]string, len(config.FromStation))
		for k, v := range config.FromStation {
			fromStationStr = fromStationStr + stationMap[v]
			fromStation[k] = stationMap[v]
		}
	}
	//终点站
	if len(config.ToStation) > 0 {
		toStation = make([]string, len(config.ToStation))
		for k, v := range config.ToStation {
			toStationStr = toStationStr + stationMap[v]
			toStation[k] = stationMap[v]
		}
	}

	log.Info("起点站:", fromStationStr, ":", fromStation)
	log.Info("终点站:", toStationStr, ":", toStation)
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, nil, err
	}
	if err = b.Close(); err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func GetResponse(req *http.Request) *myResp {
	total := 2
	ch := make(chan *myResp, total)
	defer func() {
		go func() {
			for re := range ch {
				if re.Resp != nil {
					log.Info("resp is closed!")
					re.Resp.Body.Close()
				} else {
					delete(cdns, re.IP)
					log.Info("resp is null!")
				}
			}
		}()
	}()
	for i := 0; i < total; i++ {
		go func(r *http.Request) {
			c := <-cdnChan
			log.Info(c, " requested!")
			resp, err := DoForWardRequest(c, r)
			if err != nil && err != httputil.ErrPersistEOF {
				log.Error(c, " OnRequest error : ", err)
			}
			log.Info(c, " responsed!")
			ch <- &myResp{
				Resp: resp,
				IP:   c,
			}
		}(req)
	}

	return <-ch
}

func AddCDN() {
	cdns = make(map[string]string, 100)
	for i := 0; i < len(config.CDN); i++ {
		cdns[config.CDN[i]] = config.CDN[i]
	}

	go func() {
		for {
			for _, v := range cdns {
				cdnChan <- v
			}
		}
	}()

	go func() {
		for {
			if len(cdns) < 20 {
				log.Info("cdn 不够了啊")
				for i := 0; i < len(config.CDN); i++ {
					cdns[config.CDN[i]] = config.CDN[i]
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

//读取响应
func ParseResponseBody(resp *http.Response) string {
	var body []byte
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Error(err)
			return ""
		}
		defer reader.Close()
		bodyByte, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Error(err)
			return ""
		}
		body = bodyByte
	default:
		bodyByte, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return ""
		}
		body = bodyByte
	}
	return string(body)
}
