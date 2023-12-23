package tls_on_http

type NewRequestResponse struct {
	ConnectionID int64  `json:"connectionID"`
	Err          string `json:"err"`
}

type TransactRequest struct {
	Data []byte `json:"data"`
}

type TransactResponse struct {
	N   int    `json:"n"`
	Err string `json:"err"`
}

type ReadRequest struct {
	Size int `json:"size"`
}

type ReadResponse struct {
	Data []byte `json:"data"`
	N    int    `json:"n"`
	Err  string `json:"err"`
}

type CloseResponse struct {
	Err string `json:"err"`
}
