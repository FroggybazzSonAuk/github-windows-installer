// MIT License
//
// Copyright (c) 2017, b3log.org & hacpai.com
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/b3log/wide/log"
	"github.com/b3log/wide/util"
)

var logger *log.Logger
var waitGroup = sync.WaitGroup{}
var dir = "github-windows"

func main() {
	log.SetLevel("debug")
	logger = log.NewLogger(os.Stdout)

	//os.Setenv("HTTP_PROXY", "http://127.0.0.1:8087")
	//os.Setenv("HTTPS_PROXY", "https://127.0.0.1:8087")

	os.Remove(dir) // 先删除可能已经存在的
	os.Mkdir(dir, 0775)

	host := "http://github-windows.s3.amazonaws.com"

	// 下载应用元数据文件
	metadataURL := host + "/GitHub.application"
	logger.Info("Getting metadata [" + metadataURL + "]")
	res, err := http.Get(metadataURL)
	if nil != err {
		logger.Error("Get metadata failed: ", err)

		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if nil != err {
		logger.Error("Parse metadata failed: ", err)

		return
	}

	metadata := string(body)
	ioutil.WriteFile(dir+"/GitHub.application", body, 0644)

	manifestURI := metadata
	manifestURI = strings.Split(manifestURI, "codebase=\"")[1]
	manifestURI = strings.Split(manifestURI, "\"")[0]
	manifestURI = strings.Replace(manifestURI, "\\", "/", -1)

	// 下载包描述文件
	u, err := url.Parse(host + "/" + manifestURI)
	if nil != err {
		logger.Error("Parse manifest failed: ", err)

		return
	}

	manifestURL := u.String()
	ver := strings.Split(manifestURL, "Files/")[1]
	ver = strings.Split(ver, "/GitHub.exe")[0]
	verURL := strings.Split(manifestURL, "/GitHub.exe")[0] // i.e. http://github-windows.s3.amazonaws.com/Application%20Files/GitHub_3_3_4_0

	logger.Info("Getting manifest [" + manifestURL + "]")
	res, err = http.Get(manifestURL)
	if nil != err {
		logger.Error("Get manifest failed: ", err)

		return
	}

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	if nil != err {
		logger.Error("Parse manifest failed: ")

		return
	}

	manifest := string(body)
	logger.Trace(manifest)

	err = os.MkdirAll(dir+"/Application Files/"+ver, 0775)
	if nil != err {
		logger.Error("Make [Application Files] folder failed: ", err)

		return
	}

	err = ioutil.WriteFile(dir+"/Application Files/"+ver+"/"+"GitHub.exe.manifest", body, 0775)
	if nil != err {
		logger.Error("Save manifest failed: ", err)
	}

	// 解析所需包文件下载路径
	parts := strings.SplitAfter(manifest, "codebase=\"")
	deploys := []string{}
	for _, part := range parts[1:] {
		part = strings.Split(part, "\"")[0]
		part = strings.Replace(part, "\\", "/", -1)
		deploys = append(deploys, verURL+"/"+part+".deploy")
	}

	// 解析所需资源文件下载路径
	parts = strings.SplitAfter(manifest, "file name=\"")
	for _, part := range parts[1:] {
		part = strings.Split(part, "\"")[0]
		part = strings.Replace(part, "\\", "/", -1)
		deploys = append(deploys, verURL+"/"+part+".deploy")
	}

	// 并发下载
	for i, deploy := range deploys {
		logger.Debug(i, ". "+deploy)

		waitGroup.Add(1)
		go download(deploy, ver)
	}

	waitGroup.Wait()
	logger.Info("All files already downloaded, getting start zip them")

	// 打包
	zipPath := "github-windows.zip"
	os.Remove(zipPath) // 先删除可能已经存在的

	zipFile, err := util.Zip.Create(zipPath)
	if nil != err {
		logger.Error("Create zip file failed: ", err)

		return
	}

	zipFile.AddDirectory(dir, dir)
	if nil != zipFile.Close() {
		logger.Error("Create zip file failed: ", err)

		return
	}

	logger.Info("Created GitHub Windows setup :p")
}

func download(url string, ver string) {
	defer waitGroup.Done()
	logger.Info("Getting file [" + url + "]")

	res, err := http.Get(url)
	if nil != err {
		logger.Error("Get file failed: ", err)

		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if nil != err {
		logger.Error("Get file failed failed: ", err)

		return
	}

	path := strings.Split(url, ver)[1]
	dir := filepath.Dir(path)
	if "/" != dir {
		err = os.MkdirAll(dir+"/Application Files/"+ver+dir, 0755)
		if nil != err {
			logger.Error("Make [Application Files] folder failed: ", err)

			return
		}
	}

	deployPath := dir + "/Application Files/" + ver + path
	err = ioutil.WriteFile(deployPath, body, 0664)
	if nil != err {
		logger.Error("Save file failed: ", err)

		return
	}

	logger.Info("Saved file to [" + deployPath + "]")
}
