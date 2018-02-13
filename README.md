# getnovel

顾名思义~

[![Build Status](https://secure.travis-ci.org/dfordsoft/getnovel.png)](https://travis-ci.org/dfordsoft/getnovel) [![GitHub release](https://img.shields.io/github/release/dfordsoft/getnovel.svg?maxAge=2592000)](https://github.com/dfordsoft/getnovel/releases) [![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/dfordsoft/getnovel/master/LICENSE)

## 编译

```bash
go get github.com/dfordsoft/getnovel
```

## 使用方法

1. 首先，在当前目录创建名为fonts的目录，并将你喜欢的字体文件改名为CustomFont.ttf后放入该目录；
2. 然后，在当前目录执行命令获取小说内容：`getnovel 小说目录网址`，如`getnovel https://www.aszw.org/book/192/192150/`，可以通过命令行参数对程序行为进行设置，比如指定输出文件格式为PDF或epub等等；
3. 最后，如果需要得到mobi文件，则继续执行命令打包成mobi文件：`kindlegen -c2 -o xxxx.mobi content.opf`, [kindlegen工具](https://www.amazon.com/gp/feature.html?docId=1000765211)可在Amazon官网[下载](https://www.amazon.com/gp/feature.html?docId=1000765211)。可以将`kindlegen`的路径设置到环境变量`KINDLEGEN_PATH`中，或者将`kindlegen`所在目录路径添加到环境变量`PATH`中，则`getnovel`会自动调用`kindlegen`生成mobi文件，省去手动输入命令的步骤。

#### 常用用法

* 输出适合在电脑上看的PDF格式：`getnovel -f pdf -p pc https://www.aszw.org/book/192/192150/`
* 输出适合在电脑上看的PDF格式，但只要其中第11章~第20章内容：`getnovel -f pdf -p pc --fromChapter=11 --toChapter=20 https://www.aszw.org/book/192/192150/`
* 输出适合在电脑上看的PDF格式，但以每100章为一个文件：`getnovel -f pdf -p pc --chaptersPerFile=100 https://www.aszw.org/book/192/192150/`
* 输出适合在Kindle DXG上看的PDF格式：`getnovel -f pdf -p dxg https://www.aszw.org/book/192/192150/`
* 输出适合在6寸或7寸Kindle上看的mobi格式：`getnovel -f mobi https://www.aszw.org/book/192/192150/`，之后需要运行`kindlegen`工具，参考上面第3步

## 支持小说网站

* 一流吧: http://www.168xs.com
* 爱上中文: https://www.aszw.org
* 少年文学网: https://www.snwx8.com
* 大海中文: https://www.dhzw.org
* 手牵手: https://www.sqsxs.com
* 燃文小说: http://www.ranwena.com
* 笔趣阁系列: http://www.biqugezw.com, http://www.630zw.com, http://www.biquge.lu, http://www.biquge5200.com, http://www.biqudu.com, http://www.biquge.cm, http://www.qu.la, http://www.xxbiquge.com, https://www.biqugev.com
* 落霞: http://www.luoxia.com
* 飘天: http://www.piaotian.com
* 书迷楼: http://www.shumil.com
* UU看书: http://www.uukanshu.net
* 无图小说: http://www.wutuxs.com
* 幼狮书盟: http://www.yssm.org
* 新顶点笔趣阁小说网: http://www.szzyue.com
