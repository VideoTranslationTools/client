# 视频机翻客户端

> 还在更新使用教程和功能...

## 前言

本项目的初衷是为了更好的看国外没有字幕的视频（可能是 Youtube or 一些纪录片），那么就需要能够做一个机翻的工具来生成字幕。

使用者需要有一个预期，就是这个项目的翻译质量显然是无法跟专业的字幕人员去对比的。同时生成的字幕断句合理性也有极大的提升空间。

目前是强制生成**中文**的**双语**字幕（也许后期会开放其他语言的翻译支持）。

## 如何使用？

### client_youtube

举个例子：（这里假定你的系统是 Windows，请举一反三）

现在有一个 Youtube 的视频：https://www.youtube.com/watch?v=lhYGXYeMq_E

想要获取对应的中文字幕。那么可以使用本项目提供的 [client_youtube](https://github.com/VideoTranslationTools/client/tree/master/cmd/client_youtube)

1. 先下载这个程序，放到 `C:\WorkSpace\videio_translation`

2. 需要下载 Youtube 的视频下载程序：[yt-dlp](https://github.com/yt-dlp/yt-dlp/releases)，将这个`yu-dlp.exe`放到`C:\WorkSpace\videio_translation`中

3. 申请 ApiKey（待添加教程）

4. 在这个目录下新建一个 `etc` 目录

5. 在 `etc` 目录下新建一个 `client_youtube.yaml` 文件，这个是该程序的配置文件

   ```
   ServerBaseUrl: "xxx"
   ApiKey: "123456"
   ProxyType: "http"
   ProxyUrl: "http://192.168.50.252:20171"
   CacheRootFolder: ".\\cache"
   YTdlpFilePath: "C:\\WorkSpace\\youtube\\yt-dlp.exe"
   ```

> ServerBaseUrl： 机翻服务器的 API 地址
>
> ApiKey：用户的密钥（每天有使用次数限制）
>
> 如果你所在的地区访问 Youtube 有限制，那么你需要使用代理，支持 http or socks5 的代理（不能有登录验证的限制），那么你就需要填写 `ProxyType` 和 `ProxyUrl` 参数，请举一反三的使用
>
> CacheRootFolder：这个是缓存的目录，也是下载 Youtube 视频的缓存目录，同时也是字幕最终的存放目录。会根据下载的视频标题来新建目录存储。
>
> YTdlpFilePath：这个是 Youtube 下载器的存放位置。

基本的使用的参数为：

```shell
client_youtube.exe -yt_url "https://www.youtube.com/watch?v=lhYGXYeMq_E"
```

那么将在以下目录生成对应的下载视频的目录

![image-20230615174406755](assets/image-20230615174406755.png)
