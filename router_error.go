package tls_sidecar

type RouterErrorCode struct {
	Code int
}

func (x RouterErrorCode) GetCode() int {
	return x.Code
}

const (
	ErrCodeInvalidPeerCertificate       = 40000
	ErrCodeMissingTargetDeployIDHeader  = 40001
	ErrCodeMissingTargetServiceIDHeader = 40002
	ErrCodeFailedToParseJWTToken        = 40003
	ErrCodeInvalidJWTToken              = 40004
	ErrCodeInvalidJWTClaims             = 40005
	//404
	ErrCodeTargetServiceIDNotFound = 40400
	ErrCodeAPIPathNotFound         = 40401
	//403
	ErrCodeUnsupportedIdentityType = 40300
)
