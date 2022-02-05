package main

import "github.com/aliyun/packagist-mirror/util"

func main() {
	util.Execute()
	util.Wg.Wait()
}
