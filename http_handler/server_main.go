package tls_on_http

import (
	"bufio"
	"crypto/tls"
	"fmt"
	tls_sidecar "github.com/AkaAny/tls-sidecar"
	"github.com/AkaAny/tls-sidecar/trust_center"
	"github.com/gin-gonic/gin"
	"github.com/jordwest/mock-conn"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type clientHandle struct {
	connectionID   int64
	underlyingConn *mock_conn.Conn
	protocolConn   net.Conn
	writeBuf       *bufio.ReadWriter
	rwLock         *sync.RWMutex
}

const clientHandleKey = "ClientHandle"

func getClientHandle(c *gin.Context) *clientHandle {
	var handle = c.MustGet("ClientHandle").(*clientHandle)
	return handle
}

type HttpSidecarServer struct {
	ServerTLSConfigFactory *tls_sidecar.ServerTLSConfigFactory
	ClientTLSConfigFactory *tls_sidecar.ClientTLSConfigFactory
	BackendServiceHost     string
	DeployIDHostMap        map[string]string
}

func NewConnectionIDFromHeaderMiddleware(clientMap map[int64]*clientHandle, newConnLock *sync.RWMutex) gin.HandlerFunc {
	return func(c *gin.Context) {
		var connectionIDStr = c.GetHeader("X-Connection-ID")
		connectionID, err := strconv.ParseInt(connectionIDStr, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, &ReadResponse{
				N:   0,
				Err: errors.Errorf("invalid connection id:%s", connectionIDStr).Error(),
			})
			return
		}
		newConnLock.RLock()
		handle, ok := clientMap[connectionID]
		newConnLock.RUnlock()
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, &ReadResponse{
				N:   0,
				Err: errors.Errorf("connection id:%d does not exist", connectionID).Error(),
			})
			return
		}
		c.Set(clientHandleKey, handle)
		c.Next()
	}
}

func (h *HttpSidecarServer) InitRequestRouter(engine gin.IRoutes) {
	var clientMap = make(map[int64]*clientHandle)
	var newConnLock = &sync.RWMutex{}
	var connectionIDMiddleware = NewConnectionIDFromHeaderMiddleware(clientMap, newConnLock)
	var tlsConfig = h.ServerTLSConfigFactory.NewServerConfig()
	var corsMiddleware = func(c *gin.Context) { c.Next() }
	engine.POST("/newRequest", corsMiddleware, func(c *gin.Context) {
		newConnLock.Lock()
		var connectionID = int64(len(clientMap))
		newConnLock.Unlock()
		fmt.Println("new connection id:", connectionID)
		var underlyingConn = mock_conn.NewConn()
		//var underlyingConn = NewMockServerConn(connectionID)
		//if err := underlyingConn.Init(); err != nil {
		//	return
		//}
		var tlsConn = tls.Server(underlyingConn.Server, tlsConfig)
		var handle = &clientHandle{
			connectionID:   connectionID,
			underlyingConn: underlyingConn,
			protocolConn:   tlsConn,
			rwLock:         &sync.RWMutex{},
		}
		clientMap[connectionID] = handle
		go func() {
			tls_sidecar.DoTransfer(tlsConn, h.BackendServiceHost)
			fmt.Println("after transfer request")
			//tlsConn.Close()
		}()
		c.JSON(http.StatusOK, &NewRequestResponse{
			ConnectionID: connectionID,
			Err:          "",
		})
	})
	engine.PUT("/clientWrite", corsMiddleware, connectionIDMiddleware, func(c *gin.Context) {
		var handle = getClientHandle(c)
		handle.rwLock.Lock()
		defer handle.rwLock.Unlock()
		//先读取请求body，然后写入raw conn里
		var transactRequest = new(TransactRequest)
		if err := c.BindJSON(transactRequest); err != nil {
			c.JSON(http.StatusBadRequest, &TransactResponse{
				N:   0,
				Err: errors.Wrap(err, "bind json body err").Error(),
			})
			return
		}
		var writeSize = len(transactRequest.Data)
		fmt.Println("connection:", handle.connectionID, " before write size:", writeSize)
		go func() {
			actualWriteLength, err := handle.underlyingConn.Client.Write(transactRequest.Data)
			if err != nil {
				fmt.Println("write to mock client err:", err)
			}
			fmt.Println("connection:", handle.connectionID, " after write size(actual):", actualWriteLength)
		}()
		//if err != nil {
		//	c.JSON(http.StatusInternalServerError, &TransactResponse{
		//		N:   actualWriteLength,
		//		Err: errors.Wrap(err, "write pending data").Error(),
		//	})
		//	return
		//}
		c.JSON(http.StatusOK, &TransactResponse{
			N:   writeSize, //从客户端处实际读取的len
			Err: "",
		})
	})
	engine.PUT("/clientRead", corsMiddleware, connectionIDMiddleware, func(c *gin.Context) {
		var handle = getClientHandle(c)
		handle.rwLock.RLock()
		defer handle.rwLock.RUnlock()
		var requestBody = new(ReadRequest)
		if err := c.BindJSON(requestBody); err != nil {
			c.JSON(http.StatusBadRequest, &ReadResponse{
				N:   0,
				Err: errors.Wrap(err, "bind request body").Error(),
			})
			return
		}
		//将protocol conn写入到raw conn的东西写入到响应body里
		fmt.Println("before read size:", requestBody.Size, " for connection:", handle.connectionID)
		var data = make([]byte, requestBody.Size)

		n, err := handle.underlyingConn.Client.Read(data) //客户端读
		if err != nil {
			c.JSON(http.StatusInternalServerError, &TransactResponse{
				N:   0,
				Err: errors.Wrap(err, "read pending data").Error(),
			})
			return
		}
		c.JSON(http.StatusOK, &ReadResponse{
			Data: data[:n],
			N:    n,
			Err:  "",
		})
	})
	engine.PUT("/clientClose", corsMiddleware, connectionIDMiddleware, func(c *gin.Context) {
		var handle = getClientHandle(c)
		handle.rwLock.Lock()
		defer handle.rwLock.Unlock()
		fmt.Println("before close connection:", handle.connectionID)
		//go func() {
		//
		//	if err != nil {
		//		fmt.Println("connection:", handle.connectionID, " close protocol conn err:", err)
		//	}
		//}()
		go func() {
			io.ReadAll(handle.underlyingConn.Client)
			fmt.Println("after read all server sent data")
		}()
		err := handle.protocolConn.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, &CloseResponse{
				Err: errors.Wrap(err, "close protocol conn").Error(),
			})
			return
		}
		newConnLock.Lock()
		delete(clientMap, handle.connectionID)
		newConnLock.Unlock()
		c.JSON(http.StatusOK, &CloseResponse{
			Err: "",
		})
	})
}

func (h *HttpSidecarServer) InitWrapRouter(wrapEngine gin.IRoutes) {
	var tlsConfig = h.ClientTLSConfigFactory.NewClientConfig()
	var corsMiddleware = func(c *gin.Context) { c.Next() }
	wrapEngine.POST("/wrap", corsMiddleware, func(c *gin.Context) {
		var targetMethod = c.GetHeader("X-Target-Method")
		var targetPath = c.GetHeader("X-Target-Path")
		var targetQuery = c.GetHeader("X-Target-Query")
		var targetDeployID = c.GetHeader("X-Target-Deploy")
		var targetServiceID = c.GetHeader("X-Target-Service")
		deployHost, ok := h.DeployIDHostMap[targetDeployID]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"msg": errors.Errorf("deploy id:%s does not exist", targetDeployID).Error(),
			})
			return
		}
		var serverBaseUrl = fmt.Sprintf("http://%s/%s/http", deployHost, targetServiceID)
		var realUrl = fmt.Sprintf("http://%s%s?%s", targetServiceID, targetPath, targetQuery)
		req, err := http.NewRequest(targetMethod, realUrl, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": errors.Wrap(err, "new http request").Error(),
			})
			return
		}
		req.Header = c.Request.Header
		req.Header.Set(trust_center.IdentityTypeHeaderKey, trust_center.IdentityTypeCertRPC)

		resp, err := NewHttpClient(serverBaseUrl, tlsConfig, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": errors.Wrap(err, "do tls request").Error(),
			})
			return
		}
		for headerKey, headerValues := range resp.Header {
			c.Writer.Header()[headerKey] = headerValues
		}
		respBodyData, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": errors.Wrap(err, "read response body"),
			})
			return
		}
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBodyData)
		//if _, err := c.Writer.Write(respBodyData); err != nil {
		//	c.JSON(http.StatusInternalServerError, gin.H{
		//		"msg": errors.Wrap(err, "write response body"),
		//	})
		//	return
		//}
	})
}
