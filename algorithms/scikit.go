package algorithms

import (
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path"
	"time"
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

func ScikitLearn(groupName string) string {
	tempFile := groupName + ".scikit.json"
	var outList map[string]float64
	//var _, err = os.Stat(path.Join(glb.RuntimeArgs.SourcePath, tempFile))
	//if !os.IsNotExist(err) {
	//	var err = os.Remove(path.Join(glb.RuntimeArgs.SourcePath, tempFile))
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//glb.RuntimeArgs.NeedToFilter[groupName] = true
	//dbm.SetRuntimePrf(groupName,"NeedToFilter",true)
	//
	//// Check existence of the groupName
	//exist := dbm.GroupExists(groupName)
	//if !exist {
	//	glb.Error.Println("groupName not exists")
	//	return "nil"
	//}

	// Todo: .scikit file must be in /data folder
	// Creating a file that its name is groupName+.rf.json
	glb.Debug.Println("Writing " + tempFile)
	f, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, tempFile), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return "nil"
	}

	// Write fingerprints to the groupname.rf.json file in format of json(same as dump db result)
	/*_,fingerprintInMemory,err := dbm.GetLearnFingerPrints(groupName,true)
	if err!=nil{
		return "nil"
	}*/

	rd := dbm.GM.GetGroup(groupName).Get_RawData()
	fingerprintInMemory := rd.Get_Fingerprints()
	for _, fp := range fingerprintInMemory {
		bJSON, _ := json.Marshal(fp)
		f.WriteString(string(bJSON) + "\n")
	}
	f.Close()

	// Do learning
	conn, _ := net.Dial("tcp", "127.0.0.1:"+glb.RuntimeArgs.ScikitPort)
	// send to socket
	fmt.Fprintf(conn, groupName+"=")
	// listen for reply
	out, _ := bufio.NewReader(conn).ReadString('\n')

	// After a successful learning, python client response the calculation time to go
	glb.Debug.Println("scikit learn output")

	err = json.Unmarshal([]byte(out), &outList)
	if err != nil {
		glb.Error.Println(err)
	}

	glb.Debug.Println(outList)
	classSuccessResStr := ""
	for _, classSuccessRes := range outList {
		glb.Debug.Printf("Scikit classification success for '%s' is %2.2f", groupName, classSuccessRes)
		classSuccessResStr += " "
	}

	//os.Remove(path.Join(glb.RuntimeArgs.SourcePath, tempFile))
	return classSuccessResStr
}

func ScikitClassify(group string, fingerprint parameters.Fingerprint) map[string]string {
	var algorithmsPrediction map[string]string
	tempFile := RandomString(10)
	d1, _ := json.Marshal(fingerprint)

	glb.Debug.Println(tempFile)
	// Sending track fingerprint to python client as a file
	err := ioutil.WriteFile(path.Join(glb.RuntimeArgs.SourcePath, tempFile+".scikittemp"), d1, 0644)
	if err != nil {
		glb.Error.Println("Could not write file: " + err.Error())
		return algorithmsPrediction
	}

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:"+glb.RuntimeArgs.ScikitPort)

	// send to socket
	//glb.Debug.Println(tempFile)
	fmt.Fprintf(conn, group+"="+tempFile)

	// listen for reply
	message, _ := bufio.NewReader(conn).ReadString('\n')

	//glb.Debug.Println(message)
	err = json.Unmarshal([]byte(message), &algorithmsPrediction)
	if err != nil {
		glb.Error.Println(err)
	}

	//os.Remove(path.Join(glb.RuntimeArgs.SourcePath, tempFile+ ".scikittemp"))

	glb.Debug.Println(algorithmsPrediction)

	//if len(algorithmsPrediction) == 0{
	//	return bestLocation, res
	//}

	return algorithmsPrediction
}
