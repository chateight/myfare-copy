//
// chat app skelton is from https://www.cetus-media.info/article/2021/line-chat/
// melody sample https://github.com/olahol/melody/tree/master/examples/multichat
//
// the gin & melody are used to create this application
//


package main

import (
	"fmt"
	"sync"

	chat "myfare-copy/go_chat/chat"
	sql_db "myfare-copy/go_chat/sqldb"

	"html/template"
	"log"
	"myfare-copy/uidSerial"
	"net/http"

	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"golang.org/x/net/websocket"
)

// registered ninjya
type Ninjya struct {
	Uid              string `gorm:"primaryKey"`
	Name             string
	Time             int
	Stat             int
	Timepresentation int
	Presentation     int
}

var ninjyas []Ninjya                // slice of the ninjyas information
var ninjyaSlice []string            // slice of the attended ninjyas name
var ninjyaPresentaionSlice []string // presentation ninjyas slice

var tableName string // tableNmae = "tbl" + year + day from Jan 1st

var mu sync.RWMutex // message from card reader
var msg string      // mutex

// implements TableName of the Tabler interface
func (Ninjya) TableName() string {
	t := time.Now()
	tableName = "tbl" + strconv.Itoa(t.Year()) + strconv.Itoa(t.YearDay())
	return tableName
}

// to get active/presentation ninjyas slice
func ninjya() {
	db, err := gorm.Open(sqlite.Open("./myfare.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Ninjya{}) // if you use Automigrate and change struct, it won't be reflected automatically

	// to make attended nijyas slice
	db.Order("Time desc").Where("Stat = ?", 1).Find(&ninjyas) // SELECT * FROM where Stat = tbl*****;
	ninjyaSlice = nil
	for _, p := range ninjyas {
		ninjyaSlice = append(ninjyaSlice, p.Name)
	}
	// to make presentation ninjyas slice
	db.Order("TimePresentation desc").Where("presentation = ?", 1).Find(&ninjyas) // SELECT * FROM where Stat = tbl*****;
	ninjyaPresentaionSlice = nil

	for _, p := range ninjyas {
		ninjyaPresentaionSlice = append(ninjyaPresentaionSlice, p.Name)
	}
}

// to entry the presentation
func presentationSet(name string) {
	db, err := gorm.Open(sqlite.Open("./myfare.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// to check presentationFalg of the clicked ninjya's name
	db.Where("Name = ?", name).First(&ninjyas)
	presentationFlag := 0
	p := ninjyas[0]
	// toggle the presentaionFlag
	presentationFlag = p.Presentation
	if presentationFlag == 0 {
		presentationFlag = 1
	} else {
		presentationFlag = 0
	}
	// update the db table
	now := time.Now().Unix()
	db.Model(&ninjyas).Where("Name = ?", name).Updates(map[string]interface{}{"TimePresentation": strconv.FormatInt(now, 10), "presentation": presentationFlag})
}

func handler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("layout.html", "pageData.html"))
	// to update ninjya status
	if r.FormValue("nickname") != "" {
		presentationSet(r.FormValue("nickname"))
	}
	// to get updated ninjyas slice(ninjyaSlice/ninjyaPresentaionSlice)
	ninjya()

	tm := time.Now().Format(time.RFC1123)
	err := t.Execute(w, map[string]interface{}{
		"Time":              tm,
		"Slice":             ninjyaSlice,
		"PresentationSlice": ninjyaPresentaionSlice,
	})
	if err != nil {
		fmt.Fprintln(w, err)
	}
}

func wevServer() {
	mux := http.NewServeMux()
	// to include resoureces
	mux.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources/"))))
	mux.Handle("/ws", websocket.Handler(msgHandler))
	mux.HandleFunc("/", handler)
	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  3000 * time.Millisecond,
		WriteTimeout: 3000 * time.Millisecond,
	}
	err := server.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}

func msgHandler(ws *websocket.Conn) {
	defer ws.Close()

	premsg := msg // initialize websocket message
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
label:
	for {
		msgr := ""
		err := websocket.Message.Receive(ws, &msgr)
		if err != nil {
			//log.Println("receive error")		// main pupose is to check timeout (to detect unused session)
			break label
		}

		select {
		// to send websocket message triggered by the timer
		// the reason to separate receive and send is ws are running multi thread
		case <- t.C:
			if premsg != msg {
				premsg = msg
				err := websocket.Message.Send(ws, msg)
				if err != nil {
					log.Println("send err")
					break label
				}
			}
		case msgSerial := <- uidSerial.Notice: // wait for message from serial.go via channel
			mu.Lock()
			msg = msgSerial.(string)
			mu.Unlock()
		}
	}
}

func main() {
	// to call card reader function()
	go uidSerial.SerialMain()

	// chat data base create and make table to store chat messages
	sql_db.DbCreate()
	// start chat service
	go chat.Run()
	
	// start myfare card service
	wevServer()
}
