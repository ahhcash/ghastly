package colbert

const (
	modelHfPath = "colbert-ir/colbertv2.0"
)

type platformConfig interface {
	OnnxPath() string
}

var getConfig func() platformConfig
