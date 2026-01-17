# getnovel

顾名思义~

[![GitHub Releases](https://img.shields.io/github/release/missdeer/getnovel.svg?maxAge=2592000)](https://github.com/missdeer/getnovel/releases) 
[![Github All Releases Download Count](https://img.shields.io/github/downloads/missdeer/getnovel/total.svg)](https://github.com/missdeer/getnovel/releases) 
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/missdeer/getnovel/master/LICENSE)

## 编译

```bash
git clone https://github.com/missdeer/getnovel.git
cd getnovel
CGO_ENABLED=0 go build -ldflags="-s -w"
```

无需CGO，可在任何平台直接编译。

## 下载

https://github.com/missdeer/getnovel/releases

## 使用方法

1. 首先，在当前目录创建名为`fonts`的目录，并将你喜欢的字体文件改名为`CustomFont.ttf`后放入该目录；
2. 然后，在当前目录执行命令获取内容：`getnovel 目录网址`，如`getnovel https://www.piaotia.com/html/15/15316/`，可以通过命令行参数对程序行为进行设置，比如指定输出文件格式为PDF或epub等等；
3. 最后，如果需要得到mobi文件，则继续执行命令打包成mobi文件：`kindlegen -c2 -o xxxx.mobi content.opf`, `kindlegen`工具可在Amazon官网[下载Kindle Previewer](https://www.amazon.com/Kindle-Previewer/b?ie=UTF8&node=21381691011)安装后获取。可以将`kindlegen`的路径设置到环境变量`KINDLEGEN_PATH`中，或者将`kindlegen`所在目录路径添加到环境变量`PATH`中，则`getnovel`会自动调用`kindlegen`生成mobi文件，省去手动输入命令的步骤。

### 常用用法

* 输出适合在多种设备上看的epub格式（推荐）：`getnovel -f epub https://www.piaotia.com/html/15/15316/`
* 输出适合在电脑上看的PDF格式：`getnovel -f pdf -c pc https://www.piaotia.com/html/15/15316/`
* 输出适合在电脑上看的PDF格式，但只要其中第11章~第20章内容：`getnovel -f pdf -c pc --fromChapter=11 --toChapter=20 https://www.piaotia.com/html/15/15316/`
* 输出适合在电脑上看的PDF格式，但以每100章为一个文件：`getnovel -f pdf -c pc --chaptersPerFile=100 https://www.piaotia.com/html/15/15316/`
* 输出适合在Kindle DXG上看的PDF格式：`getnovel -f pdf -c dxg https://www.piaotia.com/html/15/15316/`
* 输出适合在6寸或7寸Kindle上看的mobi格式：`getnovel -f mobi https://www.piaotia.com/html/15/15316/`，之后需要运行`kindlegen`工具，参考上面第3步

## 内建支持网站

|  ID   | 网站名称                                | 已支持 | 大陆 IP | 非大陆 IP |⚠️ 需要额外注意                                                                          |
|-------|----------------------------------------|-------|--------|----------|-------------------------------------------------------------------------------------|
| 1     | [书海阁小说网](https://www.shuhaige.net/)   | ❌      | ✅     | ✅          | 搜索限流 (Unexpected end of file from server)                                          |
| 2     | [全本小说网](https://quanben5.com/)        | ❌      | ❓     | ✅          | 需要梯子，完本很全，连载基本搜不到，同 quanben5.io, quanben-xiaoshuo.com                              |
| 3     | [69书吧](https://www.69shuba.com/)     | ✅       | ❌     | ✅          | 需要梯子，搜索有CF，章节限流，推荐线程数1                                                             |
| 4     | [大熊猫文学](https://www.dxmwx.org/)       | ❌      | ✅     | ✅          |                                                                                    |
| 5     | [笔趣阁22](https://www.22biqu.com/)      | ❌      | ✅     | ✅          |                                                                                    |
| 6    | [零点小说](https://www.0xs.net/)          | ❌      | ✅     | ✅          | 限流程度和69书吧相似，爬取过快会封IP且获取不到正文                                                        |
| 7    | [得奇小说网](https://www.deqixs.com/)      | ❌      | ❌     | ✅          | 基本只有新书，爬取频率过快会封禁IP (Remote host terminated the handshake)，搜索有限流 (Connection reset) |
| 8    | [小说虎](https://www.xshbook.com/)       | ❌      | ✅     | ✅          | 正文广告较多，需手动过滤                                                                       |
| 9    | [略更网](https://www.luegeng.com/)       | ❌      | ✅     | ✅        |                                                                                    |
| 10    | [书林文学](http://www.shu009.com/)       | ❌       | ✅     | ✅    | 源站目录有重复、缺章的情况，章节限流                                                                 |
| 11    | [速读谷](https://www.sudugu.com/)        | ❌      | ❌     | ✅        | 同得奇小说网                                                                             |
| 12    | [八一中文网](http://www.81zwwww.com/)     | ❌       | ✅     | ✅    |                                                                                    |
| 13    | [阅读库](http://www.yeudusk.com/)        | ❌      | ✅     | ✅        |                                                                                    |
| 14    | [顶点小说](https://www.wxsy.net/)        | ❌       | ✅     | ✅    | 搜索、详情、章节限流                                                                         |
| 15    | [飘天文学网](https://www.piaotia.com)   | ✅       | ✅     | ✅     | 需要梯子                                                                        |

> [!IMPORTANT]
> 使用大陆 IP 为 ❌ 的书源时，国内用户（可能）需要梯子（需要非大陆 IP）
>
> 使用非大陆 IP 为 ❌ 的书源时，国外用户（可能）需要梯子（需要大陆 IP）
>


## 注意

* 输出为PDF格式时，如果遇到打开PDF文件为空白，原因可能是所使用的自定义字体文件中未包含某些字符却被使用了，可以尝试更换嵌入字体文件为字符集较大的，比如“方正准雅宋GBK”等。
* 推荐适合屏幕阅读的简体中文字体：方正准雅宋、方正屏显雅宋、方正莹雪、霞鹜文楷、仓耳今楷 etc.
* 如果需要使用代理，则设置环境变量`HTTP_PROXY/HTTPS_PROXY/SOCKS5_PROXY`，注意`HTTP_PROXY/HTTPS_PROXY`需要带scheme，如`HTTP_PROXY=http://127.0.0.1:7890`，`SOCKS5_PROXY`则不用，如`SOCKS5_PROXY=127.0.0.1:7891`。
* 69看吧等有反爬机制的网站，可以通过轮循代理的方式解决，例如使用clash配置多个代理服务器，加到同一个代理组并设置`type: load-balance, strategy: round-robin`，然后让getnovel走clash的代理端口即可。

## 待办事项

- [x] 支持[阅读](https://github.com/gedoor/legado)书源格式

## 免责声明

此程序旨在用于与网络爬虫和网页处理技术相关的教育和研究目的。不应将其用于任何非法活动或侵犯他人权利的行为。用户对使用此程序引发的任何法律责任和风险负有责任，作者和项目贡献者不对因使用程序而导致的任何损失或损害承担责任。

在使用此程序之前，请确保遵守相关法律法规以及网站的使用政策，并在有任何疑问或担忧时咨询法律顾问。

## 加星历史趋势
![](https://starchart.cc/missdeer/getnovel.svg)
