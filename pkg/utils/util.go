// Copyright 2022 The kubegems.io Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	regA_Z  = regexp.MustCompile("[A-Z]+")
	rega_z  = regexp.MustCompile("[a-z]+")
	reg0_9  = regexp.MustCompile("[0-9]+")
	reg_chs = regexp.MustCompile(`[!\.@#$%~]+`)
)

var (
	TimeForever               = time.Date(9893, time.December, 26, 0, 0, 0, 0, time.UTC) // 伟人8000年诞辰
	MaxDuration time.Duration = 1<<63 - 1
)

func StrOrDef(s string, def string) string {
	if s == "" {
		return def
	}
	return s
}

// 保留n位小数
func RoundTo(n float64, decimals uint32) float64 {
	return math.Round(n*math.Pow(10, float64(decimals))) / math.Pow(10, float64(decimals))
}

// DayStartTime 返回当天0点
func DayStartTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func NextDayStartTime(t time.Time) time.Time {
	return DayStartTime(t).Add(24 * time.Hour)
}

func ToUint(id string) uint {
	idint, err := strconv.Atoi(id)
	if err != nil {
		return 0
	}
	return uint(idint)
}

func ValidPassword(input string) error {
	if len(input) < 8 {
		return errors.New("密码长度至少8位,包含大小写字母和数字以及特殊字符(.!@#$%~)")
	}
	if !regA_Z.Match([]byte(input)) {
		return errors.New("密码长度至少8位,包含大小写字母和数字以及特殊字符(.!@#$%~)")
	}
	if !rega_z.Match([]byte(input)) {
		return errors.New("密码长度至少8位,包含大小写字母和数字以及特殊字符(.!@#$%~)")
	}
	if !reg0_9.Match([]byte(input)) {
		return errors.New("密码长度至少8位,包含大小写字母和数字以及特殊字符(.!@#$%~)")
	}
	if !reg_chs.Match([]byte(input)) {
		return errors.New("密码长度至少8位,包含大小写字母和数字以及特殊字符(.!@#$%~)")
	}

	return nil
}

func MakePassword(input string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func GeneratePassword() string {
	r := []rune{}
	r = append(r, RandomRune(4, RuneKindLower)...)
	r = append(r, RandomRune(3, RuneKindUpper)...)
	r = append(r, RandomRune(2, RuneKindNum)...)
	r = append(r, RandomRune(1, RuneKindChar)...)
	rand.Shuffle(len(r), func(i, j int) {
		if rand.Intn(10) > 5 {
			r[i], r[j] = r[j], r[i]
		}
	})
	return string(r)
}

func ValidatePassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func JoinFlagName(prefix, key string) string {
	if prefix == "" {
		return strings.ToLower(key)
	}
	return strings.ToLower(prefix + "-" + key)
}

const (
	RuneKindNum   = "num"
	RuneKindLower = "lower"
	RuneKindUpper = "upper"
	RuneKindChar  = "char"
)

var (
	lowerLetterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	upperLetterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numRunes         = []rune("0123456789")
	charRunes        = []rune("!.@#$%~")
)

func RandomRune(n int, kind string) []rune {
	b := make([]rune, n)
	var l []rune
	switch kind {
	case RuneKindChar:
		l = charRunes
	case RuneKindUpper:
		l = upperLetterRunes
	case RuneKindLower:
		l = lowerLetterRunes
	case RuneKindNum:
		l = numRunes
	default:
		l = lowerLetterRunes
	}
	length := len(l)
	for i := range b {
		b[i] = l[rand.Intn(length)]
	}
	return b
}

func BoolToString(a bool) string {
	if a {
		return "1"
	}
	return "0"
}

func BoolToFloat64(a bool) float64 {
	if a {
		return 1
	}
	return 0
}

func TimeZeroToNull(t *time.Time) *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	return t
}

func FormatMysqlDumpTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05.000")
}

func UintToStr(i *uint) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(int(*i))
}

type DesEncryptor struct {
	Key []byte
}

func (e *DesEncryptor) EncryptBase64(input string) (string, error) {
	data := []byte(input)
	block, err := des.NewCipher(e.Key)
	if err != nil {
		return "", err
	}
	data = e.Padding(data, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, e.Key)
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func (e *DesEncryptor) DecryptBase64(input string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	block, err := des.NewCipher(e.Key)
	if err != nil {
		return "", err
	}
	blockMode := cipher.NewCBCDecrypter(block, e.Key)
	origData := make([]byte, len(data))
	blockMode.CryptBlocks(origData, data)
	origData = e.UnPadding(origData)
	return string(origData), nil
}

func (e *DesEncryptor) Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func (e *DesEncryptor) UnPadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return []byte{}
	}
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
	PETABYTE
	EXABYTE
)

// ConvertBytes 保留两位小数
func ConvertBytes(bytes float64) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= EXABYTE:
		unit = "EB"
		value = value / EXABYTE
	case bytes >= PETABYTE:
		unit = "PB"
		value = value / PETABYTE
	case bytes >= TERABYTE:
		unit = "TB"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "GB"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "MB"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "KB"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0B"
	}

	result := strconv.FormatFloat(value, 'f', 2, 64)
	result = strings.TrimSuffix(result, ".00")
	return result + unit
}

func WaitGroupWithTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

// CheckStructFieldsEmpty check all struct fileds not empty.
// Now only check string and int.
func CheckStructFieldsEmpty(obj any) error {
	value := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
		t = t.Elem()
	}

	emptyFieldName := ""
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		switch field.Kind() {
		case reflect.String:
			if field.String() == "" {
				emptyFieldName = t.Field(i).Name
				break
			}
		case reflect.Int:
			if field.Int() == 0 {
				emptyFieldName = t.Field(i).Name
			}
		}
	}
	if emptyFieldName != "" {
		return fmt.Errorf("field %s is tmpty", emptyFieldName)
	}
	return nil
}
