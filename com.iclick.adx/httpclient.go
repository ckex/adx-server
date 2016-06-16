package com_iclick_adx

import (
	"net/http"
	"adx-server/com.iclick.adx/message"
	//	"io/ioutil"
	//	"strings"
	//	"net/url"
	//	"fmt"
	"github.com/golang/protobuf/proto"
	"bytes"
	//	"fmt"
	"io/ioutil"
	//	"golang.org/x/net/http2"
	//	"crypto/tls"
	"time"
	"net"
	"strings"
	"errors"

	logger "github.com/alecthomas/log4go"
)

func SayHello(name string) (s string) {
	s = "Hello " + name
	return
}

func SimplePost(request *message.BidRequest, u string) (response message.BidResponse) {
	data, _ := proto.Marshal(request)
	http.Post(u, "application/x-protobuf", bytes.NewReader(data))
	response = message.BidResponse{}
	return
}

var tr = &http.Transport{
	ResponseHeaderTimeout:time.Millisecond * time.Duration(150),
	MaxIdleConnsPerHost:100,
	DisableKeepAlives:true,
	Dial:(&net.Dialer{
		Timeout:60 * time.Second,
		KeepAlive:2 * time.Minute,
	}).Dial,
}

var client = &http.Client{
	Transport:tr,
}

func RequestDSP(request *message.BidRequest, u string) (response *message.BidResponse, err error) {

	data, _ := proto.Marshal(request)

	//	client := &http.Client{
	//		Transport:tr,
	//	}

	// http2
	//	client := &http.Client{
	//		Transport:http2.Transport{
	//
	//		},
	//	}

	req, _ := http.NewRequest("POST", u, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/x-protobuf") //这个一定要加，不加form的值post不过去，被坑了两小时
	req.Header.Set("x-openrtb-version", "2.3");

	//	fmt.Printf("%+v\n", req)                                                         //看下发送的结构

	resp, err := client.Do(req) //发送
	if err != nil {
		neterr, ok := err.(net.Error)
		if !ok {
			logger.Error("Error ==> %v,%v", err, neterr)
			err = errors.New(err.Error())
			return
		} else if !neterr.Timeout() {
			logger.Error("net.Error.Timeout = false; want true")
			err = errors.New(neterr.Error())
			return
		}
		if got := neterr.Error(); !strings.Contains(got, "Client.Timeout exceeded") {
			logger.Error("error string = %q; missing timeout substring", got)
			err = errors.New(neterr.Error())
			return
		}
		logger.Error("client do TimeOut.  %v", err)
		err = errors.New(neterr.Error())
		return
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()     //一定要关闭resp.Body
		} else {
			logger.Warn(" resp is null .")
		}
	}()
	respdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("response error %+v ", err)
	}
	response = &message.BidResponse{}
	proto.Unmarshal(respdata, response)
	return
}