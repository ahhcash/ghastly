package colbert

type osConfig struct {
	onnxPath string
}

func init() {
	conf := &osConfig{
		onnxPath: getPlatformOnnxPath(),
	}

	getConfig = func() platformConfig {
		return conf
	}
}

func (osc *osConfig) OnnxPath() string {
	return osc.onnxPath
}
