package chat

import (
	//"fmt"
	sql_db "myfare-copy/go_chat/sqldb"
	"net/http"
	"time"
	"strconv"

	"encoding/json"

	"github.com/gin-gonic/gin"
	melody "gopkg.in/olahol/melody.v1" // minimalist websocket framework
)

var Chatmsg struct {
	Icon	string `json:"icon"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

func Run() {
	gin.SetMode(gin.ReleaseMode) // suppress debug mode
	r := gin.Default()
	m := melody.New()

	r.Static("/static", "./go_chat/view/static")
	r.LoadHTMLGlob("./go_chat/view/*.html")		// directory root is "Run" call module(main.go) existing directory

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.GET("/room/:name", func(c *gin.Context) {
		chattername := c.Query("name")
		pmsg := chatbuild(chattername)
		c.Writer.WriteString(pmsg)		// to genrate initial chat page instead of the static html
		/*
			c.HTML(http.StatusOK, "room.html", gin.H{
				"Name": c.Param("name"),
			})
		*/
	})

	r.GET("/room/:name/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		json.Unmarshal(msg, &Chatmsg)
		sql_db.DbInsert(Chatmsg.Icon, Chatmsg.Name, Chatmsg.Message)
		m.BroadcastFilter(msg, func(q *melody.Session) bool {
			return q.Request.URL.Path == s.Request.URL.Path
		})

	})
	r.Run(":8082")
}

func chatbuild(chatname string) string {
	//
	// str1 & str2 are base html format for chat page
	//
	str1 := `<html>
	<head>
	  <title>chat</title>
	  <meta name="viewport" content="width=device-width">
	  <link rel="stylesheet" href="/static/css/common.css">
	</head>
	<body>
	  	<div class="line">
			<div id="room" class="line-header">chat room</div>
			<div id="chat" class="line-container">`

	str2 := `</div>
			<div class="line-form">
	  			<input id="text" type="text">
	  			<button class="line-form-button" onclick="send_data()">Send</button> 
			</div>
  		</div>
  		<script type="text/javascript" src="/static/js/index.js"></script>
	</body>
	</html>`

	sql_db.DbRead()

	strC := ""

	//
	// assemble previous chat message
	//
	for _, p := range sql_db.Messages{
		// convert the data base Unix time to time Hour & Minute format
		intt, _ :=strconv.ParseInt(p.Time, 10, 64)
		ttime := time.Unix(intt, 0)
		timestiring := strconv.Itoa(ttime.Hour()) + ":" + strconv.Itoa(ttime.Minute())

		if chatname == p.Name{
			strC += liner(p.Msg, timestiring)
		}else{
			strC += linel(p.Icon, p.Name, p.Msg, timestiring)
		}
	}
	return str1 + strC + str2
}

func liner(msg string, time string) string {
	// lineR & lineL are chat message format
	lineR := `<div class='line-right'>
	<p class='line-right-text'>` + msg + `</p>
	<div class="line-right-time">` + time + `</div>
	</div>`
	return lineR
}

func linel(icon string, name string, msg string, time string) string {
	lineL := `<div class='line-left'>
	<img src="/static/img/` + icon + `.png">
	<div class='line-left-container'>
		<p class='line-left-name'>` + name + `</p>
		<p class='line-left-text'>` + msg + `</p>
		<div class='line-left-time'>` + time + `</div>
	</div>
	</div>`
	return lineL
}
