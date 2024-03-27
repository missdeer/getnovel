// https://greasyfork.org/zh-CN/scripts/479460-%E4%B8%83%E7%8C%AB%E5%85%A8%E6%96%87%E5%9C%A8%E7%BA%BF%E5%85%8D%E8%B4%B9%E8%AF%BB
// https://github.com/shing-yu/7mao-novel-downloader

package handler

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/missdeer/getnovel/config"
	"github.com/missdeer/golib/httputil"
)

func init() {
	registerNovelSiteHandler(&config.NovelSiteHandler{
		Sites: []config.NovelSite{
			{
				Title: `七猫`,
				Urls:  []string{`https://www.qimao.com/`},
			},
		},
		CanHandle: func(u string) bool {
			patterns := []string{
				`https://www\.qimao\.com/shuku/[0-9\-]+/`,
			}
			for _, pattern := range patterns {
				reg := regexp.MustCompile(pattern)
				if reg.MatchString(u) {
					return true
				}
			}
			return false
		},
		ExtractChapterList:    extractQimaoChapterList,
		ExtractChapterContent: extractQimaoChapterContent,
		PreprocessContentLink: preprocessQimaoChapterLink,
	})
}

func preprocessQimaoChapterLink(u string) (string, http.Header) {
	matchb, _ := url.Parse(u)

	// 从URL中提取id和chapterId
	paths := strings.Split(strings.Trim(matchb.Path, "/"), "/")
	lastPath := paths[len(paths)-1]
	ids := strings.Split(lastPath, "-")

	if len(ids) < 2 {
		fmt.Println("URL does not contain expected ids")
		return u, http.Header{}
	}

	// 构造参数
	params := map[string]string{
		"id":        ids[0],
		"chapterId": ids[1],
	}

	const signKey = "d3dGiJc651gSQ8w1"
	params["sign"] = generateMD5Sign(params, signKey)

	// 构造Headers
	headers := map[string]string{
		"app-version":    "51110",
		"platform":       "android",
		"reg":            "0",
		"AUTHORIZATION":  "",
		"application-id": "com.****.reader",
		"net-env":        "1",
		"channel":        "unknown",
		"qm-params":      "",
	}
	headers["sign"] = generateMD5Sign(headers, signKey)
	// convert headers to http.Header
	header := http.Header{}
	for key, value := range headers {
		header.Set(key, value)
	}
	// 构造最终请求URL
	finalURL := "https://api-ks.wtzw.com/api/v1/chapter/content?" + toParams(params)
	return finalURL, header
}

func extractQimaoChapterList(u string, rawPageContent []byte) (title string, chapters []*config.NovelChapterInfo) {
	reg := regexp.MustCompile(`https://www\.qimao\.com/shuku/([0-9\-]+)/`)
	// extract book id from url
	ss := reg.FindAllStringSubmatch(u, -1)
	s := ss[0]
	if len(s) < 2 {
		return
	}
	bookId := s[1]
	// if bookId is xxxx-yyyy pattern, then split it and use xxxx as bookId
	if strings.Contains(bookId, "-") {
		bookId = strings.Split(bookId, "-")[0]
	}
	// extract chapter list, https://www.qimao.com/api/book/chapter-list?book_id=1710753
	chapterListUrl := "https://www.qimao.com/api/book/chapter-list?book_id=" + bookId
	chapterListResp, err := httputil.GetBytes(chapterListUrl, http.Header{}, 60*time.Second, 3)
	if err != nil {
		log.Println("get chapter list failed", err)
		return
	}
	// unmarshal chapter list as JSON
	var chapterList struct {
		Data struct {
			Chapters []struct {
				Id    string `json:"id"`
				Title string `json:"title"`
				Index string `json:"index"`
			} `json:"chapters"`
		} `json:"data"`
	}
	err = json.Unmarshal(chapterListResp, &chapterList)
	if err != nil {
		log.Println("unmarshal chapter list failed", err)
		return
	}
	for _, chapter := range chapterList.Data.Chapters {
		chapters = append(chapters, &config.NovelChapterInfo{
			Index: len(chapters),
			Title: chapter.Title,
			URL:   "https://www.qimao.com/shuku/" + bookId + "-" + chapter.Id,
		})
	}

	// extract <title> tag from page content as title
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(rawPageContent))
	if err != nil {
		log.Println("parse page content failed", err)
		return
	}
	title = doc.Find("title").Text()
	index := strings.Index(title, `免费阅读`)
	if index > 0 {
		title = title[:index]
	}
	return
}

func extractQimaoChapterContent(u string, rawPageContent []byte) (c []byte) {
	var response QimaoArticleContentResponse
	if err := json.Unmarshal(rawPageContent, &response); err != nil {
		return
	}

	// 提取iv和密文
	txt := response.Data.Content
	iv := txt[:32]
	content := txt[32:]

	// 解密
	decryptedContent, err := decrypt(content, iv)
	if err != nil {
		return
	}

	// 替换换行符
	result := strings.ReplaceAll(decryptedContent, "<br>", "\n")
	return []byte(result)
}

func toParams(params map[string]string) string {
	var parts []string
	for key, value := range params {
		parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}
	return strings.Join(parts, "&")
}

func generateMD5Sign(params map[string]string, signKey string) string {
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var signString string
	for _, key := range keys {
		signString += key + "=" + params[key]
	}
	signString += signKey

	return fmt.Sprintf("%x", md5.Sum([]byte(signString)))
}

// 假设QimaoArticleContentResponse是从API获取的结构体类型
type QimaoArticleContentResponse struct {
	Data struct {
		Content string `json:"content"`
	} `json:"data"`
}

// decrypt 解密函数
func decrypt(data, ivString string) (string, error) {
	key, _ := hex.DecodeString("32343263636238323330643730396531")
	iv, _ := hex.DecodeString(ivString)

	// 假设data是hex编码的，先转换为字节
	cipherText, _ := hex.DecodeString(data)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", err // Cipher text too short
	}

	// CBC模式解密
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	// PKCS#7 unpadding
	unpadSize := int(cipherText[len(cipherText)-1])
	cipherText = cipherText[:len(cipherText)-unpadSize]

	return string(cipherText), nil
}
