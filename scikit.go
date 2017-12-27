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

func scikitLearn(group string) string {
	tempFile := group + ".scikit.json"
	var outList map[string]float64
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
		return "nil"
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
	conn, _ := net.Dial("tcp", "127.0.0.1:"+RuntimeArgs.ScikitPort)
	// send to socket
	fmt.Fprintf(conn, group+"=")
	// listen for reply
	out, _ := bufio.NewReader(conn).ReadString('\n')

	// After a successful learning, python client response the calculation time to go
	Debug.Println("scikit learn output")

	err = json.Unmarshal([]byte(out), &outList)
	if err != nil {
		Error.Println(err)
	}

	Debug.Println(outList)
	classSuccessResStr := ""
	for _, classSuccessRes := range outList {
		Debug.Printf("Scikit classification success for '%s' is %2.2f", group, classSuccessRes)
		classSuccessResStr += " "
	}

	os.Remove(tempFile)
	return classSuccessResStr
}

func scikitClassify(group string, fingerprint Fingerprint) (map[string]string) {
	var algorithmsPrediction map[string]string
	tempFile := RandomString(10)
	d1, _ := json.Marshal(fingerprint)

	// Sending track fingerprint to python client as a file
	err := ioutil.WriteFile(tempFile+".scikittemp", d1, 0644)
	if err != nil {
		Error.Println("Could not write file: " + err.Error())
		return algorithmsPrediction
	}

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:"+RuntimeArgs.ScikitPort)

	// send to socket
	//Debug.Println(tempFile)
	fmt.Fprintf(conn, group+"="+tempFile)

	// listen for reply
	message, _ := bufio.NewReader(conn).ReadString('\n')

	//Debug.Println(message)
	err = json.Unmarshal([]byte(message), &algorithmsPrediction)
	if err != nil {
		Error.Println(err)
	}

	os.Remove(tempFile + ".scikittemp")

	Debug.Println(algorithmsPrediction)

	//if len(algorithmsPrediction) == 0{
	//	return bestLocation, res
	//}

	return algorithmsPrediction
}
