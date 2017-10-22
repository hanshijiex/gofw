package main

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"../../gofw"
)

type User struct {
	User string
	Age int64
}

func test(ctx *gofw.Context) {
	userid := ctx.Param("a")
	fmt.Fprintf(ctx.Response, "Request Method:%s, paramvalue:%s", ctx.Request.Method, userid)
}

func main() {
	v := gofw.NewGoFw()
	v.AddRoute("GET", "/a", test)

	log.Fatal(v.Listen(":8080"))
}