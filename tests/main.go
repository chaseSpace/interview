package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
)

func main() {
	num := 6
	for index := 0; index < num; index++ {
		resp, _ := http.Get("https://www.baidu.com")
		_, _ = ioutil.ReadAll(resp.Body)
	}
	fmt.Printf("此时goroutine个数= %d\n", runtime.NumGoroutine())
}
