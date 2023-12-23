package tls_on_http

import (
	"fmt"
	"github.com/guonaihong/gout"
	mock_conn "github.com/jordwest/mock-conn"
	"github.com/pkg/errors"
	"net"
	"time"
)

// 这个Conn给到tls.Client
type MockClientConn struct {
	connectionID   int64
	underlyingConn *mock_conn.Conn
	serverBaseUrl  string

	closeChan chan bool
}

func NewMockClientConn(serverBaseUrl string) *MockClientConn {
	var clientConn = &MockClientConn{
		connectionID:   0,
		underlyingConn: mock_conn.NewConn(),
		serverBaseUrl:  serverBaseUrl,
	}
	return clientConn
}

func (m *MockClientConn) Init() error {
	var newRequestUrl = fmt.Sprintf("%s/newRequest", m.serverBaseUrl)
	var newConnectionResponse = new(NewRequestResponse)
	if err := gout.New().POST(newRequestUrl).BindJSON(newConnectionResponse).Do(); err != nil {
		return errors.Wrap(err, "http request err")
	}
	if newConnectionResponse.Err != "" {
		return errors.New(newConnectionResponse.Err)
	}
	m.connectionID = newConnectionResponse.ConnectionID
	fmt.Println("connection id:", m.connectionID)
	//init buffers

	//go m.workingThread()
	return nil
}

func (m *MockClientConn) Read(b []byte) (n int, err error) {
	//var availableSize = m.readBuffer.Reader.Size()
	//fmt.Println("conn id:", m.connectionID, "wants to read size:", len(b), "/", availableSize)
	//m.requiredSizeLocked = len(b)
	var requiredSize = len(b)
	fmt.Println("before connection:", m.connectionID, " read size:", requiredSize)
	data, n, err := m.doReadPendingData(requiredSize)
	if err != nil {
		return n, errors.Wrap(err, "read pending data")
	}
	fmt.Println("after connection:", m.connectionID, " actual read size(remote):", n)
	go func() {
		m.underlyingConn.Server.Write(data[:n])
	}()
	n, err = m.underlyingConn.Client.Read(b)
	fmt.Println("after connection:", m.connectionID, " actual read size(local):", n)
	return n, err
}

func (m *MockClientConn) WriteForRead(data []byte) (n int, err error) {
	return m.underlyingConn.Server.Write(data) //服务器写=客户端读
}

func (m *MockClientConn) Write(b []byte) (n int, err error) {
	var requiredSize = len(b)
	fmt.Println("before connection:", m.connectionID, "write size:", requiredSize)
	var serverTransactUrl = fmt.Sprintf("%s/clientWrite", m.serverBaseUrl)
	var statusCode = int(0)
	var transactResponse = new(TransactResponse)
	if err := gout.New().PUT(serverTransactUrl).SetHeader(map[string]any{
		"X-Connection-ID": m.connectionID,
	}).SetJSON(&TransactRequest{
		Data: b,
	}).Code(&statusCode).BindJSON(transactResponse).Do(); err != nil {
		return 0, err
	}
	n = transactResponse.N
	fmt.Println("after connection:", m.connectionID, "actual write size:", n)
	if transactResponse.Err != "" {
		return n, errors.New(transactResponse.Err)
	}
	return n, nil
}

func (m *MockClientConn) doReadPendingData(readSize int) (data []byte, n int, err error) {
	var url = fmt.Sprintf("%s/clientRead", m.serverBaseUrl)
	var readPendingDataResponse = new(ReadResponse)
	if err := gout.New().PUT(url).SetHeader(map[string]any{
		"X-Connection-ID": m.connectionID,
	}).SetJSON(&ReadRequest{
		Size: readSize,
	}).BindJSON(readPendingDataResponse).Do(); err != nil {
		return nil, 0, errors.Wrap(err, "http request err")
	}
	data = readPendingDataResponse.Data
	n = readPendingDataResponse.N
	if readPendingDataResponse.Err != "" {
		return data, n, errors.New(readPendingDataResponse.Err)
	}
	return data, n, nil
}

func (m *MockClientConn) Close() error {
	var url = fmt.Sprintf("%s/clientClose", m.serverBaseUrl)
	var closeResponse = new(CloseResponse)
	if err := gout.New().PUT(url).SetHeader(map[string]any{
		"X-Connection-ID": m.connectionID,
	}).BindJSON(closeResponse).Do(); err != nil {
		return errors.Wrap(err, "http request err")
	}
	if closeResponse.Err != "" {
		return errors.New(closeResponse.Err)
	}
	//go func() {
	//	io.ReadAll(m.underlyingConn.Server) //读取所有的数据，解除管道阻塞
	//}()
	return m.underlyingConn.Close()
}

func (m *MockClientConn) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (m *MockClientConn) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (m *MockClientConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockClientConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockClientConn) SetWriteDeadline(t time.Time) error {
	return nil
}
