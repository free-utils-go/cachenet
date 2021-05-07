package cachenet

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Cache ...
type Cache struct {
	Tmp string
}

var cache *Cache

// NewCache ...
func NewCache(tmp string) *Cache {
	if cache != nil {
		return cache
	}

	var path_ string

	if filepath.IsAbs(tmp) {
		path_ = tmp
	} else {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		path_ = path.Join(path.Dir(dir), tmp)
	}

	_ = os.MkdirAll(tmp, os.ModePerm)
	cache = &Cache{Tmp: path_}

	return cache
}

//从unicode获得cache文件句柄
func (c *Cache) ReaderData(req_url string, url_data string) (reader io.ReadCloser, e error) {
	wd_str := c.PrepareDataPath(req_url, url_data)

	stat, e := os.Stat(wd_str)

	if (e == nil && stat.Size() != 0) || !os.IsNotExist(e) {
		open, e := os.Open(wd_str)
		if e != nil {
			panic(e)
		}
		return open, nil
	}

	closer, err := PostRequest(req_url, url_data)
	if err != nil {
		return nil, err
	}

	return c.CacheData(closer, req_url, url_data)
}

//从URL获得cache文件句柄
func (c *Cache) Reader(req_url string) (reader io.ReadCloser, e error) {
	file_path := c.PrepareGetPath(req_url)

	stat, e := os.Stat(file_path)

	if (e == nil && stat.Size() != 0) || !os.IsNotExist(e) {
		open, e := os.Open(file_path)
		if e != nil {
			panic(e)
		}
		return open, nil
	}

	closer, err := GetRequest(req_url)
	if err != nil {
		return nil, err
	}

	PreparePath(file_path)

	return c.Cache(closer, req_url)
}

//从URL和http响应内容句柄获得cache文件句柄
func (c *Cache) Cache(closer io.ReadCloser, req_url string) (io.ReadCloser, error) {
	file_path := c.PrepareGetPath(req_url)

	stat, e := os.Stat(file_path)

	if (e == nil && stat.Size() != 0) || !os.IsNotExist(e) {
		return nil, os.ErrExist
	}
	file, e := os.OpenFile(file_path, os.O_TRUNC|os.O_CREATE|os.O_RDONLY|os.O_SYNC, os.ModePerm)
	if e != nil {
		panic(e)
	}

	_, e = io.Copy(file, closer)
	if e != nil {
		panic(e)
	}
	file.Close()
	closer.Close()

	cachefile, e := os.Open(file_path)
	if e != nil {
		panic(e)
	}

	return cachefile, nil
}

//从Unicode和http响应内容句柄获得cache文件句柄
func (c *Cache) CacheData(closer io.ReadCloser, req_url string, url_data string) (io.ReadCloser, error) {
	file_path := c.PrepareDataPath(req_url, url_data)

	stat, e := os.Stat(file_path)

	if (e == nil && stat.Size() != 0) || !os.IsNotExist(e) {
		return nil, os.ErrExist
	}
	file, e := os.OpenFile(file_path, os.O_TRUNC|os.O_CREATE|os.O_RDONLY|os.O_SYNC, os.ModePerm)
	if e != nil {
		panic(e)
	}

	_, e = io.Copy(file, closer)
	if e != nil {
		panic(e)
	}
	file.Close()
	closer.Close()

	cachefile, e := os.Open(file_path)
	if e != nil {
		panic(e)
	}

	return cachefile, nil
}

// URL存为cache文件
func (c *Cache) Get(req_url string) (e error) {
	file_path := c.PrepareGetPath(req_url)

	stat, e := os.Stat(file_path)

	if (e == nil && stat.Size() != 0) || !os.IsNotExist(e) {
		return os.ErrExist
	}

	closer, err := GetRequest(req_url)
	if err != nil {
		log.Fatal(err)
	}

	file, e := os.OpenFile(file_path, os.O_TRUNC|os.O_CREATE|os.O_RDONLY|os.O_SYNC, os.ModePerm)
	if e != nil {
		return e
	}
	written, e := io.Copy(file, closer)
	if e != nil {
		return e
	}
	//ignore written
	_ = written
	closer.Close()
	return nil
}

//为目标路径创建文件夹，返回目标路径的绝对路径，url_data为url.Values的encode()过的形态
func (c *Cache) PrepareDataPath(req_url string, url_data string) string {
	file_path := c.dataPath(req_url, url_data)

	file_path_abs, filename := PreparePath(file_path)

	return fixSuffix(file_path_abs, filename)
}

//为目标路径创建文件夹，返回目标路径的绝对路径，url_data为url.Values的encode()过的形态
func (c *Cache) dataPath(req_url string, url_data string) string {
	if len(url_data) == 0 {
		return c.getPath(req_url)
	}

	req_url_path := strings.Split(req_url, "://")[1]

	req_url_path_sub := strings.ReplaceAll(url_data, "=", "_")

	req_url_path_sub = strings.ReplaceAll(req_url_path_sub, "&", "__")

	req_url_path = req_url_path + "/" + req_url_path_sub

	return path.Join(path.Dir(c.Tmp), req_url_path)

}

//为目标路径创建文件夹，返回目标路径的绝对路径
func (c *Cache) PrepareGetPath(req_url string) string {
	file_path := c.getPath(req_url)

	file_path_abs, filename := PreparePath(file_path)

	return fixSuffix(file_path_abs, filename)
}

//为目标路径创建文件夹，返回目标路径的绝对路径
func (c *Cache) getPath(req_url string) string {
	req_url_path := strings.Split(req_url, "://")[1]

	req_url_path = strings.ReplaceAll(req_url_path, "?", "/")

	req_url_path = strings.ReplaceAll(req_url_path, "=", "_")

	req_url_path = strings.ReplaceAll(req_url_path, "&", "__")

	return path.Join(path.Dir(c.Tmp), req_url_path)
}

//修复无扩展名路径
func fixSuffix(file_path_abs string, filename string) string {
	if !strings.Contains(filename, ".") {
		file_path_abs += ".htm"
	}
	return file_path_abs
}

//为目标路径创建文件夹，返回目标路径的绝对路径和文件名
func PreparePath(path string) (string, string) {
	file_path_abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	if !strings.Contains(file_path_abs, ":") {
		fmt.Println("")
	}

	dir, filename := filepath.Split(file_path_abs)

	dir_stat, _ := os.Stat(dir)

	if dir_stat != nil && !dir_stat.IsDir() {
		panic("cache dir conflict to a file")
	}

	_ = os.MkdirAll(dir, os.ModePerm)
	return file_path_abs, filename
}

// URL对应的缓存文件转存成指定文件
func (c *Cache) Save(req_url string, to string) (written int64, e error) {
	file_path := c.PrepareGetPath(req_url)
	info, e := os.Stat(file_path)
	if e != nil && os.IsNotExist(e) {
		panic(errors.Wrap(e, "cache get error"))
	}
	if info.IsDir() {
		panic("cache get a dir")
	}

	abs_to, _ := PreparePath(to)

	file, e := os.Open(file_path)
	if e != nil {
		panic(e)
	}

	pj := path.Join(abs_to)

	openFile, e := os.OpenFile(pj, os.O_TRUNC|os.O_CREATE|os.O_RDONLY|os.O_SYNC, os.ModePerm)
	if e != nil {
		panic(e)
	}
	return io.Copy(openFile, file)
}
