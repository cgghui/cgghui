package cgghui

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomString 随机字符串 数字+大小字母
func RandomString(lenX int, strL ...string) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if len(strL) == 1 {
		str = strL[0]
	}
	size := len(str)
	byteArr := []byte(str)
	var result []byte
	for i := 0; i < lenX; i++ {
		result = append(result, byteArr[rand.Intn(size)])
	}
	return string(result)
}

// RandomSliceString 从切片中随机返回
func RandomSliceString(arr *[]string) string {
	return (*arr)[rand.Intn(len(*arr))]
}

// MD5Byte MD5值
func MD5Byte(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

// MD5 MD5值
func MD5(str string) string {
	return MD5Byte([]byte(str))
}

// Str2Int string转int
func Str2Int(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		i = def
	}
	return i
}

// LoadFileLine 加载文件 按行读取
func LoadFileLine(filePath string, call func([]byte) bool) error {
	lines, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	for _, line := range bytes.Split(lines, []byte{10}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if !call(line) {
			break
		}
	}
	return nil
}

// LoadFileLineNo 加载文件 按行读取 (带行号)
func LoadFileLineNo(filePath string, call func(int, []byte) bool) error {
	lines, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	for no, line := range bytes.Split(lines, []byte{10}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if !call(no, line) {
			break
		}
	}
	return nil
}
