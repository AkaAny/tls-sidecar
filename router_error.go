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
	ErrCodeTargetDeployIDNotFound  = 40400
	ErrCodeTargetServiceIDNotFound = 40401
	ErrCodeAPIPathNotFound         = 40402
	ErrCodeAPIMethodNotFound       = 40403
	//403
	ErrCodeUnsupportedIdentityType = 40300
)
