package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/go-resty/resty/v2"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	//"strconv"
)

var (
	title = `<title>([\s\S]+?)</title>`
	conf  = int(0)  // 配置、终端默认设置
	bg    = int(0)  // 背景色、终端默认设置
	green = int(32) //绿色
	//read = int(31)
	aliveUrlNumber = int(0)
)

type Info struct {
	Code       int
	Title      string
	Url        string
	Bodylength int
}

type ChanNode struct {
	Url     string
	Trytime int
}

func writeFile(resultChan chan Info, inputFileName string, outputFileName *string) string {
	t := time.Now().Format("2006-01-02-15-04-05")
	// jishu
	//var aliveUrlNumber int = 0
	// 获取文件名称
	if *outputFileName == "" {
		*outputFileName = inputFileName
	}
	outputFileNametmp := filepath.Base(*outputFileName)
	outputFileNameWithoutExt := outputFileNametmp[:len(outputFileNametmp)-len(filepath.Ext(outputFileNametmp))]
	*outputFileName = t + "-" + outputFileNameWithoutExt + ".txt"
	f, err := os.OpenFile(*outputFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	//else {
	//	_, err = f.Write([]byte(wireteString + "\n"))
	//}
	for nodee := range resultChan {
		_, err = f.WriteString(fmt.Sprintf("%s,%d,%s\n", nodee.Url, nodee.Code, nodee.Title))
		//_, err = f.WriteString(nodee.Url + string(nodee.Title) + "\n")
		//fmt.Sprintf("%s,%s,%s\n", nodee.Url, nodee.Code, nodee.Title)
		//_, err = f.Write([]byte(fmt.Sprintf("%s,%d,%s\n", nodee.Url, nodee.Code, string(nodee.Title))))
		if err != nil {
			fmt.Printf("%s write to outputfile failed", nodee.Url)
		}
		aliveUrlNumber += 1
	}

	return *outputFileName
}

func get(urlchan chan ChanNode, resultChan chan Info, retryNum int, wg *sync.WaitGroup) {
	defer wg.Done()
	//for node := range urlchan {
	time.Sleep(time.Second)
	//for {
	for node := range urlchan {
		//node, ok := <-urlchan // 通道关闭后再取值ok=false
		//if !ok {
		//	break
		//}
		//fmt.Println(node)
		//continue
		var info Info
		url := node.Url
		//if node.Trytime <= retryNum {
		//	continue
		//}
		if !strings.Contains(url, "http") {
			url = "http://" + url //默认为http
		}
		client := resty.New().SetTimeout(5 * time.Second).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		client.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
		resp, err := client.R().EnableTrace().Get(url)
		//aa:=client.R().EnableTrace()
		//aa.Get()
		if err != nil {
			//fmt.Println(err)
			//fmt.Println(err.Error())
			node.Trytime += 1
			//urlchan <- node 不能这么写是因为本来管道就是满的，可以通过len(chan)获取元素个数，满的这个就加不进去，这个协程就会阻塞住，最终全部协程都阻塞住了，就没有协程去处理新任务了，就会卡住
			//循环尝试
			for node.Trytime <= retryNum {
				// err当前协程停一秒再试一次
				time.Sleep(time.Second)
				resp, err = client.R().EnableTrace().Get(url)
				//fmt.Println(resp, err)
				if err != nil {
					node.Trytime += 1
				} else {
					//fmt.Println(url, resp, err)
					break
				}
			}
			if err != nil {
				continue
			}
		}
		if strings.Contains(string(resp.Body()), "HTTP request was sent to HTTPS") {
			url = strings.Replace(url, "http", "https", -1)
			resp, err = client.R().EnableTrace().Get(url)
			if err != nil {
				//fmt.Println(err)
				//fmt.Println(err.Error())
				continue
			}
		}
		info.Code = resp.StatusCode()
		//str := resp.Body()
		//body := string(str)
		//body := mahonia.NewDecoder("gbk").ConvertString(string(str))
		body := mahonia.NewEncoder("utf-8").ConvertString(resp.String())
		//body := resp.String()
		//body, _, _ := charset.DetermineEncoding(str, "")
		if strings.Contains(body, "<title>") {
			re := regexp.MustCompile(title)
			title_name := re.FindAllStringSubmatch(body, 1)
			if len(title_name) == 0 {
				continue
			}

			info.Title = title_name[0][1]
		}
		info.Url = url
		info.Bodylength = len(body)
		resultChan <- info
		fmt.Printf("%c[%d;%d;%dm%s%c[0m%s", 0x1B, conf, bg, green, "[+]", 0x1B, fmt.Sprintf(info.Url+" "+info.Title+" %d %d\n", info.Code, info.Bodylength))

	}

}
func readFromFile(path string, urlchan chan ChanNode) {
	// 读完及时关闭管道，否则其他协程会在没关闭的时候依然从空的管道中读取数据，进而产生死锁
	defer close(urlchan)
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("[-] failed %s", err.Error())
		os.Exit(0)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		urlchan <- ChanNode{Url: strings.TrimSpace(scanner.Text()), Trytime: 0}
	}
}

func main() {
	var CurrentVersion = "1.0 #dev"
	fmt.Printf(`		Url_alive_scan %s
					by komomon
`, CurrentVersion)
	//os.Exit(1)
	var wg sync.WaitGroup
	var urlChan = make(chan ChanNode, 10)
	var resultChan = make(chan Info, 20)
	var inputFileName string
	var outputFileName string
	var threads int
	var retry int
	flag.StringVar(&inputFileName, "i", "urls.txt", "the file of the targets")
	flag.IntVar(&threads, "t", runtime.NumCPU(), "the threads of the program")
	flag.IntVar(&retry, "retry", 2, "after failure, the number of request attempts")
	flag.StringVar(&outputFileName, "o", "", "the result file,default time-inputfilename.txt")

	flag.Parse()
	//if inputFileName == "" {
	//	fmt.Printf("please input the path of the targets,-h for help")
	//	return
	//}
	//inputFileName := "url.txt"
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go get(urlChan, resultChan, retry, &wg)
	}
	go readFromFile(inputFileName, urlChan)
	time.Sleep(time.Second)
	go writeFile(resultChan, inputFileName, &outputFileName)

	wg.Wait()
	//close(urlChan)
	fmt.Println("[+] Outputfile:", outputFileName)
	fmt.Println("[+] Alive url number:", aliveUrlNumber)
}
