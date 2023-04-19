package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
)

func main() {
	var bindPort = 0
	fmt.Println("try get bind port from env")
	{
		var bindPortStr = os.Getenv("BIND_PORT")
		fmt.Println("bind port str:", bindPortStr)
		bpl, err := strconv.ParseInt(bindPortStr, 10, 64)
		if err != nil {
			panic(err)
		}
		bindPort = int(bpl)
	}
	//fmt.Print("enter bind port:")
	//_, err := fmt.Scanf("%d\n", &bindPort)
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println("bind port:", bindPort)
	var engine = gin.Default()
	engine.NoRoute(func(c *gin.Context) {
		fmt.Println(c.Request)
		identJsonData, err := json.MarshalIndent(c.Request.Header, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(identJsonData))
		var certHeaderValue = c.Request.Header.Get("ssl-client-cert")
		fmt.Println("cert header value:", certHeaderValue)
	})
	err := engine.Run(fmt.Sprintf(":%d", bindPort))
	if err != nil {
		panic(err)
	}
}
