//go:build linux

package colbert

func getPlatformOnnxPath() string {
	return "/usr/local/lib/onnxruntime.so"
}
