package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
)

func main() {
	var bindPort = 0
	flag.IntVar(&bindPort, "bindPort", 0, "bind port")
	flag.Parse()
	if bindPort == 0 {
		fmt.Println("try get bind port from env")
		var bindPortStr = os.Getenv("BIND_PORT")
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
	})
	err := engine.Run(fmt.Sprintf(":%d", bindPort))
	if err != nil {
		panic(err)
	}
}
