// http_server project main.go
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"

	//"http_moni/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	port         = "66"  //默认端口
	wwwroot      = "www" //网站根目录
	fileListHead = "<html><meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" /><div align=\"center\"><br/><p><br/><p>"
	fileListEnd  = "</div></html>"
	uploadDir    = "www/upload/"
	downloadDir  = "www/down/"
)

type ResultInfo struct {
	Success int
	Message string
	Header  string
	Time    string
}

type TestInfo struct {
	Version string
	Url     string
	Hash    string
}

func main() {
	if !PathExists(uploadDir) {
		os.Mkdir(uploadDir, os.ModePerm)
	}
	if !PathExists(downloadDir) {
		os.Mkdir(downloadDir, os.ModePerm)
	}
	//go RunClient()
	go startHTTP()
	ti := time.Tick(time.Second * 60)
	for {
		<-ti
	}
}

func startHTTP() {
	mux := http.NewServeMux()
	mux.HandleFunc("/filedown", FileHandle)
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/moni", MoniHandler)
	mux.HandleFunc("/bs", getFileList)
	mux.HandleFunc("/uploadfile", UploadHandle)
	mux.HandleFunc("/downloadfile", DownloadHandle)
	mux.HandleFunc("/testdown", TestDownHandle)
	log.Println("HTTP Server Listening to 0.0.0.0: " + port)
	server := &http.Server{Addr: "0.0.0.0:" + port, Handler: mux}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

func FileHandle(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		w.Write([]byte("error:" + err.Error()))
		return
	}
	path = filepath.Dir(path)
	r.URL.Path = "/" + r.URL.RawQuery
	http.ServeFile(w, r, path+"/"+wwwroot+"/"+r.URL.Path)
}

func MoniHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var paramNum = 0
	var headerNum = 0
	result := &ResultInfo{
		Message: "Empty",
		Header:  "Empty-Header",
		Success: 1,
		Time:    "0.001s",
	}
	//fmt.Println(r.Form)
	targetUrl := r.PostFormValue("url")
	method := r.PostFormValue("method")
	header := r.PostFormValue("headers")
	r.PostForm.Del("url")
	r.PostForm.Del("method")
	for key, _ := range r.PostForm {
		if strings.Contains(key, "params") {
			paramNum++
		}
		if strings.Contains(key, "headers") {
			headerNum++
		}

	}

	fmt.Println(r.PostForm)
	targetHeader := make(map[string]string, 64)
	//var targetData url.Values
	for i := 0; i < headerNum/2; i++ {
		name := r.PostForm["headers["+strconv.Itoa(i)+"][name]"][0]
		value := r.PostForm["headers["+strconv.Itoa(i)+"][value]"]
		targetHeader[name] = value[0]
	}

	start := time.Now().UnixNano()
	if method == "POST" {
		var targetData = ""
		for i := 0; i < paramNum/2; i++ {
			name := r.PostForm["params["+strconv.Itoa(i)+"][name]"][0]
			value := r.PostForm["params["+strconv.Itoa(i)+"][value]"]
			//targetData[name] = value
			targetData = targetData + name + "=" + value[0] + "&"
		}
		resp, err := httpPOST(targetUrl, targetHeader, targetData) //http.PostForm(targetUrl, targetData)
		if err != nil {
			fmt.Printf("post data error:%v\n", err)
		} else {
			fmt.Println("post a data successful.\r\n", targetData)
			respBody, _ := ioutil.ReadAll(resp.Body)
			/*aa, err := gzip.GzipDecode(respBody)
			if err == nil {
				fmt.Println("aa:", string(aa), "sec:", string(respBody))
			} else {
				fmt.Println("decode-error:", err)
			}*/
			result.Message = string(respBody)
			result.Header = resp.Proto + " " + resp.Status + "\r\n"
			//result.Header = resp.Proto + " " + resp.Status + "\r\nContent-Type:" + resp.Header["Content-Type"][0] + "\r\nDate:" + resp.Header["Date"][0] + "\r\nContent-Length:" + resp.Header["Content-Length"][0]
		}
	} else if method == "GET" {
		var targetData = make(url.Values, 64)
		for i := 0; i < paramNum/2; i++ {
			name := r.PostForm["params["+strconv.Itoa(i)+"][name]"][0]
			value := r.PostForm["params["+strconv.Itoa(i)+"][value]"]
			targetData[name] = value

		}
		resp, err := httpGET(targetUrl, targetHeader, targetData) //http.PostForm(targetUrl, targetData)
		if err != nil {
			fmt.Printf("post data error:%v\n", err)
		} else {
			fmt.Println("get a data successful.")
			respBody, _ := ioutil.ReadAll(resp.Body)
			result.Message = string(respBody)
			result.Header = resp.Proto + " " + resp.Status + "\r\n"
			//result.Header = resp.Proto + " " + resp.Status + "\r\nContent-Type:" + resp.Header["Content-Type"][0] + "\r\nDate:" + resp.Header["Date"][0] + "\r\nContent-Length:" + resp.Header["Content-Length"][0]

		}
	}
	result.Time = fmt.Sprintf("%v ms", (time.Now().UnixNano()-start)/1e6)
	ResponseCheck(result, w)
	fmt.Println(header)
}

func httpPOST(url string, header map[string]string, data string) (resp *http.Response, err error) {
	/*tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}*/
	//fmt.Println("data:", data)
	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewReader([]byte(data)))
	if err != nil {
		fmt.Println("err:", err)
	}
	for key, value := range header {
		request.Header.Add(key, value)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Connection", "Keep-Alive")
	resp, err = client.Do(request)
	return resp, err
}

func httpGET(url string, header map[string]string, data url.Values) (resp *http.Response, err error) {

	if len(data) > 0 {
		url = url + "?"
	}
	for name, value := range data {
		url = url + name + "=" + value[0] + "&"
	}
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	for key, value := range header {
		request.Header.Add(key, value)
	}
	request.Header.Set("Connection", "Keep-Alive")
	fmt.Println(url)
	resp, err = client.Do(request)
	/*tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err = client.Get(url)*/
	//resp, err = http.Get(url)
	return resp, err

}

func TestDownHandle(w http.ResponseWriter, r *http.Request) {
	version := r.FormValue("version")
	VersionCode, _ := strconv.ParseFloat(version, 32)

	if VersionCode < 1.0 && VersionCode > 0 {
		testInfo := &TestInfo{
			Version: "1.0",
			Url:     "http://192.168.123.46:66/upload/FODCreate.exe",
			Hash:    "123456",
		}
		ResponseCheck(testInfo, w)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	/*path := r.URL.Path
	if path == "/" {
		w.Header().Set("content-Type", "text/html")
		w.Header().Set("Server", "nosib")
		content := getHtmlFile("index.html")
		w.Write([]byte(content))
	} else {
		if strings.Contains(path[1:], ".html") {
			w.Header().Set("content-Type", "text/html")
		}
		if strings.Contains(path[1:], ".css") {
			w.Header().Set("content-Type", "text/css")
		}
		if strings.Contains(path[1:], ".js") {
			w.Header().Set("Content-Type", "text/javascript")
		}
		content := getHtmlFile(path[1:])

		w.Write([]byte(content))
	}*/
	/*if strings.HasPrefix(r.URL.String(), "/down/") {
		had := http.StripPrefix("/down/", http.FileServer(http.Dir("www/down")))
		had.ServeHTTP(w, r)
		return
	} else if strings.HasPrefix(r.URL.String(), "/upload/") {
		had := http.StripPrefix("/upload/", http.FileServer(http.Dir("www/upload")))
		had.ServeHTTP(w, r)
		return
	} else*/

	if strings.HasPrefix(r.URL.String(), "/") {
		had := http.StripPrefix("/", http.FileServer(http.Dir(wwwroot)))
		had.ServeHTTP(w, r)
	} else {
		http.Error(w, "404 not found", 404)
	}

}

func getHtmlFile(path string) (fileHtml string) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	realPath := dir + "/" + wwwroot + "/" + path
	if PathExists(realPath) {
		file, err := os.Open(realPath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fileContent, _ := ioutil.ReadAll(file)
		fileHtml = string(fileContent)

	} else {
		fileHtml = "404 page not found"
	}

	return fileHtml
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func getFileList(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fileListHead))
	dir_list, e := ioutil.ReadDir("./www/down/")
	if e != nil {
		log.Fatal("read dir error ", e.Error())
		return
	}
	for _, v := range dir_list {
		if !v.IsDir() {
			w.Write([]byte("<a href=\"down/" + v.Name() + "\"><font color=\"red\" size=\"20\">" + v.Name() + "</font></a><br/><p><br/><p> \n"))
		}
	}
	upload_list, err := ioutil.ReadDir("./www/upload/")
	if err != nil {
		log.Fatal("read dir error ", e.Error())
		return
	}
	for _, v := range upload_list {
		if !v.IsDir() {
			w.Write([]byte("<a href=\"upload/" + v.Name() + "\"><font color=\"red\" size=\"20\">" + v.Name() + "</font></a><br/><p><br/><p> \n"))
		}
	}
	w.Write([]byte(fileListEnd))
}

func UploadHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		w.Write([]byte(fileListHead))
		r.ParseMultipartForm(32 << 20)
		fileHeaders, findFile := r.MultipartForm.File["fileName"]
		if !findFile || len(fileHeaders) == 0 {
			log.Println("file count == 0.")
			w.Write(([]byte)("没有上传文件"))
			return
		}
		//fmt.Printf("成功上传了%d个文件 %t\n", len(fileHeaders), findFile)
		for _, fileHeader := range r.MultipartForm.File["fileName"] {
			fileName := fileHeader.Filename
			//w.Write([]byte(fileName + "\n"))
			file, err := fileHeader.Open()
			if err != nil {
				log.Println("file error:", err.Error())
			}
			//defer file.Close()
			outputFilePath := uploadDir + fileName
			writer, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE, 0666)
			if err == nil {
				io.Copy(writer, file)
			}
			writer.Close()
			file.Close()
		}
		rs := fmt.Sprintf("成功上传了%d个文件 \n", len(fileHeaders))
		w.Write([]byte(rs))
		w.Write([]byte(fileListEnd))
	} else {
		w.Write([]byte("not upload file !"))
	}
}

func DownloadHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var downNum = 0
		w.Write([]byte(fileListHead))
		r.ParseForm()
		for _, url := range r.PostForm["fileurl"] {
			//fmt.Printf("%d： %s\n", i, url)
			if legalUrl(url) {
				//fmt.Printf("filename： %s\n", url)
				if downloadFile(url) {
					downNum++
				}
			}
		}
		rs := fmt.Sprintf("成功下载了%d个文件 \n", downNum)
		w.Write([]byte(rs))
		w.Write([]byte(fileListEnd))
	} else {
		w.Write([]byte("not legal url !"))
	}
}

func legalUrl(url string) bool {
	if strings.Contains(url, "http") {
		return true
	}
	return false
}

func getFilename(url string) string {
	index := strings.LastIndex(url, "/")
	filename := string([]rune(url)[index+1:])
	return filename
}

func downloadFile(url string) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(url)
	if err != nil {
		return false
	}
	filename := getFilename(url)
	f, err := os.Create(downloadDir + filename)
	defer f.Close()
	if err != nil {
		return false
	}
	io.Copy(f, res.Body)
	return true
}

func ResponseCheck(rs interface{}, w http.ResponseWriter) {
	json, _ := json.Marshal(rs)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Server", "nosib")
	w.Write(json)

}
