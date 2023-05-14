package config_tv

type TypeAndValuePlugin interface {
	ReadRawData(tv TypeAndValue) []byte
}
