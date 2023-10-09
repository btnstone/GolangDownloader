package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"golang.org/x/sync/semaphore"
)

const maxConcurrentDownloads = 5

func main() {
	urls := []string{
		"https://techfens.cachefly.net/Default.jpg",
		"http://bjbgp01.baidupcs.com/file/aefc6a8a6n6bc8c58d1a031c0cfe61a6?bkt=en-2e2b5030dd6ff0370e9e5c4492f3176f98fa73511d8076a083af1bba91a99da67cf6d84f3a1e2ba3&fid=1103143671529-250528-478789339097084&time=1696865195&sign=FDTAXUbGERQlBHSKfWaqi-DCb740ccc5511e5e8fedcff06b081203-FxS0TvxPtabDuPdDu9QoNOfFCp4%3D&to=14&size=64266974&sta_dx=64266974&sta_cs=3&sta_ft=exe&sta_ct=6&sta_mt=6&fm2=MH%2CBaoding%2CAnywhere%2C%2C%E5%8C%97%E4%BA%AC%2Cany&ctime=1666278158&mtime=1680848941&resv0=-1&resv1=0&resv2=rlim&resv3=5&resv4=64266974&vuk=1103143671529&iv=2&htype=&randtype=&tkbind_id=0&newver=1&newfm=1&secfm=1&flow_ver=3&pkey=en-022c322c0c6f9e2baf88bd9936a39ebb7ebf001df4e5453db12efc7a48441ff2481cb6c97cf136f8&expires=8h&rt=pr&r=236822159&vbdid=1225670896&fin=diygw-windows-1.0.19-x64.exe&rtype=1&dp-logid=952233870179897422&dp-callid=0.1&tsl=0&csl=0&fsl=-1&csign=HdLrKYLpXnl5WYxk3tUbw%2B9WoUI%3D&so=1&ut=1&uter=0&serv=0&uc=1559654663&ti=4744d9fc935001bf59957ebb1dd70648960c7febe4ba074e&hflag=30&from_type=3&adg=a_388a27b665d995a25ea7de5887ef0f5a&reqlabel=25571201_f_77f76245df01c276512acf5176b91349_-1_f67aaa19eaf757cdf574a014ae266325&by=themis",
		"http://bjbgp01.baidupcs.com/file/5521e177fga65f83e782be032efe7eec?bkt=en-cf7b18a7c51d90783b2ac5336783540c816d5ddd30945e65c44db3699aaa452ce0edd68f2205656a&fid=1103143671529-250528-1124872726617641&time=1696865725&sign=FDTAXUbGERQlBHSKfWaqi-DCb740ccc5511e5e8fedcff06b081203-Wj9aFBcoHtZdEjEiOuf3%2FkhXTbY%3D&to=14&size=148260984&sta_dx=148260984&sta_cs=22&sta_ft=exe&sta_ct=5&sta_mt=5&fm2=MH%2CBaoding%2CAnywhere%2C%2C%E5%8C%97%E4%BA%AC%2Cany&ctime=1689728742&mtime=1692683915&resv0=-1&resv1=0&resv2=rlim&resv3=5&resv4=148260984&vuk=1103143671529&iv=2&htype=&randtype=&tkbind_id=0&newver=1&newfm=1&secfm=1&flow_ver=3&pkey=en-bd7826cba2123b8fb83e0d8f3358a14c9e6016815b876ca17544b0bff477bf8d5c9996a619a5ae15&expires=8h&rt=pr&r=510197662&vbdid=1225670896&fin=jdk-11.0.20_windows-x64_bin.exe&rtype=1&dp-logid=1004919713919720797&dp-callid=0.1&tsl=0&csl=0&fsl=-1&csign=HdLrKYLpXnl5WYxk3tUbw%2B9WoUI%3D&so=1&ut=1&uter=0&serv=0&uc=1559654663&ti=05df9239daa40647a92b2c09f88158bf055ae8dcc2aa8ed3305a5e1275657320&hflag=30&from_type=3&adg=a_388a27b665d995a25ea7de5887ef0f5a&reqlabel=25571201_f_77f76245df01c276512acf5176b91349_-1_f67aaa19eaf757cdf574a014ae266325&fpath=Java+SE+Development+Kit+11.0.20+%28LTS%29&by=themis",
		// ... 可以继续添加更多的URL来模拟多个文件的下载
	}

	var wg sync.WaitGroup
	p := mpb.New() // 创建主进度条容器

	// 初始化信号量
	sem := semaphore.NewWeighted(maxConcurrentDownloads)

	for index, url := range urls {
		wg.Add(1)
		sem.Acquire(context.Background(), 1) // 获取信号量

		go func(url string, index int) {
			defer wg.Done()
			defer sem.Release(1) // 释放信号量

			destPath := filepath.Join("download", fmt.Sprintf("Default_%d.jpg", index+1))
			err := downloadFileWithProgress(url, destPath, p)
			if err != nil {
				fmt.Printf("Error downloading file from %s: %v\n", url, err)
			}
		}(url, index)
	}

	wg.Wait()
	p.Wait() // 等待所有进度条完成
}

func downloadFileWithProgress(url, destPath string, p *mpb.Progress) error {
	// 创建目录
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 设置自定义User-Agent
	customUserAgent := "pan.baidu.com"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", customUserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 获取文件名
	filename := filepath.Base(destPath)

	bar := p.AddBar(
		resp.ContentLength,
		mpb.PrependDecorators(
			decor.Name(filename+" "), // 添加文件名到进度条前面
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_MMSS, 60),
			decor.Name(" ] "),
			decor.AverageSpeed(decor.UnitKB, "% .2f"),
		),
	)

	proxyReader := bar.ProxyReader(resp.Body)
	defer proxyReader.Close()

	_, err = io.Copy(out, proxyReader)
	return err
}
