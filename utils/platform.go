package utils

import (
	"fmt"
	"maps"
)

type PlatformSpec struct {
	InputKey string
	NodeOS   string
	NodeArch string
	GoOS     string
	GoArch   string
	GoSuffix string
	Ext      string
}

var supportedPlatformSpec = map[string]PlatformSpec{
	"linux-amd64":   {InputKey: "linux-amd64", NodeOS: "linux", NodeArch: "x64", GoOS: "linux", GoArch: "amd64", GoSuffix: "linux_amd64", Ext: ""},
	"linux-arm64":   {InputKey: "linux-arm64", NodeOS: "linux", NodeArch: "arm64", GoOS: "linux", GoArch: "arm64", GoSuffix: "linux_arm64", Ext: ""},
	"darwin-amd64":  {InputKey: "darwin-amd64", NodeOS: "darwin", NodeArch: "x64", GoOS: "darwin", GoArch: "amd64", GoSuffix: "macos_amd64", Ext: ""},
	"darwin-arm64":  {InputKey: "darwin-arm64", NodeOS: "darwin", NodeArch: "arm64", GoOS: "darwin", GoArch: "arm64", GoSuffix: "macos_arm64", Ext: ""},
	"windows-amd64": {InputKey: "windows-amd64", NodeOS: "win32", NodeArch: "x64", GoOS: "windows", GoArch: "amd64", GoSuffix: "windows_amd64", Ext: ".exe"},
}

func SupportedPlatformSpecs() map[string]PlatformSpec {
	copyMap := make(map[string]PlatformSpec, len(supportedPlatformSpec))
	maps.Copy(copyMap, supportedPlatformSpec)
	return copyMap
}

func ResolvePlatformSpec(platform string) (PlatformSpec, error) {
	spec, ok := supportedPlatformSpec[platform]
	if !ok {
		return PlatformSpec{}, fmt.Errorf("unsupported platform: %s", platform)
	}
	return spec, nil
}

func NodeKey(spec PlatformSpec) string {
	return spec.NodeOS + "-" + spec.NodeArch
}

func ReleaseAssetName(binaryName, version string, spec PlatformSpec, archive string) string {
	base := fmt.Sprintf("%s_%s_%s", binaryName, version, spec.GoSuffix)
	switch archive {
	case "zip":
		return base + ".zip"
	case "tar.gz":
		return base + ".tar.gz"
	default:
		return base + spec.Ext
	}
}
