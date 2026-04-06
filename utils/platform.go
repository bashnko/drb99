package utils

type PlatformSpec struct {
	InputKey string
	NodeOS   string
	NodeArch string
	GoSuffix string
	Ext      string
}

var supportedPlatformSpec = map[string]PlatformSpec{
	"linux-amd64":   {InputKey: "linux-amd64", NodeOS: "linux", NodeArch: "x64", GoSuffix: "linux-amd64", Ext: ""},
	"linux-arm64":   {InputKey: "linux-arm64", NodeOS: "linux", NodeArch: "arm64", GoSuffix: "linux-arm64", Ext: ""},
	"darwin-amd64":  {InputKey: "darwin-amd64", NodeOS: "darwin", NodeArch: "x64", GoSuffix: "darwin-amd64", Ext: ""},
	"darwin-arm64":  {InputKey: "darwin-arm64", NodeOS: "darwin", NodeArch: "arm64", GoSuffix: "darwin-arm64", Ext: ""},
	"windows-amd64": {InputKey: "windows-amd64", NodeOS: "win32", NodeArch: "win32", GoSuffix: "windows-amd64", Ext: "exe"},
}
