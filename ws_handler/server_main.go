package ws_handler

import (
	"context"
	tls2 "crypto/tls"
	"fmt"
	tls_sidecar "github.com/AkaAny/tls-sidecar"
	"github.com/AkaAny/tls-sidecar/trust_center"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

type WebSocketSidecarServer struct {
	ServerTLSConfigFactory *tls_sidecar.ServerTLSConfigFactory
	ClientTLSConfigFactory *tls_sidecar.ClientTLSConfigFactory
	DeployHostMap          map[string]string
	BackendServiceHost     string
}

func (h *WebSocketSidecarServer) InitRequestRouter(routes gin.IRoutes) {
	var corsMiddleware = func(c *gin.Context) { c.Next() }
	routes.GET("/tlsRequest", corsMiddleware, func(c *gin.Context) {
		w, r := c.Writer, c.Request
		websocketConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns:  []string{"*"},
			CompressionMode: websocket.CompressionDisabled,
		})
		if err != nil {
			panic(err)
		}
		ctx, cancelFn := context.WithTimeout(r.Context(), 1*time.Minute)
		defer cancelFn()
		var netConn = websocket.NetConn(ctx, websocketConn, websocket.MessageBinary)

		var tlsConfig = h.ServerTLSConfigFactory.NewServerConfig()
		fmt.Println(tlsConfig)
		var tlsConn = tls2.Server(netConn, tlsConfig)
		defer tlsConn.Close()
		tls_sidecar.DoTransfer(tlsConn, h.BackendServiceHost) //handle one at a time
	})
}

func (h *WebSocketSidecarServer) InitWrapRouter(wrapEngine gin.IRoutes) {
	var corsMiddleware = func(c *gin.Context) { c.Next() }
	wrapEngine.POST("/wrap", corsMiddleware, func(c *gin.Context) {
		var targetMethod = c.GetHeader("X-Target-Method")
		var targetPath = c.GetHeader("X-Target-Path")
		var targetQuery = c.GetHeader("X-Target-Query")
		var targetDeployID = c.GetHeader("X-Target-Deploy")
		var targetServiceID = c.GetHeader("X-Target-Service")
		deployHost, ok := h.DeployHostMap[targetDeployID]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"msg": errors.Errorf("deploy id:%s does not exist", targetDeployID).Error(),
			})
			return
		}
		var serverBaseUrl = fmt.Sprintf("http://%s/%s", deployHost, targetServiceID)
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

		var tlsConfig = h.ClientTLSConfigFactory.NewClientConfig()
		resp, err := NewWSClient(serverBaseUrl, tlsConfig, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": errors.Wrap(err, "do tls request").Error(),
			})
			return
		}
		for headerKey, headerValues := range resp.Header {
			c.Writer.Header()[headerKey] = headerValues
		}
		defer resp.Body.Close()
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
