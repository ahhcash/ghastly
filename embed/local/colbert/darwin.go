//go:build darwin

package colbert

func getPlatformOnnxPath() string {
	return "/opt/homebrew/lib/libonnxruntime.1.20.1.dylib"
}
