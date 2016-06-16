package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"adx-server/com.iclick.adx"
	"adx-server/com.iclick.adx/message"
	logger "github.com/alecthomas/log4go"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))
var curr = flag.Int("curr", 1, "并发数")
var rnum = flag.Int("rnum", 1, "请求数")
var tint = flag.Int("tint", 500, "每次请求间隔")
var u = flag.String("url", "http://127.0.0.1:9090/v1/bid", "请求目录地址")

func init() {
	logger.LoadConfiguration("./conf/log4go.xml")
}

func main() {
	defer logger.Close()
	flag.Parse()
	cpus := runtime.NumCPU()

	logger.Info("CPU核心数=%d,\t并发数=%d,\t单线程请求数=%d,\t每次请求间隔时间=%d,\t请求地址=%s", cpus, *curr, *rnum, *tint, *u)

	timeStart := time.Now().UnixNano()
	runtime.GOMAXPROCS(cpus)
	count := *curr
	subcount := *rnum
	all := count * subcount

	resultlist := make([]string, all)
	ch := make(chan string, all)

	for i := 0; i < count; i++ {
		go func(index int) {
			for j := 0; j < subcount; j++ {
				st := time.Now().UnixNano()
				res, err := request("\t [" + strconv.Itoa(index) + " -- " + strconv.Itoa(j) + "]", *u)
				if err != nil {
					res = "-1"
				}
				ch <- res
				uset := (time.Now().UnixNano() - st)
				sleepTime := int64(*tint) * 1000 * 1000 - uset
				if sleepTime > 0 {
					time.Sleep(time.Nanosecond * time.Duration(sleepTime))
				}
			}
		}(i)
	}

	quit := make(chan string)
	go func() {
		st := time.Now().UnixNano()
		time.Sleep(time.Second * 60)
		us := (time.Now().UnixNano() - st) / (1000 * 1000)
		quit <- "quit,useTime=" + strconv.FormatInt(us, 10) + " -ms"
	}()

	next := true
	var rc int = 0
	for next {
		select {
		case result := <-ch:
			resultlist[rc] = result
			rc += 1
			if rc == all {
				next = false
			}
		case quit := <-quit:
			fmt.Println("quit==>", quit)
			next = false
		}
	}
	allTime := (time.Now().UnixNano() - timeStart) / (1000 * 1000)
	var success = 0
	for index, ele := range resultlist {
		if len(ele) > 5 {
			success += 1
		}
		logger.Info(index, ele)
	}

	allTimeSecond := float64(allTime) / 1000.00
	if allTimeSecond <= 0 {
		allTimeSecond = 1
	}
	qps := float64(all) / allTimeSecond
	if qps < 1 && all > 0 {
		qps = 1
	}
	logger.Info("Game over. All use time=%d, All reqeust=%d, Avg=%d, 成功率=%.3f%s, 预计当前QPS=%d", allTime, all, allTime / int64(all), ( (float64(success) / float64(all)) * 100.0), "%", int(qps))

}

func request(index, u string) (str string, err error) {
	st := time.Now().UnixNano()
	bidRequest := builderBidRequest()
	response, err := com_iclick_adx.RequestDSP(bidRequest, u)
	if err != nil {
		logger.Error("Http error. %v", err)
		return
	}
	useTime := (time.Now().UnixNano() - st) / (1000 * 1000)
	str = response.Id + "\t index:" + index + " use time:" + strconv.FormatInt(useTime, 10) + "-ms"
	return
}

func builderBidRequest() (bidRequest *message.BidRequest) {

	publisher := &message.BidRequest_Publisher{
		Id: strconv.Itoa(1000 + r.Intn(100)),
	}

	banner := &message.BidRequest_Impression_Banner{
		W: 300,
		H: 250,
	}

	video := &message.BidRequest_Impression_Video{
		Mimes:       []string{"mimes"},
		Minduration: int32(10 + r.Intn(5)),
		Maxduration: int32(20 + r.Intn(5)),
		Protocols:   []int32{int32(1 + r.Intn(6))},
		W:           300,
		H:           250,
		Startdelay:  0,
	}

	imp := []*message.BidRequest_Impression{
		&message.BidRequest_Impression{// Banner
			Id:          strconv.Itoa(2000 + r.Intn(100)),
			Banner:      banner,
			Bidfloor:    1.25 + r.Float32(),
			Bidfloorcur: "CNY",
		},
		&message.BidRequest_Impression{// Video
			Id:          strconv.Itoa(2000 + r.Intn(100)),
			Video:       video,
			Bidfloor:    1.25 + r.Float32(),
			Bidfloorcur: "CNY",
		},
	}

	site := &message.BidRequest_Site{
		Id:        strconv.Itoa(3000 + r.Intn(100)),
		Page:      "http://wwww.i-click.com",
		Publisher: publisher,
	}

	app := &message.BidRequest_App{
		Id:        strconv.Itoa(4000 + r.Intn(100)),
		Publisher: publisher,
	}

	device := &message.BidRequest_Device{
		Ua:  "Go-http-client/1.1",
		Ip:  "127.0.0.1",
		Ifa: "ifa",
	}

	user := &message.BidRequest_User{
		Id:       strconv.Itoa(5000 + r.Intn(100)),
		Buyeruid: strconv.Itoa(6000 + r.Intn(100)),
	}

	bidRequests := []*message.BidRequest{
		&message.BidRequest{// Site
			Id:     strconv.Itoa(7000 + r.Intn(100)),
			Imp:    []*message.BidRequest_Impression{imp[r.Intn(len(imp))]},
			Site:   site,
			Device: device,
			User:   user,
			Test:   1,
			Tmax:   100,
			Cur:    []string{"CNY"},
		},
		&message.BidRequest{// App
			Id:     strconv.Itoa(7000 + r.Intn(100)),
			Imp:    []*message.BidRequest_Impression{imp[r.Intn(len(imp))]},
			App:    app,
			Device: device,
			User:   user,
			Test:   1,
			Tmax:   100,
			Cur:    []string{"CNY"},
		},
	}
	bidRequest = bidRequests[r.Intn(len(bidRequests))]
	return
}
