package cachenet

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"
	"github.com/zzossig/rabbit"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

//由编码后的url.Values获取cache
// data := url.Values{
// 	"Tid": {"10"},
// 	"wd":  {wd},
// }
// dataStr := data.Encode()
func CacheDataQuery(base_url string, url_data string) (*html.Node, error) {
	reader, err := cache.ReaderData(base_url, url_data)
	if err != nil {
		return nil, err
	}

	node, err := htmlquery.Parse(reader)

	reader.Close()

	if err != nil {
		return nil, err
	}

	return node, err
}

//由URL获取cache
func CacheQuery(req_url string) (*html.Node, error) {
	reader, err := cache.Reader(req_url)
	if err != nil {
		return nil, err
	}

	node, err := htmlquery.Parse(reader)

	reader.Close()

	if err != nil {
		return nil, err
	}

	return node, err
}

var count_max int = 0

//由URL获取cache
func CacheQueryString(req_url string) (string, error) {
	reader, err := cache.Reader(req_url)
	if err != nil {
		return "", err
	}

	data := make([]byte, 102400)

	count, err := reader.Read(data)
	if err != nil {
		return "", err
	}

	reader.Close()

	if count > count_max {
		count_max = count
	}

	if err != nil {
		return "", err
	}

	return string(data), nil
}

//由URL获取cache
func CacheQueryXml(req_url string) (*xmlquery.Node, error) {
	reader, err := cache.Reader(req_url)
	if err != nil {
		return nil, err
	}

	doc, err := xmlquery.Parse(reader)

	reader.Close()

	if err != nil {
		return nil, err
	}

	return doc, nil
}

func CacheQueryXpath(req_url string, xpath_str string) []*html.Node {
	reader, err := cache.Reader(req_url)
	if err != nil {
		panic(err)
	}

	node, err := htmlquery.Parse(reader)

	reader.Close()

	if err != nil {
		panic(err)
	}

	x := rabbit.New()
	x.SetDocN(node)
	if len(x.Errors()) > 0 {
		panic(x.Errors())
	}
	x.Eval(xpath_str)
	if len(x.Errors()) > 0 {
		panic(x.Errors())
	}
	return x.NodeAll()
}

func CacheQueryOneXpath(req_url string, xpath_str string) *html.Node {
	reader, err := cache.Reader(req_url)
	if err != nil {
		panic(err)
	}

	node, err := htmlquery.Parse(reader)

	reader.Close()

	if err != nil {
		panic(err)
	}

	x := rabbit.New()
	x.SetDocN(node)
	if len(x.Errors()) > 0 {
		panic(x.Errors())
	}
	x.Eval(xpath_str)
	if len(x.Errors()) > 0 {
		panic(x.Errors())
	}
	return x.Node()
}

// Generated by curl-to-Go: https://mholt.github.io/curl-to-go
// data := url.Values{
// 	"Tid": {"10"},
// 	"wd":  {wd},
// }
// dataStr := data.Encode()
func PostRequest(req_url string, url_data string) (io.ReadCloser, error) {
	req, err := http.NewRequest("POST", req_url, strings.NewReader(url_data))
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (;) / (,) / / ")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Set("Content-Length", strconv.Itoa(len(url_data)))

	resp, err := Request(req)
	if err != nil {
		//resp.Body.Close()
		return nil, err
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	html_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	html_bytes_reader := bytes.NewReader(html_bytes)
	html_encoding, _, _ := charset.DetermineEncoding(html_bytes, "")

	utf8_reader := transform.NewReader(html_bytes_reader, html_encoding.NewDecoder())

	html_bytes_utf8, err := ioutil.ReadAll(utf8_reader)
	if err != nil && html_bytes_utf8 == nil {
		panic(err)
	}

	html_utf8_string := strings.ReplaceAll(string(html_bytes_utf8), "gb2312", "utf-8")

	resp_body_utf8 := ioutil.NopCloser(bytes.NewReader([]byte(html_utf8_string)))

	return resp_body_utf8, nil
}

//https://dreamerjonson.com/2019/01/22/golang-48-gbkatoUtf8/index.html
//探测encoding
//https://golangnote.com/topic/195.html
//Golang io.ReadCloser 和[]byte 相互转化

//由字的url获取页面
func GetRequest(req_url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (;) / (,) / / ")

	resp, err := Request(req)
	if err != nil {
		//resp.Body.Close()
		return nil, err
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	html_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	html_bytes_reader := bytes.NewReader(html_bytes)
	html_encoding, _, _ := charset.DetermineEncoding(html_bytes, "")

	utf8_reader := transform.NewReader(html_bytes_reader, html_encoding.NewDecoder())

	html_bytes_utf8, err := ioutil.ReadAll(utf8_reader)
	if err != nil && html_bytes_utf8 == nil {
		panic(err)
	}

	html_utf8_string := strings.ReplaceAll(string(html_bytes_utf8), "gb2312", "utf-8")

	resp_body_utf8 := ioutil.NopCloser(bytes.NewReader([]byte(html_utf8_string)))

	return resp_body_utf8, nil
}
