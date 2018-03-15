// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// utils.go is a collection of generic functions that are not specific to FIND.

package glb

import (
	"bytes"
	"compress/flate"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
	"log"
	"sort"
	"runtime"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var (
	// Trace is a logging handler
	Trace *log.Logger
	// Info is a logging handler
	Info *log.Logger
	// Warning is a logging handler
	Warning *log.Logger
	// Debug is a logging handler
	Debug *log.Logger
	// Error is a logging handler
	Error *log.Logger
)

// Init function for generating the logging handlers
func Init(
	traceHandle io.Writer,
	infoHandle io.Writer,
	debugHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE : ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO : ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Debug = log.New(debugHandle,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARN : ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERR  : ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func init() {
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	// Trace.Println("I have something standard to say")
	// Info.Println("Special Information")
	// Warning.Println("There is something you need to know about")
	// Error.Println("Something has failed")
}

// GetLocalIP returns the local ip address
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	bestIP := "localhost"
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && (strings.Contains(ipnet.IP.String(), "192.168.1") || strings.Contains(ipnet.IP.String(), "192.168")) {
				return ipnet.IP.String()
			}
		}
	}
	return bestIP
}

// stringInSlice returns boolean of whether a string is in a slice.
func StringInSlice(s string, strings []string) bool {
	for _, k := range strings {
		if s == k {
			return true
		}
	}
	return false
}

// timeTrack can be defered to provide function timing.
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	Debug.Println(name, "took", elapsed)
}

// getMD5Hash returns a md5 hash of string.
func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// average64 computes the average of a float64 slice.
func Average64(vals []float64) float64 {
	sum := float64(0)
	for _, val := range vals {
		sum += float64(val)
	}
	return sum / float64(len(vals))
}

// standardDeviation64 computes the standard deviation of a float64 slice.
func Variance64(vals []float64) float64 {
	meanVal := Average64(vals)

	sum := float64(0)
	for _, val := range vals {
		sum += math.Pow(float64(val)-meanVal, 2)
	}
	sum = sum / (float64(len(vals)) - 1)

	return float64(sum)
}

// standardDeviation64 computes the standard deviation of a float64 slice.
func StandardDeviation64(vals []float64) float64 {
	meanVal := Average64(vals)

	sum := float64(0)
	for _, val := range vals {
		sum += math.Pow(float64(val)-meanVal, 2)
	}
	sum = sum / (float64(len(vals)) - 1)
	sd := math.Sqrt(sum)

	return float64(sd)
}

// standardDeviation comptues the standard deviation of a float32 slice.
func StandardDeviation(vals []float32) float32 {
	sum := float64(0)
	for _, val := range vals {
		sum += float64(val)
	}
	meanVal := sum / float64(len(vals))

	sum = float64(0)
	for _, val := range vals {
		sum += math.Pow(float64(val)-meanVal, 2)
	}
	sum = sum / (float64(len(vals)) - 1)
	sd := math.Sqrt(sum)

	return float32(sd)
}

// compressByte returns a compressed byte slice.
func CompressByte(src []byte) []byte {
	compressedData := new(bytes.Buffer)
	Compress(src, compressedData, 9)
	return compressedData.Bytes()
}

// decompressByte returns a decompressed byte slice.
func DecompressByte(src []byte) []byte {
	compressedData := bytes.NewBuffer(src)
	deCompressedData := new(bytes.Buffer)
	Decompress(compressedData, deCompressedData)
	return deCompressedData.Bytes()
}

// compress uses flate to compress a byte slice to a corresponding level
func Compress(src []byte, dest io.Writer, level int) {
	compressor, _ := flate.NewWriter(dest, level)
	compressor.Write(src)
	compressor.Close()
}

// compress uses flate to decompress an io.Reader
func Decompress(src io.Reader, dest io.Writer) {
	decompressor := flate.NewReader(src)
	io.Copy(dest, decompressor)
	decompressor.Close()
}

// src is seeds the random generator for generating random strings
var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandStringBytesMaskImprSrc prints a random string
func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// exists returns whether the given file or directory exists or not
// from http://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
// from http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = CopyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
// from http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
func CopyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func SortDictByVal(W map[string]float64) []string {
	var keySorted []string
	reverseMap := map[float64][]string{}
	var valueList sort.Float64Slice
	for k, v := range W {
		reverseMap[v] = append(reverseMap[v], k)
	}
	for k := range reverseMap {
		valueList = append(valueList, k)
	}
	valueList.Sort()
	sort.Sort(sort.Reverse(valueList))

	for _, k := range valueList {
		for _, s := range reverseMap[k] {
			keySorted = append(keySorted, s)
		}
	}
	return keySorted
}

func StringMap2String(stringMap map[string]string) string{
	res := ""

	for k,v := range stringMap{
		res += k+": "+v+" "
	}
	return res
}

// MaxParallelism returns the maximum parallelism https://stackoverflow.com/questions/13234749/golang-how-to-verify-number-of-processors-on-which-a-go-program-is-running
func MaxParallelism() int {
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}

// BindJSON is a shortcut for c.BindWith(obj, binding.JSON)
func BindJSON(obj interface{}, c *gin.Context) error {
	return BindWith(obj, binding.JSON, c)
}

// BindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func BindWith(obj interface{}, b binding.Binding, c *gin.Context) error {
	if err := b.Bind(c.Request, obj); err != nil {
		//c.AbortWithError(400, err).SetType(ErrorTypeBind)
		return err
	}
	return nil
}

func SliceLike(obj1 []interface{},obj2 []interface{}) bool {
	listEqual1 := true
	listEqual2 := true

	itemFound1 := false
	for _,item1 := range obj1{
		for _,item2 := range obj2{
			if (item1 == item2){
				itemFound1 = true
				break
			}
		}
		if (!itemFound1){
			listEqual1 = false
			break
		}
		itemFound1 = false
	}

	itemFound2 := false
	for _,item1 := range obj1{
		for _,item2 := range obj2{
			if (item1 == item2){
				itemFound2 = true
				break
			}
		}
		if (!itemFound2){
			listEqual2 = false
			break
		}
		itemFound2 = false
	}

	if (listEqual1 && listEqual2){
		return true
	}else{
		return false
	}
}

func MapLike(obj1in interface{},obj2in  interface{}) bool {
	listEqual1 := true
	listEqual2 := true

	switch obj1in.(type) {
	case map[string]int:
		obj1 := obj1in.(map[string]int)
		obj2 := obj2in.(map[string]int)

		itemFound1 := false
		for key1,val1 := range obj1{
			for key2,val2 := range obj2{
				if (key1 == key2 ){
					if (val1 == val2){
						itemFound1 = true
						break
					}
				}
			}
			if (!itemFound1){
				listEqual1= false
				break
			}
			itemFound1 = false
		}

		itemFound2 := false
		for key1,val1 := range obj1{
			for key2,val2 := range obj2{
				if (key1 == key2 ){
					if (val1 == val2){
						itemFound2 = true
						break
					}
				}
			}
			if (!itemFound2){
				listEqual2 = false
				break
			}
			itemFound2 = false
		}

		if (listEqual1 && listEqual2){
			return true
		}else{
			return false
		}
	case map[string]float64:
		obj1 := obj1in.(map[string]float64)
		obj2 := obj2in.(map[string]float64)

		itemFound1 := false
		for key1,val1 := range obj1{
			for key2,val2 := range obj2{
				if (key1 == key2 ){
					if (val1 == val2){
						itemFound1 = true
						break
					}
				}
			}
			if (!itemFound1){
				listEqual1= false
				break
			}
			itemFound1 = false
		}

		itemFound2 := false
		for key1,val1 := range obj1{
			for key2,val2 := range obj2{
				if (key1 == key2 ){
					if (val1 == val2){
						itemFound2 = true
						break
					}
				}
			}
			if (!itemFound2){
				listEqual2 = false
				break
			}
			itemFound2 = false
		}

		if (listEqual1 && listEqual2){
			return true
		}else{
			return false
		}
	default:
	}

	return false


}
