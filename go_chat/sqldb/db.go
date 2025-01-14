package sql_db

import (
	//"fmt"
	"strconv"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

//
// chat data store and retrieve using sqlite3 ddatabase
//

type Message struct {
	Time string `gorm:"primaryKey"`
	Icon string
	Name string
	Msg  string
}

func (Message) TableName() string {
	t := time.Now()
	tableName := "tbl" + strconv.Itoa(t.Year()) + strconv.Itoa(t.YearDay())
	return tableName
}

var Messages []Message

func DbCreate(){
	db, err := gorm.Open(sqlite.Open("./chat.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Message{})
}

func DbRead(){
	db, err := gorm.Open(sqlite.Open("./chat.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.Find(&Messages)
}

func DbInsert(icon string, name string, msg string){
	db, err := gorm.Open(sqlite.Open("./chat.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	now := time.Now().Unix() 
	db.Create(&Message{Time: strconv.FormatInt(now, 10), Icon: icon, Name: name, Msg: msg})

}
