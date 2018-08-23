package main

import (
	"flag"
	"time"
	"os"
	"fmt"
	"bufio"
	"io"
	"strings"
	"github.com/mgutz/str"
	"net/url"
	"crypto/md5"
	"encoding/hex"
)


/**

基于Golang协程实现流量统计分析系统

 */

const HANDLE_DIG  = "/dig?"

type cmdParams struct {
	logFilePath string
	routineNum int
}
//统计上传的数据
type digData struct {
	time string
	url string
	refer string
	ua string
}


type urlData struct {
	data digData
	uid string
	unode urlNode
}

type urlNode struct {
	unType string
	unRid int
	unUrl string
	unTime string
} 

type storageBlock struct {
	counterType string
	storageModel string
	unode  urlNode
}

//var log  =

func init()  {
	
}

/**

基于Golang协程实现流量统计分析系统

逐行消费日志 --> channel --> 日志解析 --> channel --> UV统计、PV统计 --> channel -->  数据存储

 */
func main()  {

	//1. 获取参数
	logFilePath := flag.String("logFilePath","/xxx/xx.log", "log file path")
	routineNum := flag.Int("routineNum", 5, "consumer number by goroutine")
	l := flag.String("l", "/temp/log", "this programe runtime log target file path")
	flag.Parse()

	params := cmdParams{logFilePath: *logFilePath,routineNum: *routineNum}
	//2. 打日志
	logFd, err := os.OpenFile(*l, os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil{
		fmt.Println(logFd)
		defer logFd.Close()
	}

	fmt.Println("Exec start.")
	fmt.Println("Params: logFilePath")

	//3. 初始化一些channel ，用于数据传递
	var logChannel = make(chan string, 3 * params.routineNum)
	var pvChannel = make(chan urlData, params.routineNum)
	var uvChannel = make(chan urlData, params.routineNum)
	var storageChannel = make(chan storageBlock, params.routineNum)
	//4. 日志消费
	go readFileLineByLine( params, logChannel )

	//5. 创建一组日志处理
	for i:=0; i<params.routineNum; i++{
		go logConsumer(logChannel, pvChannel, uvChannel)
	}

	//6. 创建UV、PV统计器
	go pvCounter(pvChannel, storageChannel)
	go uvCounter(uvChannel, storageChannel)

	//7. 创建存储器
	go dataStorage(storageChannel)

	//方便调试
	time.Sleep(1000 * time.Second)

}

/**
逐行读取日志，并放到channel中
 */
func readFileLineByLine(params cmdParams, logChannel chan string) error{
	fd, err := os.Open(params.logFilePath)
	if err != nil{
		fmt.Println("ReadFileLineByLine Can`t open file : ", err)
	}
	defer fd.Close()

	count := 0
	bufferRead := bufio.NewReader(fd)
	for{
		line,err := bufferRead.ReadString('\n')
		logChannel <- line
		count++

		if count%(1000 * params.routineNum) == 0{
			//打印日志
			fmt.Printf("ReadFileLineByLine line : %d", count)
		}
		if err != nil{
			if err == io.EOF{//文件读完了
				time.Sleep(3 * time.Second)
				fmt.Printf("ReadFileLineByLine wait, readline: %d", count)
			}else {
				fmt.Println("ReadFileLineByLine read log error :", err)
			}
		}
	}

	return nil
}

/**
消费channel中的日志,解析每行日志数据， 并将数据放到对应的channel中
 */
func logConsumer(logChannel chan string, pvChannel chan urlData, uvChannel chan urlData) error{
	for logStr := range logChannel{
		//切割日志字符串，拿到打点上报数据
		data := cutLogFetchData(logStr)
		//uid 用户标示 ，目前使用md5(refer + ua)模拟
		hasher := md5.New()
		hasher.Write([]byte( data.refer + data.ua))
		uid := hex.EncodeToString(hasher.Sum(nil))

		//将数据放到对应的channel
		uData := urlData{data:data, uid:uid}
		pvChannel <- uData
		uvChannel <- uData

	}

	return nil
}

/**
获取上报信息，格式化成digdata
 */
func cutLogFetchData(logStr string) digData {
	logStr = strings.TrimSpace(logStr)//去掉两边空格
	pos1 := str.IndexOf(logStr, HANDLE_DIG, 0)
	if pos1 == -1{
		return digData{}
	}
	pos1 += len(HANDLE_DIG)
	pos2 := str.IndexOf(logStr, "HTTP/", pos1)
	d := str.Substr(logStr, pos1, pos2 - pos1)

	urlInfo,err := url.Parse("http://localhost/?" + d)
	if err != nil{
		return digData{}
	}
	data := urlInfo.Query()
	return digData{
		time:data.Get("time"),
		url:data.Get("url"),
		refer:data.Get("refer"),
		ua: data.Get("ua"),
	}
}


/**
uv统计 多少人访问，需要去重复
 */
func uvCounter(uvChannel chan urlData, storageChannle chan storageBlock) {
	for data := range uvChannel{
		//HyperLoglog redis
	}
}

/**
pv统计 页面访问量
 */
func pvCounter(pvChannel chan urlData, storageChannle chan storageBlock) {
	for data := range pvChannel{
		sItem := storageBlock{counterType:"pv", storageModel:"ZINCREBY", unode:data.unode}
		storageChannle <- sItem
	}
}

func dataStorage(blocks chan storageBlock) {

}