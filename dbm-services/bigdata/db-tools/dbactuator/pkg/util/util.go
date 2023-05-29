// Package util TODO
package util

import (
	"crypto/md5"
	"dbm-services/common/go-pubpkg/logger"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/golang/glog"
	"github.com/juju/ratelimit"
	"github.com/pkg/errors"
)

// RetryConfig TODO
type RetryConfig struct {
	Times     int           // 重试次数
	DelayTime time.Duration // 每次重试间隔
}

// Retry 重试
// 第 0 次也需要 delay 再运行
func Retry(r RetryConfig, f func() error) (err error) {
	for i := 0; i < r.Times; i++ {
		time.Sleep(r.DelayTime)
		if err = f(); err == nil {
			return nil
		}
		logger.Warn("第%d次重试,函数错误:%s", i, err.Error(), err.Error())
	}
	return
}

// AtWhere TODO
func AtWhere() string {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		fileName, line := runtime.FuncForPC(pc).FileLine(pc)
		result := strings.Index(fileName, "/bk-dbactuator/")
		if result > 1 {
			preStr := fileName[0:result]
			fileName = strings.Replace(fileName, preStr, "", 1)
		}
		return fmt.Sprintf("%s:%d", fileName, line)
	} else {
		return "Method not Found!"
	}
}

// HasElem TODO
func HasElem(elem interface{}, slice interface{}) bool {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("HasElem error %s ", err)
		}
	}()
	arrV := reflect.ValueOf(slice)
	if arrV.Kind() == reflect.Slice || arrV.Kind() == reflect.Array {
		for i := 0; i < arrV.Len(); i++ {
			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if reflect.DeepEqual(arrV.Index(i).Interface(), elem) {
				return true
			}
		}
	}
	return false
}

const (
	tcpDialTimeout = 3 * time.Second
)

// HostCheck TODO
func HostCheck(host string) bool {
	_, err := net.DialTimeout("tcp", host, time.Duration(tcpDialTimeout))
	if err != nil {
		glog.Infof(err.Error())
		return false
	}
	return true
}

// GetFileMd5 TODO
func GetFileMd5(fileAbPath string) (md5sum string, err error) {
	rFile, err := os.Open(fileAbPath)
	if err != nil {
		return "", err
	}
	defer rFile.Close()
	h := md5.New()
	if _, err := io.Copy(h, rFile); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Struct2Map TODO
func Struct2Map(s interface{}, tag string) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("only accept struct or pointer, got %T", v)
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		if tagValue := f.Tag.Get(tag); tagValue != "" {
			out[tagValue] = v.Field(i).Interface()
		}
	}
	return out, nil
}

// SetField TODO
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}

// Convert2Map TODO
func Convert2Map(m interface{}) map[string]string {
	ret := make(map[string]string)
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var fd string
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Struct:
			fallthrough
		case reflect.Ptr:
			Convert2Map(f.Interface())
		default:
			fd = f.String()
		}
		ret[v.Type().Field(i).Tag.Get("json")] = fd
	}
	return ret
}

// StrIsEmpty TODO
func StrIsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// FileExists 检查目录是否已经存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// IsDirectory 检查路径是否是目录
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// FileExistsErr 如果文件不存在则抛出 error
func FileExistsErr(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		err = errors.WithMessage(err, path)
	}
	return nil
}

// GetFileSize TODO
func GetFileSize(path string) int64 {
	f, err := os.Stat(path)
	if err != nil {
		// 有可能没权限，有可能不存在
		if os.IsNotExist(err) {
			return -1
		} else if os.IsPermission(err) {
			return -2
		} else {
			return -3
		}
	}
	return f.Size()
}

// OutputPrettyJson 直接传一个空结构体过来
func OutputPrettyJson(p interface{}) {
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(p)
	json.Unmarshal(inrec, &inInterface)
	// Make a custom formatter with indent set
	f := colorjson.NewFormatter()
	f.Indent = 4
	pp, err := f.Marshal(inInterface)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Payload Example: ")
	fmt.Println("")
	fmt.Println(string(pp))
	fmt.Println("")
}

// IntSlice2String 效果：[]int{1,2,3,4} -> "1,2,3,4"
func IntSlice2String(elements []int, sep string) string {
	elemStr := ""
	if len(elements) > 0 {
		for i, elem := range elements {
			if i == (len(elements) - 1) {
				elemStr += fmt.Sprintf("%d", elem)
				break
			}
			elemStr += fmt.Sprintf("%d%s", elem, sep)
		}
	}
	return elemStr
}

// ConverMapInterface2MapString TODO
func ConverMapInterface2MapString(mi map[string]interface{}) (ms map[string]string, err error) {
	ms = make(map[string]string)
	for key, v := range mi {
		dv, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("key:%s 断言string 失败", key)
		}
		ms[key] = dv
	}
	return
}

// RegexReplaceSubString TODO
func RegexReplaceSubString(str, old, new string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(%s)`, old))
	return re.ReplaceAllString(str, new)
}

// IOLimitRate TODO
// io.Copy 限速
func IOLimitRate(dst io.Writer, src io.Reader, bwlimitMB int64) (written int64, err error) {
	bwlimit := bwlimitMB * 1024 * 1024
	srcBucket := ratelimit.NewBucketWithRate(float64(bwlimit), bwlimit)
	return io.Copy(dst, ratelimit.Reader(src, srcBucket))
}

// GetSuffixWithLenAndSep 获取后缀
// 先截取后面 maxlen 长度字符串，再根据 separator 分隔取后缀
func GetSuffixWithLenAndSep(strList []string, separator string, maxlen int) []string {
	if maxlen > 0 {
		for i, s := range strList {
			l := len(s)
			if l-maxlen > 0 {
				strList[i] = s[l-maxlen:]
			}
		}
	}
	seqList := make([]string, len(strList))
	for i, s := range strList {
		seqList[i] = LastElement(strings.Split(s, separator))
	}
	return seqList
}

// LastElement TODO
func LastElement(arr []string) string {
	return arr[len(arr)-1]
}

// ReverseRead  ·		逆序读取文件，类型tail -n 10
//
//	@receiver name
//	@receiver lineNum 		读取最后多少上内容
//	@return []string	返回逆序读取的文件内容
//	@return error
func ReverseRead(name string, lineNum uint) ([]string, error) {
	// 打开文件
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// 获取文件大小
	fs, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fs.Size()

	var offset int64 = -1   // 偏移量，初始化为-1，若为0则会读到EOF
	char := make([]byte, 1) // 用于读取单个字节
	lineStr := ""           // 存放一行的数据
	buff := make([]string, 0, 100)
	for (-offset) <= fileSize {
		// 通过Seek函数从末尾移动游标然后每次读取一个字节
		file.Seek(offset, io.SeekEnd)
		_, err := file.Read(char)
		if err != nil {
			return buff, err
		}
		if char[0] == '\n' {
			offset--  // windows跳过'\r'
			lineNum-- // 到此读取完一行
			buff = append(buff, lineStr)
			lineStr = ""
			if lineNum == 0 {
				return buff, nil
			}
		} else {
			lineStr = string(char) + lineStr
		}
		offset--
	}
	buff = append(buff, lineStr)
	return buff, nil
}

// SliceErrorsToError TODO
func SliceErrorsToError(errs []error) error {
	var errStrs []string
	for _, e := range errs {
		errStrs = append(errStrs, e.Error())
	}
	errString := strings.Join(errStrs, "\n")
	return errors.New(errString)
}
