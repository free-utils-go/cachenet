package cachenet

import (
	"crypto/sha256"
	"fmt"
	"net/url"
)

func Hash(url string) string {
	sum256 := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", sum256)
}

//URL拼接
func UrlMerge(base_url string, wd_url string) string {
	base_u, _ := url.Parse(base_url)
	u_ref, _ := base_u.Parse(wd_url)

	return u_ref.String()
}
