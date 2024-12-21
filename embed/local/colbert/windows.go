//go:build windows

package colbert

func getPlatformOnnxPath() string {
	return "C:\\Program Files\\onnxruntime\\bin\\onnxruntime.dll"
}
