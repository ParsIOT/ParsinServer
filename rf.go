package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func rfLearn(group string) float64 {
	tempFile := group + ".rf.json"

	//var _, err = os.Stat(path.Join(RuntimeArgs.SourcePath, tempFile))
	//if !os.IsNotExist(err) {
	//	var err = os.Remove(path.Join(RuntimeArgs.SourcePath, tempFile))
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	RuntimeArgs.NeedToFilter[group] = true

	// Check existence of the group
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0664, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Creating a file that its name is group+.rf.json
	Debug.Println("Writing " + tempFile)
	f, err := os.OpenFile(path.Join(RuntimeArgs.SourcePath, tempFile), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return -1
	}

	// Write fingerprints to the groupname.rf.json file in format of json(same as dump db result)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v, true)
			bJSON, _ := json.Marshal(v2)
			f.WriteString(string(bJSON) + "\n")
		}
		return nil
	})
	f.Close()

	// Do learning
	conn, _ := net.Dial("tcp", "127.0.0.1:"+RuntimeArgs.RFPort)
	// send to socket
	fmt.Fprintf(conn, group+"=")
	// listen for reply
	out, _ := bufio.NewReader(conn).ReadString('\n')

	// After a successful learning, python client response the calculation time to go
	Debug.Println("rf learn output")
	Debug.Println(string(out))

	classificationSuccess, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		Error.Println(string(out))
	}
	Debug.Printf("RF classification success for '%s' is %2.2f", group, classificationSuccess)

	os.Remove(tempFile)
	return classificationSuccess
}

func rfClassify(group string, fingerprint Fingerprint) (string, map[string]float64) {
	var m map[string]float64
	var bestLocation string
	tempFile := RandomString(10)
	d1, _ := json.Marshal(fingerprint)

	// Sending track fingerprint to python client as a file
	err := ioutil.WriteFile(tempFile+".rftemp", d1, 0644)
	if err != nil {
		Error.Println("Could not write file: " + err.Error())
		return bestLocation, m
	}

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:"+RuntimeArgs.RFPort)
	// send to socket
	//Debug.Println(tempFile)
	fmt.Fprintf(conn, group+"="+tempFile)
	// listen for reply
	message, _ := bufio.NewReader(conn).ReadString('\n')

	err = json.Unmarshal([]byte(message), &m)
	if err != nil {
		// do nothing
	}

	os.Remove(tempFile + ".rftemp")

	Debug.Println(m)

	if len(m) == 0{
		return bestLocation, m
	}
	bestLocation = sortDictByVal(m)[0]
	Debug.Println(bestLocation)
	return bestLocation, m
}
