# getnovel

顾名思义~

[![Appveyor Build status](https://ci.appveyor.com/api/projects/status/jbcqh6a7jgmonaf1?svg=true)](https://ci.appveyor.com/project/missdeer/getnovel-6xgd2)
[![Travis Build Status](https://secure.travis-ci.org/missdeer/getnovel.png)](https://travis-ci.org/missdeer/getnovel) 
[![GitHub Releases](https://img.shields.io/github/release/missdeer/getnovel.svg?maxAge=2592000)](https://github.com/missdeer/getnovel/releases) 
[![Github All Releases Download Count](https://img.shields.io/github/downloads/missdeer/getnovel/total.svg)](https://github.com/missdeer/getnovel/releases) 
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/missdeer/getnovel/master/LICENSE)

## Build

```bash
go install github.com/missdeer/getnovel@latest
```

## Download prebuilt binaries

所有预编译可以执行文件由[Appveyor 服务](https://ci.appveyor.com/project/missdeer/getnovel-6xgd2)编译生成

## Usage

1. 首先，在当前目录创建名为fonts的目录，并将你喜欢的字体文件改名为CustomFont.ttf后放入该目录；
2. 然后，在当前目录执行命令获取小说内容：`getnovel 小说目录网址`，如`getnovel https://www.aszw.org/book/192/192150/`，可以通过命令行参数对程序行为进行设置，比如指定输出文件格式为PDF或epub等等；
3. 最后，如果需要得到mobi文件，则继续执行命令打包成mobi文件：`kindlegen -c2 -o xxxx.mobi content.opf`, [kindlegen工具](https://www.amazon.com/gp/feature.html?docId=1000765211)可在Amazon官网[下载](https://www.amazon.com/gp/feature.html?docId=1000765211)。可以将`kindlegen`的路径设置到环境变量`KINDLEGEN_PATH`中，或者将`kindlegen`所在目录路径添加到环境变量`PATH`中，则`getnovel`会自动调用`kindlegen`生成mobi文件，省去手动输入命令的步骤。

#### 常用用法

* 输出适合在电脑上看的PDF格式：`getnovel -f pdf -c pc https://www.aszw.org/book/192/192150/`
* 输出适合在电脑上看的PDF格式，但只要其中第11章~第20章内容：`getnovel -f pdf -c pc --fromChapter=11 --toChapter=20 https://www.aszw.org/book/192/192150/`
* 输出适合在电脑上看的PDF格式，但以每100章为一个文件：`getnovel -f pdf -c pc --chaptersPerFile=100 https://www.aszw.org/book/192/192150/`
* 输出适合在Kindle DXG上看的PDF格式：`getnovel -f pdf -c dxg https://www.aszw.org/book/192/192150/`
* 输出适合在6寸或7寸Kindle上看的mobi格式：`getnovel -f mobi https://www.aszw.org/book/192/192150/`，之后需要运行`kindlegen`工具，参考上面第3步

## Supported websites

* 飘天: https://www.ptwxz.com
* UU看书: https://www.uukanshu.com

## Notice

* 输出为PDF格式时，如果遇到打开PDF文件为空白，原因可能是所使用的自定义字体文件中未包含某些字符却被使用了，可以尝试更换嵌入字体文件为字符集较大的，比如“方正准雅宋GBK”等。

## To-Do

- [ ] 使用Lua配置小说网站支持
- [ ] 支持[阅读](https://github.com/gedoor/legado)最新的书源格式

## Number of historical stars
![](https://starchart.cc/missdeer/getnovel.svg)