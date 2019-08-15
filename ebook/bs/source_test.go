package bs

import (
	"fmt"
	"testing"
)

func TestReadBookSourceFromLocalFileSystem(t *testing.T) {

}

func TestReadBookSourceFromURL(t *testing.T) {
	urls := []string{
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_176",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_176_1",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_1909tv",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_hy",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_miui",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_qidian",
		"https://gitee.com/gekunfei/web/raw/master/myBookshelf/bookSource_tingfree",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/00ksw.net",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/kygso",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/530p",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/9txs",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/biquge5200",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/dmzj",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/jpxs123",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/midu",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/motie",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/quanben5",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/shushu8",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/wenku8",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/xiaobiquge",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/xiaoshuo2016",
		"https://gitee.com/alanskycn/yuedu/raw/master/booksource/yomou.syosetu.com",
		"https://blackholep.github.io/20190815set1",
		"https://blackholep.github.io/31xsw",
		"https://blackholep.github.io/37shuwu",
		"https://blackholep.github.io/58xsw",
		"https://blackholep.github.io/abcxs",
		"https://blackholep.github.io/abcxsw",
		"https://blackholep.github.io/ayg",
		"https://blackholep.github.io/bjzww",
		"https://blackholep.github.io/bqgb5200",
		"https://blackholep.github.io/bqgbiqubao",
		"https://blackholep.github.io/bqgbiquge",
		"https://blackholep.github.io/bqgbiquwu",
		"https://blackholep.github.io/bqgbqg5",
		"https://blackholep.github.io/bqgibiquge",
		"https://blackholep.github.io/bqgkuxiaoshuo",
		"https://blackholep.github.io/bqgwqge",
		"https://blackholep.github.io/ddxs208xs",
		"https://blackholep.github.io/dyddu1du",
		"https://blackholep.github.io/dydduyidu",
		"https://blackholep.github.io/fqxs",
		"https://blackholep.github.io/gsw",
		"https://blackholep.github.io/hysy",
		"https://blackholep.github.io/mhtxsw",
		"https://blackholep.github.io/psw",
		"https://blackholep.github.io/shlwxw",
		"https://blackholep.github.io/slk",
		"https://blackholep.github.io/uxs",
		"https://blackholep.github.io/wcxsw",
		"https://blackholep.github.io/wlzww",
		"https://blackholep.github.io/wxm",
		"https://blackholep.github.io/xbqgxbaquge",
		"https://blackholep.github.io/xbqgxbiquge6",
		"https://blackholep.github.io/xbyzww",
		"https://blackholep.github.io/xsz",
		"https://blackholep.github.io/xszww",
		"https://blackholep.github.io/ybzw",
		"https://blackholep.github.io/ylgxs",
		"https://blackholep.github.io/ymx",
		"https://blackholep.github.io/yssm",
		"https://blackholep.github.io/ywxs",
		"https://blackholep.github.io/zsw",
		"https://booksources.github.io/",
		"https://booksources.github.io/list/biqudao_com.json",
		"https://booksources.github.io/list/cn3k5_com.json",
		"https://booksources.github.io/list/gzmeal_com.json",
		"https://booksources.github.io/list/novel101_com.json",
		"https://booksources.github.io/list/qinxiaoshuo.com.json",
		"https://booksources.github.io/list/qxs.la.json",
		"https://booksources.github.io/list/x23qb_com.json",
		"https://booksources.github.io/list/x23us.com.json",
		"https://booksources.github.io/list/xiashutxt_com.json",
		"https://booksources.github.io/list/xslou_com.json",
		"https://booksources.github.io/list/zhaishuyuan_com.json",
		"http://cloud.iszoc.com/booksource/booksource.json",
		//"http://cloud.iszoc.com/booksource/booksources.json",
	}

	for _, u := range urls {
		bs := ReadBookSourceFromURL(u)
		for _, b := range bs {
			fmt.Println(b.BookSourceGroup, b.BookSourceName, b.BookSourceURL, b.Enable)
		}
	}
}
