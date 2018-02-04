// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// mqtt.go contains functions for performing MQTT transactions.

package routes

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"

	MQTT "github.com/schollz/org.eclipse.paho.mqtt.golang"
	"ParsinServer/glb"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/algorithms"
	"errors"
)

var adminClient *MQTT.Client

func SetupMqtt() {
	server := "tcp://" + glb.RuntimeArgs.MqttServer
        opts := MQTT.NewClientOptions()

        if glb.RuntimeArgs.MqttExisting {
                opts.AddBroker(server).SetClientID(glb.RandStringBytesMaskImprSrc(5)).SetCleanSession(true)
        } else {
                UpdateMosquittoConfig()
                opts.AddBroker(server).SetClientID(glb.RandStringBytesMaskImprSrc(5)).SetUsername(glb.RuntimeArgs.MqttAdmin).SetPassword(glb.RuntimeArgs.MqttAdminPassword).SetCleanSession(true)
        }

	opts.OnConnect = func(c *MQTT.Client) {
		if token := c.Subscribe("#", 1, MessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	adminClient = MQTT.NewClient(opts)

	if token := adminClient.Connect(); token.Wait() && token.Error() != nil {
		glb.Debug.Println(token.Error())
	}
	glb.Debug.Println("Finished setup")
}

func PutMQTT(c *gin.Context) {
	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	reset := strings.ToLower(c.DefaultQuery("reset", "noneasdf"))
	if !glb.RuntimeArgs.Mqtt {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "MQTT is not enabled on this server"})
		return
	}
	if group != "noneasdf" {
		password, err := GetMQTT(group)
		if len(password) == 0 || reset == "true" {
			password, err = SetMQTT(group)
			if err == nil {
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "You have successfuly set your password.", "password": password})
				UpdateMosquittoConfig()
			} else {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Your password exists.", "password": password})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Usage: PUT /mqtt?group=X or reset using PUT /mqtt?group=X&reset=true"})
	}
}

func SetMQTT(group string) (string, error) {
	password := glb.RandStringBytesMaskImprSrc(6)
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.Cwd, "global.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("mqtt"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		err = bucket.Put([]byte(group), []byte(password))
		if err != nil {
			return fmt.Errorf("could add to bucket: %s", err)
		}
		return err
	})
	return password, err
}

func GetMQTT(group string) (string, error) {
	password := ""
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.Cwd, "global.db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("mqtt"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		v := b.Get([]byte(group))
		password = string(v)
		return nil
	})
	return password, nil
}

func UpdateMosquittoConfig() {
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.Cwd, "global.db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	acl := "user " + glb.RuntimeArgs.MqttAdmin + "\ntopic readwrite #\n\n"
	passwd := "admin:" + glb.RuntimeArgs.MqttAdminPassword + "\n"
	conf := "allow_anonymous false\n\nacl_file " + path.Join(glb.RuntimeArgs.Cwd, "mosquitto") + "/acl\n\npassword_file " + path.Join(glb.RuntimeArgs.Cwd, "mosquitto") + "/passwd"

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("mqtt"))
		if b == nil {
			return fmt.Errorf("No such bucket yet")
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			group := string(k)
			pass := string(v)
			acl = acl + "user " + group + "\ntopic readwrite " + group + "/#\n\n"
			passwd = passwd + group + ":" + pass + "\n"
		}

		return nil
	})
	os.MkdirAll(path.Join(glb.RuntimeArgs.Cwd, "mosquitto"), 0644)
	ioutil.WriteFile(path.Join(glb.RuntimeArgs.Cwd, "mosquitto/acl"), []byte(acl), 0644)
	ioutil.WriteFile(path.Join(glb.RuntimeArgs.Cwd, "mosquitto/passwd"), []byte(passwd), 0644)
	ioutil.WriteFile(path.Join(glb.RuntimeArgs.Cwd, "mosquitto/mosquitto.conf"), []byte(conf), 0644)

	cmd := "mosquitto_passwd"
	args := []string{"-U", path.Join(glb.RuntimeArgs.Cwd, "mosquitto/passwd")}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		glb.Warning.Println(err)
	}
	cmd = "kill"
	args = []string{"-HUP", glb.RuntimeArgs.MosquittoPID}
	if err = exec.Command(cmd, args...).Run(); err != nil {
		glb.Warning.Println(err)
	}
}

func SendMQTTLocation(message string, group string, user string) error {
	pubTopic := strings.Join([]string{group, "/location/", user}, "")

	if token := adminClient.Publish(pubTopic, 1, false, message); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Failed to send message")
	}
	return nil
}

func MessageReceived(client *MQTT.Client, msg MQTT.Message) {
	jsonFingerprint, route, err := MqttBuildFingerprint(msg.Topic(), msg.Payload())
	if err != nil {
		return
	}
	glb.Debug.Println("Got valid MQTT request for group " + jsonFingerprint.Group + ", user " + jsonFingerprint.Username)
	if route == "track" {
		algorithms.TrackFingerprint(jsonFingerprint)
	} else {
		algorithms.LearnFingerprint(jsonFingerprint)
	}
}

func MqttBuildFingerprint(topic string, message []byte) (jsonFingerprint parameters.Fingerprint, route string, err error) {
	err = nil
	route = "track"
	topics := strings.Split(strings.ToLower(topic), "/")
	jsonFingerprint.Location = ""
	if len(topics) < 3 || (topics[1] != "track" && topics[1] != "learn") {
		err = fmt.Errorf("Must define track or learn")
		return
	}
	route = topics[1]
	if route == "track" && len(topics) != 3 {
		err = fmt.Errorf("Track needs a user name")
		return
	}
	if route == "learn" {
		if len(topics) != 4 {
			err = fmt.Errorf("Track needs a user name and location")
			return
		} else {
			jsonFingerprint.Location = topics[3]
		}
	}
	jsonFingerprint.Group = topics[0]
	jsonFingerprint.Username = topics[2]
	routers := []parameters.Router{}
	for i := 0; i < len(message); i += 14 {
		if (i + 14) > len(message) {
			break
		}
		mac := string(message[i:i+2]) + ":" + string(message[i+2:i+4]) + ":" + string(message[i+4:i+6]) + ":" + string(message[i+6:i+8]) + ":" + string(message[i+8:i+10]) + ":" + string(message[i+10:i+12])
		val, _ := strconv.Atoi(string(message[i+12 : i+14]))
		rssi := -1 * val
		routers = append(routers, parameters.Router{Mac: mac, Rssi: rssi})
	}
	jsonFingerprint.WifiFingerprint = routers
	return
}
