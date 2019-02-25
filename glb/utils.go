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
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
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

func RemoveStringSliceItem(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
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
func FindStringInSlice(s string, strings []string) int {
	for i, k := range strings {
		if s == k {
			return i
		}
	}
	return -1
}

func Int64InSlice(item int64, itemList []int64) bool {
	for _, k := range itemList {
		if item == k {
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

func SortReverseDictByVal(W map[string]float64) []string {
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

func SortIntKeyDictByIntVal(W map[int]int) []int {
	var keySorted []int
	reverseMap := map[int][]int{}
	var valueList sort.IntSlice
	for k, v := range W {
		reverseMap[v] = append(reverseMap[v], k)
	}
	for k := range reverseMap {
		valueList = append(valueList, k)
	}
	valueList.Sort()
	sort.Sort(valueList)

	for _, k := range valueList {
		for _, s := range reverseMap[k] {
			keySorted = append(keySorted, s)
		}
	}
	return keySorted
}

// Like SortReverseDictByVal but when there are some fingerprints(specific mac) with same rss,
//		sort them according to their timestamp(actually there is no difference between them) to avoid side effects
// 			(because random ordering of these FP may cause some wrong priorities).
//		of course, we must solve this problem by correcting the algorithms. It's difficult, because maybe there are many FPs(specific mac) with same rss
func SortFPByRSS(W map[string]float64) []string {
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
		sort.Strings(reverseMap[k]) // sort keys that have same value by their name (it's not logical but avoid some sideeffects)
		// todo: solve this problem !
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

func RoundLocationDim(loc string) string{
	x_y := strings.Split(loc, ",")
	if !(len(x_y) == 2) {
		err := errors.New("Location names aren't in the format of x,y")
		Debug.Println(err)
	}
	locXstr := x_y[0]
	locYstr := x_y[1]
	locX, _ := strconv.ParseFloat(locXstr, 64)
	locY, _ := strconv.ParseFloat(locYstr, 64)
	locXint :=  int(math.Floor(locX))
	locYint :=  int(math.Floor(locY))
	return IntToString(locXint) + ".0," + IntToString(locYint)+".0"
}
func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 1, 64)
}

func StringToFloat(input_str string) (float64, error) {
	return strconv.ParseFloat(input_str, 64)
}

func StringToInt(input_str string) (int, error) {
	return strconv.Atoi(input_str)
}


func IntToString(input_num int) string {
	// to convert a float number to a string
	return strconv.Itoa(input_num)
}

func Float64toInt(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func Round(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Float64toInt(num*output)) / output
}
func GetFloatPrecision(x float64) int { //It doesn't work correctly(exactly), but i don't need exact precision
	i := 0
	for (true) {
		temp := math.Pow(10, float64(i)) * x
		if temp-float64(int(temp)) == float64(0) {
			return i
		}
		i++
	}
	return 0
}


func MakeRange(min, max int, options ...interface{}) []int {
	optionsLength := len(options)
	if optionsLength > 0 {
		step := options[0].(int)
		if min < max {
			a := make([]int, (max-min)/step+1)
			for i := range a {
				a[i] = min + i*step
			}
			return a
		} else {
			a := make([]int, (min-max)/step+1)
			for i := range a {
				a[i] = min - i*step
			}
			return a
		}

	}else{
		if min < max {
			a := make([]int, max-min+1)
			for i := range a {
				a[i] = min + i
			}
			return a
		} else {
			a := make([]int, min-max+1)
			for i := range a {
				a[i] = min - i
			}
			return a
		}
	}
}

func MakeRangeFloat(min, max float64, options ...interface{}) []float64 {
	optionsLength := len(options)
	if optionsLength > 0 {
		step := options[0].(float64)
		precision := GetFloatPrecision(step)
		if min < max {
			a := make([]float64, int((max-min)/step)+1)
			for i := range a {
				a[i] = Round(min+float64(i)*step, precision)
			}
			return a
		} else {
			a := make([]float64, int((min-max)/step)+1)
			for i := range a {
				a[i] = Round(min-float64(i)*step, precision)
			}
			return a
		}

	} else {
		if min < max {
			a := make([]float64, int(max-min)+1)
			for i := range a {
				a[i] = min + float64(i)
			}
			return a
		} else {
			a := make([]float64, int(min-max)+1)
			for i := range a {
				a[i] = min - float64(i)
			}
			return a
		}
	}
}

func DuplicateCountString(list []string) map[string]int {

	duplicate_frequency := make(map[string]int)

	for _, item := range list {
		// check if the item/element exist in the duplicate_frequency map

		_, exist := duplicate_frequency[item]

		if exist {
			duplicate_frequency[item] += 1 // increase counter by 1 if already in the map
		} else {
			duplicate_frequency[item] = 1 // else start counting from 1
		}
	}
	return duplicate_frequency
}

func DuplicateCountFloat64(list []float64) map[float64]int {

	duplicate_frequency := make(map[float64]int)

	for _, item := range list {
		// check if the item/element exist in the duplicate_frequency map

		_, exist := duplicate_frequency[item]

		if exist {
			duplicate_frequency[item] += 1 // increase counter by 1 if already in the map
		} else {
			duplicate_frequency[item] = 1 // else start counting from 1
		}
	}
	return duplicate_frequency
}

func UniqueListFloat64(list []float64) []float64 {
	resList := []float64{}
	for _, l := range list {
		Additem := true
		for _, temp := range resList {
			if (temp == l) {
				Additem = false
				break
			}
		}
		if Additem {
			resList = append(resList, l)
		}
	}
	return resList
}

func DeleteSliceItemStr(slice []string, item string) []string {
	resSlice := make([]string, len(slice))
	for i, it := range slice {
		if it == item {
			resSlice = append(slice[:i], slice[i+1:]...)
		}
	}
	return resSlice
}

func PowBig(a, n int) *big.Int {
	tmp := big.NewInt(int64(a))
	res := big.NewInt(1)
	for n > 0 {
		temp := new(big.Int)
		if n%2 == 1 {
			temp.Mul(res, tmp)
			res = temp
		}
		temp = new(big.Int)
		temp.Mul(tmp, tmp)
		tmp = temp
		n /= 2
	}
	return res
}

func IsValidXY(dot string) bool {
	x_y := strings.Split(dot, ",")
	if len(x_y) == 2 {
		if (len(strings.TrimSpace(x_y[0])) > 0) && (len(strings.TrimSpace(x_y[1])) > 0) {
			_, err1 := strconv.ParseFloat(x_y[0], 64)
			_, err2 := strconv.ParseFloat(x_y[1], 64)
			if err1 == nil && err2 == nil {
				return true
			}
		}
	}
	return false
}

func GetDotFromString(dotStr string) (float64, float64) {
	x_y := strings.Split(dotStr, ",")
	//Debug.Println(x_y)
	if len(x_y) != 2 {
		Error.Println("Invalid x,y format:", x_y)
	}
	locXstr := x_y[0]
	locYstr := x_y[1]
	locX, _ := strconv.ParseFloat(locXstr, 64)
	locY, _ := strconv.ParseFloat(locYstr, 64)
	return locX, locY
}

func CheckDotFormatString(dotStr string) bool {
	x_y := strings.Split(dotStr, ",")
	//Debug.Println(x_y)
	if len(x_y) != 2 {
		Error.Println("Invalid x,y format:", x_y)
		return false
	}
	locXstr := x_y[0]
	locYstr := x_y[1]
	_, err1 := strconv.ParseFloat(locXstr, 64)
	_, err2 := strconv.ParseFloat(locYstr, 64)

	if (err1 != nil) {
		Error.Println("Invalid x,y format,can't parse:", err1)
		return false
	} else if (err2 != nil) {
		Error.Println("Invalid x,y format,can't parse:", err2)
		return false
	}

	return true
}

func CalcDist(x1, y1, x2, y2 float64) float64 {
	return math.Pow(math.Pow(float64(x1-x2), 2)+math.Pow(float64(y1-y2), 2), 0.5)
}

func Median(arr []int) int {
	l := len(arr)

	if l == 0 {
		Error.Println("empty array")
		return 123456789
	} else if l == 1 {
		return arr[0]
	}

	sort.Ints(arr)
	if l%2 == 0 {
		return int((arr[l/2-1] + arr[l/2]) / 2)
	} else {
		return arr[(l+1)/2-1]
	}
}

// komeil: listing png images from map directory
// good source: https://flaviocopes.com/go-list-files/
func ListMaps() map[string][]int {
	filesAndDimensions := make(map[string][]int)
	var files []string

	err := filepath.Walk(RuntimeArgs.MapDirectory, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".png" {
			files = append(files, info.Name())
			//Debug.Println("path: "+path)
			width, height := getImageDimension(path)
			filesAndDimensions [info.Name()] = []int{width, height}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	//Debug.Println("list of map filesAndDimensions: ", filesAndDimensions)
	sort.Sort(sort.StringSlice(files))
	return filesAndDimensions
}

func getImageDimension(imagePath string) (int, int) {

	fileName := filepath.Base(imagePath)

	if strings.Index(fileName, "dims") != -1 {
		var width, height int
		indexDims := strings.Index(fileName, "dims")
		indexDot := strings.Index(fileName, ".")
		dimStr := fileName[indexDims+5 : indexDot]
		hwIndex := strings.Index(dimStr, "_")
		width, err1 := strconv.Atoi(dimStr[:hwIndex])
		height, err2 := strconv.Atoi(dimStr[hwIndex+1:])

		if err1 != nil || err2 != nil {
			Error.Println(err1)
			Error.Println(err2)
			return 0, 0
		} else {
			return width, height
		}

	} else {
		file, err := os.Open(imagePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}

		image, _, err := image.DecodeConfig(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
		}
		return image.Width, image.Height
	}

}

func GetLocationOfFingerprint (location string) (float64,float64){ // komeil: gets location return from a fingerprint and returns two float64 numbers
	x_y := strings.Split(location, ",")
	if !(len(x_y) == 2) {
		err := errors.New("Location names aren't in the format of x,y")
		panic(err)
	}
	locXstr := x_y[0]
	locYstr := x_y[1]
	locX, _ := strconv.ParseFloat(locXstr,64)
	locY, _ := strconv.ParseFloat(locYstr,64)
	locX = Round(locX, 5)
	locY = Round(locY, 5)
	return locX,locY
}

func SortedInsertInt64(s []int64, f int64) []int64 {
	l := len(s)
	if l == 0 {
		return []int64{f}
	}

	newS := []int64{}
	for i, vali := range s {
		if (vali > f) {
			newS = append(newS, f)
			newS = append(newS, s[i:]...)
			return newS
		} else {
			newS = append(newS, vali)
		}

	}
	newS = append(s, f)
	return newS
}

func SortedInsertInt(s []int, f int) []int {
	l := len(s)
	if l == 0 {
		return []int{f}
	}

	newS := []int{}
	for i, vali := range s {
		if (vali > f) {
			newS = append(newS, f)
			newS = append(newS, s[i:]...)
			return newS
		} else {
			newS = append(newS, vali)
		}

	}
	newS = append(s, f)
	return newS
}

func GetGraphSlicesRangeRecursive(first []float64, last []float64, step float64) ([][]float64) {
	return getGraphSlicesRangeRecursive(first, last, 0, step)
}
// gets first slice, last slice and level which it must be entered 0 at normal calls. (its goal is for recursive calls inside the function)
func getGraphSlicesRangeRecursive(first []float64, last []float64, level int, step float64) ([][]float64) {
	if len(first) != len(last){
		errors.New("length of first and last are not equal")
	}
	result := [][]float64{}

	if level == len(first){
		//fmt.Println("level at end" ,level)
		return result
	}else {
		a := []float64{}
		for u:=0;u<len(first);u++{
			a=append(a,0)
		}
		copy(a, first)

		for j := float64(0); j < (last[level] - first[level]); j += step {

			b := []float64{}
			for u:=0;u<len(first);u++{
				b=append(b,0)
			}
			copy(b, a)
			result = append(result, b)

			a[level] += step
			tempResults := getGraphSlicesRangeRecursive(a, last, level+1, step)

			result = append(result,tempResults...)
		}
		b := []float64{}
		for u:=0;u<len(first);u++{
			b=append(b,0)
		}
		copy(b, a)
		result = append(result, b)
	}
	//fmt.Println(result)
	finalResult := [][]float64{}
	new := true
	for i := 0; i < len(result); i++ {
		for j:=0;j<len(finalResult);j++{
			if testEq(result[i],finalResult[j])==true {
				//fmt.Println(result[i]," ", finalResult[j])
				new = false
			}
		}
		for k := 0; k < len(result[i])-1; k++ {
			if (result[i][k] < result[i][k+1]) {
				new = false
			}
		}
		if new==true {
			finalResult = append(finalResult, result[i])
		}
		new = true
	}
	return finalResult
}

func testEq(a, b []float64) bool { //tests whether two slices are the same
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false;
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

//// Deprecated use  json.Unmarshal([]byte(str), &newStruct)
//func ConvertStr2IntSlice(str string)([]int,error){
//	str = strings.TrimSpace(str)
//	if len(str) < 3 { // like [0]
// 		return nil,errors.New("Invalid slice string")
//	}
//	str = str[1:][:len(str)-2]
//	sliceStrSplited := strings.Split(str, ",")
//	intSlice := []int{}
//
//	for _, numStr := range sliceStrSplited {
//		num, err := strconv.Atoi(numStr)
//		if err != nil {
//			//Error.Println(err)
//			return nil,errors.New("Invalid slice string")
//		}
//		intSlice = append(intSlice, num)
//	}
//	return intSlice,nil
//}
//
//
//func ConvertStr22DimIntSlice(str string)([][]int,error){
//	str = strings.TrimSpace(str)
//	str = str[1:][:len(str)-2]
//
//	// first converting to int slices
//	StrSplited1 := strings.Split(str, "[")
//	subSlicesStr := []string{}
//	for _,subStr := range StrSplited1[1:]{
//		StrSplited2 := strings.Split(subStr, "]")
//		subSlicesStr = append(subSlicesStr, StrSplited2[0])
//	}
//
//	// then convert each int slice string to int slice
//	resSlice := [][]int{}
//	for _,sliceStr := range subSlicesStr{
//		intSlice, err := ConvertStr2IntSlice("["+sliceStr+"]")
//		if err!=nil{
//			//Error.Println(err)
//			return nil,errors.New("Invalid 2 dim slice string")
//		}
//		resSlice = append(resSlice,intSlice)
//	}
//
//	return resSlice,nil
//}
