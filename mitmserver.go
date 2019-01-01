package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	errorResp = "HTTP/1.0 200 OK\r\n" +
		"Connection: close\r\n" +
		"Content-Type: application/json;charset=UTF-8\r\n" +
		"\r\n" +
		`{"data":{"flag":"1","map":{"CUW":"重庆北","HKN":"汉口","WCN":"武昌"},"result":["|预订|2400000Z490F|Z49|BXP|CDW|HKN|CUW|22:20|05:32|07:12|N|g%2Bd3SnPhBrp6gzy5Eq4zATL6rGgQn4nznpuJl5niuOcsnDutccVaBIqfJp4%3D|20190101|3|P4|05|08|0|0||||无|||无||无|无|||||10401030|1413|0|0"]},"httpstatus":200,"messages":"","status":true}`
)

const (
	Version   = "1.1"
	ONE_DAY   = 24 * time.Hour
	TWO_WEEKS = ONE_DAY * 14
	ONE_MONTH = 1
	ONE_YEAR  = 1
)

// HandlerWrapper wrapper of handler for http server
type HandlerWrapper struct {
	wrapped        http.Handler
	tlsConfig      *TLSConfig
	pk             *PrivateKey
	pkPem          []byte
	issuingCert    *Certificate
	issuingCertPem []byte
	dynamicCerts   *Cache
	certMutex      sync.Mutex
	https          bool
}

// InitConfig init HandlerWrapper
func InitConfig(tlsconfig *TLSConfig) *HandlerWrapper {
	handler := &HandlerWrapper{
		tlsConfig:    tlsconfig,
		dynamicCerts: NewCache(),
	}
	err := handler.GenerateCertForClient()
	if err != nil {
		return nil
	}
	return handler
}

// ServeHTTP the main function interface for http handler
func (handler *HandlerWrapper) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method == "CONNECT" {
		handler.https = true
		handler.InterceptHTTPS(resp, req)
	} else {
		handler.https = false
		handler.DumpHTTPAndHTTPS(resp, req)
	}
}

// DumpHTTPAndHTTPS function to dump the HTTP request header and body
func (handler *HandlerWrapper) DumpHTTPAndHTTPS(resp http.ResponseWriter, req *http.Request) {
	req.Header.Del("Proxy-Connection")
	req.Header.Set("Connection", "Keep-Alive")

	var reqDump []byte
	ch := make(chan bool)
	// handle connection
	go func() {
		reqDump, _ = httputil.DumpRequestOut(req, true)
		ch <- true
	}()

	connHj, _, err := resp.(http.Hijacker).Hijack()
	if err != nil {
		logger.Println("Hijack fail to take over the TCP connection from client's request")
	}
	defer connHj.Close()

	host := req.Host

	matched, _ := regexp.MatchString(":[0-9]+$", host)

	var connOut net.Conn
	if !handler.https {
		if !matched {
			host += ":80"
		}
		connOut, err = net.DialTimeout("tcp", host, time.Second*30)
		if err != nil {
			logger.Println("Dial to", host, "error:", err)
			return
		}
	} else {
		if !matched {
			host += ":443"
		}
		if host == "kyfw.12306.cn:443" {
			host = fastest
			if strings.HasPrefix(req.RequestURI, *queryURL) {
				host = <-cdnChan
				logger.Println("current CDN is:", host)
				connOut, err = tls.DialWithDialer(&net.Dialer{
					Timeout:   1 * time.Second,
					KeepAlive: 0,
				}, "tcp", host, handler.tlsConfig.ServerTLSConfig)
				if err != nil {
					logger.Println("Dial to", host, "error:", err)
					connHj.Write([]byte(errorResp))
					return
				}
			} else {
				connOut, err = tls.Dial("tcp", host, handler.tlsConfig.ServerTLSConfig)
				if err != nil {
					logger.Println("Dial to", host, "error:", err)
					return
				}
			}
		} else {
			connOut, err = tls.Dial("tcp", host, handler.tlsConfig.ServerTLSConfig)
			if err != nil {
				logger.Println("Dial to", host, "error:", err)
				return
			}
		}

	}

	// Write writes an HTTP/1.1 request, which is the header and body, in wire format. This method consults the following fields of the request:
	/*
		Host
		URL
		Method (defaults to "GET")
		Header
		ContentLength
		TransferEncoding
		Body
	*/
	if err = req.Write(connOut); err != nil {
		logger.Println("send to server error", err)
		connHj.Write([]byte(errorResp))
		return
	}

	respFromRemote, err := http.ReadResponse(bufio.NewReader(connOut), req)
	if err != nil && err != io.EOF {
		logger.Println("Fail to read response from remote server.", err)
		connHj.Write([]byte(errorResp))
	}

	respDump, err := httputil.DumpResponse(respFromRemote, true)
	if err != nil {
		logger.Println("Fail to dump the response.", err)
		connHj.Write([]byte(errorResp))
	}
	// Send remote response back to client
	_, err = connHj.Write(respDump)
	if err != nil {
		logger.Println("Fail to send response back to client.", err)
		connHj.Write([]byte(errorResp))
	}

	<-ch
	// why write to reqDump, and in httpDump resemble to req again
	// in test, i find that the req may be destroyed by sth i currently dont know
	// so while parsing req in httpDump directly, it will raise execption
	// so dump its content to reqDump first.
	//go httpDump(reqDump, respFromRemote)
}

// InterceptHTTPS to dump data in HTTPS
func (handler *HandlerWrapper) InterceptHTTPS(resp http.ResponseWriter, req *http.Request) {
	addr := req.Host
	host := strings.Split(addr, ":")[0]

	cert, err := handler.FakeCertForName(host)
	if err != nil {
		logger.Println("Could not get mitm cert for name: %s\nerror: %s", host, err)
		respBadGateway(resp)
		return
	}

	connIn, _, err := resp.(http.Hijacker).Hijack()
	if err != nil {
		logger.Println("Unable to access underlying connection from client: %s", err)
		respBadGateway(resp)
		return
	}

	tlsConfig := copyTlsConfig(handler.tlsConfig.ServerTLSConfig)
	tlsConfig.Certificates = []tls.Certificate{*cert}
	tlsConnIn := tls.Server(connIn, tlsConfig)
	listener := &mitmListener{tlsConnIn}
	httpshandler := http.HandlerFunc(func(resp2 http.ResponseWriter, req2 *http.Request) {
		req2.URL.Scheme = "https"
		req2.URL.Host = req2.Host
		handler.DumpHTTPAndHTTPS(resp2, req2)
	})

	go func() {
		server := &http.Server{Handler: httpshandler}
		server.SetKeepAlivesEnabled(false)
		err = server.Serve(listener)
		if err != nil && err != io.EOF {
			logger.Printf("Error serving mitm'ed connection: %s", err)
		}
	}()

	connIn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
}

func (hw *HandlerWrapper) GenerateCertForClient() (err error) {
	if hw.tlsConfig.Organization == "" {
		hw.tlsConfig.Organization = "gomitmproxy" + Version
	}
	if hw.tlsConfig.CommonName == "" {
		hw.tlsConfig.CommonName = "gomitmproxy"
	}
	if hw.pk, err = LoadPKFromFile(hw.tlsConfig.PrivateKeyFile); err != nil {
		hw.pk, err = GeneratePK(2048)
		if err != nil {
			return fmt.Errorf("Unable to generate private key: %s", err)
		}
		hw.pk.WriteToFile(hw.tlsConfig.PrivateKeyFile)
	}
	hw.pkPem = hw.pk.PEMEncoded()
	hw.issuingCert, err = LoadCertificateFromFile(hw.tlsConfig.CertFile)
	if err != nil || hw.issuingCert.ExpiresBefore(time.Now().AddDate(0, ONE_MONTH, 0)) {
		hw.issuingCert, err = hw.pk.TLSCertificateFor(
			hw.tlsConfig.Organization,
			hw.tlsConfig.CommonName,
			time.Now().AddDate(ONE_YEAR, 0, 0),
			true,
			nil)
		if err != nil {
			return fmt.Errorf("Unable to generate self-signed issuing certificate: %s", err)
		}
		hw.issuingCert.WriteToFile(hw.tlsConfig.CertFile)
	}
	hw.issuingCertPem = hw.issuingCert.PEMEncoded()
	return
}

func (hw *HandlerWrapper) FakeCertForName(name string) (cert *tls.Certificate, err error) {
	kpCandidateIf, found := hw.dynamicCerts.Get(name)
	if found {
		return kpCandidateIf.(*tls.Certificate), nil
	}

	hw.certMutex.Lock()
	defer hw.certMutex.Unlock()
	kpCandidateIf, found = hw.dynamicCerts.Get(name)
	if found {
		return kpCandidateIf.(*tls.Certificate), nil
	}

	//create certificate
	certTTL := TWO_WEEKS
	generatedCert, err := hw.pk.TLSCertificateFor(
		hw.tlsConfig.Organization,
		name,
		time.Now().Add(certTTL),
		false,
		hw.issuingCert)
	if err != nil {
		return nil, fmt.Errorf("Unable to issue certificate: %s", err)
	}
	keyPair, err := tls.X509KeyPair(generatedCert.PEMEncoded(), hw.pkPem)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse keypair for tls: %s", err)
	}

	cacheTTL := certTTL - ONE_DAY
	hw.dynamicCerts.Set(name, &keyPair, cacheTTL)
	return &keyPair, nil
}

func copyTlsConfig(template *tls.Config) *tls.Config {
	tlsConfig := &tls.Config{}
	if template != nil {
		*tlsConfig = *template
	}
	return tlsConfig
}

func copyHTTPRequest(template *http.Request) *http.Request {
	req := &http.Request{}
	if template != nil {
		*req = *template
	}
	return req
}

func respBadGateway(resp http.ResponseWriter) {
	resp.WriteHeader(502)
}
