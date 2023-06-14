package settings

type Configs struct {
	ApiKey          string // 通信密钥
	ProxyType       string `json:",options=[no,http,socks5]"`       // 代理类型
	ProxyUrl        string `json:",default=http://127.0.0.1:10809"` // 代理地址
	CacheRootFolder string `json:",default=."`                      // 缓存的根目录
	YTdlpFilePath   string `json:",default=."`                      // youtube-dl.exe 的绝对路径
}
