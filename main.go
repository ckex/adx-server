package main
import (
	"fmt"
	"adx-server/com.iclick.adx"
	"adx-server/com.iclick.adx/message"
	"time"
	"strconv"
	"runtime"
	"flag"
)


var curr = flag.Int("curr", 2, "并发数")
var rnum = flag.Int("rnum", 2, "请求数")
var tint = flag.Int("tint", 500, "每次请求间隔")
var u = flag.String("url", "http://127.0.0.1:9090/v1/bid", "请求目录地址")


func main() {

	flag.Parse()

	fmt.Println("并发数=", *curr, " 单线程请求数=", *rnum, " 每次请求间隔=", *tint, " 请求地址=", *u)

	//	com_iclick_adx.RequestDSP()
	//	str := com_iclick_adx.SayHello("Ckex.zha")
	//	fmt.Println(str)

	timeStart := time.Now().UnixNano()

	runtime.GOMAXPROCS(runtime.NumCPU())

	count := *curr
	subcount := *rnum

	all := count*subcount

	resultlist := make([]string, all)
	quit := make(chan string)
	ch := make(chan string, all)

	for i := 0; i< count; i++ {
		go func(index int) {
			for j := 0; j<subcount; j++ {
				st := time.Now().UnixNano()
				res, err := request("\t ["+strconv.Itoa(index)+" -- "+strconv.Itoa(j)+"]", *u)
				if err != nil {
					res = "-1"
				}
				ch <- res
				uset := (time.Now().UnixNano()-st) / (1000*1000)
				sleepTime := int64(*tint) - uset
				if sleepTime > 0 {
					time.Sleep(time.Millisecond*time.Duration(sleepTime))
				}
			}
		}(i)
	}

	go func() {
		st := time.Now().UnixNano()
		time.Sleep(time.Second*60)
		us := (time.Now().UnixNano() - st) / (1000*1000)
		quit <- "quit,useTime="+strconv.FormatInt(us, 10)+" -ms"
	}()

	next := true
	var rc int = 0
	for next {
		select {
		case result := <-ch:
			resultlist[rc] = result
			rc+=1
			if rc == all {
				next = false
			}
		case quit := <-quit:
			fmt.Println("quit==>", quit)
			next = false
		}
	}
	allTime := (time.Now().UnixNano() - timeStart) / (1000*1000)
	var success = 0
	for index, ele := range resultlist {
		if len(ele) > 5 {
			success +=1
		}
		fmt.Println(index, ele)
	}
	fmt.Println("Game Over . All Use Time = ", allTime, " All Request = ", all, " Avg = ", allTime / int64(all), " 成功率 = ", (float64(success)/float64(all))*100.0, "%", " 当前QPS:", int64(float64(all) / (float64(allTime)/1000.0)))


}

func request(index, u string) (str string, err error) {
	st := time.Now().UnixNano()
	bidRequest := builderBidRequest()
	response, err := com_iclick_adx.RequestDSP(bidRequest, u)
	if err != nil {
		fmt.Println(err)
		return
	}
	useTime := (time.Now().UnixNano()-st) / (1000*1000)
	str = response.Id+"\t index:"+index+" use time:"+strconv.FormatInt(useTime, 10)+"-ms"
	//	fmt.Println(str, "response==>", response)
	return
}

func builderBidRequest() (bidRequest *message.BidRequest) {
	banner := &message.BidRequest_Impression_Banner{
		W:100,
		H:250,
	}

	imp := []*message.BidRequest_Impression{
		&message.BidRequest_Impression{
			Id:"imp-123",
			Banner:banner,
			Bidfloor:0.01,
			Bidfloorcur:"CNY",
		},
	}

	publisher := &message.BidRequest_Site_Publisher{
		Id:"publisher-123",
	}

	site := &message.BidRequest_Site{
		Id:"site-123",
		Page:"http://wwww.i-click.com",
		Publisher:publisher,
	}

	device := &message.BidRequest_Device{
		Ua:"Go-http-client/1.1",
		Ip:"127.0.0.1",
	}

	user := &message.BidRequest_User{
		Id:"user-123",
		Buyeruid:"buyreuid-321",
	}

	cnys := []string{
		"CNY",
	}
	bidRequest = &message.BidRequest{
		Id:"456",
		Imp:imp,
		Site:site,
		Device:device,
		User:user,
		Test:1,
		Tmax:100,
		Cur:cnys,
	}
	return
}
