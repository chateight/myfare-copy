/*
card reader handler

1. connect to the RF cardreader vis USB port
2. make database table using json file
3. update if new read data is available
*/
package uidSerial

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type Person struct { // uid, name struct for unmarsharing
	Uid  string `json:"uid", gorm:"primaryKey"`
	Name string `json:"name"`
}

var p []Person                   // Uid, Name struct slice
var uidMap = map[string]string{} // uid, name map

var date = "" // file name("tbl" + date) : days from the new year's start

var Notice = make(chan any) // message channel between serial reader and web servvice

func getPortName() (string, error) {
	ports, error := enumerator.GetDetailedPortsList()
	if error != nil {
		return "", error
	}
	for _, port := range ports {

		if port.IsUSB && port.VID == "0403" && port.PID == "6001" {
			return port.Name, nil
		}
	}
	return "", errors.New("M5StackC plus is not conntected")
}

func readJsonFile(f string) {
	// read JSON file
	raw, err := os.ReadFile(f)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	json.Unmarshal(raw, &p)

	for _, v := range p {
		uidMap[v.Uid] = v.Name
	}
}

func tableCreate(del bool) {
	// DataBase table create

	db, err := sql.Open("sqlite3", "./myfare.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	t := time.Now()
	date = strconv.Itoa(t.Year()) + strconv.Itoa(t.YearDay())

	if del {
		cmd := "drop table if exists tbl" + date

		_, err := db.Exec(cmd)
		if err != nil {
			fmt.Println("db table drop error")
		}
	}

	// table exist check
	cmd := "select*from tbl" + date
	_, table_check := db.Query(cmd) // checked by using return code

	if table_check != nil {

		cmd = "create table if not exists tbl" + date + "(uid text primary key, name text, time int, stat integer, timepresentation integer, presentation integer)"
		_, err := db.Exec(cmd)
		if err != nil {
			fmt.Println("db table create error")
			log.Fatalln(err)
		}

		for _, v := range p {
			now := time.Now().Unix()                                                                                                 // get Unix time
			cmd = "insert into tbl" + date + " values('" + v.Uid + "','" + v.Name + "'," + strconv.FormatInt(now, 10) + ", 0, 0, 0)" // insert to the DB table
			_, err = db.Exec(cmd)
			if err != nil {
				fmt.Println("db table insert error")
				log.Fatalln(err)
			}
		}
		fmt.Println(uidMap)
	}
}

func tableUpdate(uid string) {
	// DataBase table create

	db, err := sql.Open("sqlite3", "./myfare.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// table update
	now := time.Now().Unix()                                                                                                                        // get Unix time
	cmd := "update tbl" + date + " set stat = 1, name = '" + uidMap[uid] + "', time = " + strconv.FormatInt(now, 10) + " where uid = '" + uid + "'" // update the DB table
	ret, err := db.Exec(cmd)
	if err != nil {
		fmt.Println("db table update error", ret)
		log.Fatalln(err)
	}
}

func SerialMain() {
	readJsonFile("uid.json")

	tableCreate(false) // arg true drops the table

	// start serial port & wait for data
rep:
	portName, err := getPortName()
	if err != nil {
		Notice <- "Check the card reader connection"
		select {
		case <- time.After(1 * time.Second):
			goto rep
		}
	}
	Notice <- "Please touch your card"

	mode := &serial.Mode{
		BaudRate: 115200,
	}

	port, err := serial.Open(portName, mode)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		uid := scanner.Text()
		if len(uid) < 20 {
			//fmt.Println(uid)
			tableUpdate(uid)
			Notice <- uidMap[uid] + " was added"
		}
	}
}
