package util

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// URLParamsPair 地址替换对
type URLParamsPair struct {
	Key   string
	Value interface{}
}

// ParamsReplacer URL参数替换器
func ParamsReplacer(url string, pairs []URLParamsPair) string {
	for _, pair := range pairs {
		url = strings.Replace(url, pair.Key, fmt.Sprint(pair.Value), -1)
	}
	return url
}

// RandStringRunes 返回随机字符串
func RandStringRunes(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// InArray 判断是否在数组中
func InArray(needle interface{}, haystack interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(haystack)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(needle, s.Index(i).Interface()) {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

// 在数组中匹配字符串
func MatchInArray(item string, keywords []string) (exists bool, index int) {
	for idx, keyword := range keywords {
		if strings.Contains(item, keyword) {
			return true, idx
		}
	}
	return false, 0
}

// RandNumRunes 返回数字字符串
func RandNumRunes(n int) string {
	var letterRunes = []rune("1234567890")

	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// EmojiFilter 过滤emoji
func EmojiFilter(source string) string {
	var result []rune
	for _, val := range source {
		if _, size := utf8.DecodeLastRuneInString(string(val)); size > 3 {
			continue
		}
		result = append(result, val)
	}
	return string(result)
}

// 检查图片后缀
func ImageCheck(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".png", ".jpeg":
		return true
	}
	return false
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// HashFileSha1 获取文件sh1
func HashFileSha1(filePath string) (string, error) {
	var returnSHA1String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}
	defer file.Close()
	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}
	hashInBytes := hash.Sum(nil)[:20]
	returnSHA1String = hex.EncodeToString(hashInBytes)
	return returnSHA1String, nil
}

// FileSize 获取文件大小
func FileSize(filePath string) (int64, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// 去掉文件名里的非法字符 \ / : * ? " < > | %
func RmFnameIllegalChars(name string) string {
	illegalChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|", "%"}
	for _, char := range illegalChars {
		name = strings.Replace(name, char, "", -1)
	}
	return name
}

// 自动截断
func StringCut(s string, length int) string {
	runes := []rune(s)
	if len(runes) > length {
		return string(runes[:length])
	}
	return s
}

// +8 时区
func LocalTime() *time.Location {
	return time.FixedZone("UTC-8", +8*60*60)
}

func MD5String(s string) string {
	h := md5.New()
	_, _ = io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Sha256String(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// JoinUInt64Arr uint64 数组拼接 sep 组成字符串返回
func JoinUInt64Arr(list []uint64, sep string) string {
	if len(list) == 0 {
		return ""
	}

	if len(list) == 1 {
		return strconv.FormatUint(list[0], 10)
	}

	result := make([]string, len(list))
	for i, v := range list {
		result[i] = strconv.FormatUint(v, 10)
	}

	return strings.Join(result, sep)
}

// Duration 计算延迟帮助函数，单位ms
func Duration(start time.Time) int64 {
	return int64(time.Since((start)) / time.Millisecond)
}
