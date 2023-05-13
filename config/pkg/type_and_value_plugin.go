package pkg

type TypeAndValuePlugin interface {
	ReadRawData(tv TypeAndValue) []byte
}
