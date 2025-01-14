# raspberry pi myfare card checkin & chat app

updated to cgo free code

myfare: myfare card application folder(go_chat function was integrated @2023/4/30)

uid.json file is not presented here, but you can simply create it from its structure like following
[
{"uid": "xxxxxxxxx","name": "yyyy"}
]
two *.db files are automatically created

serial: receive myfare card uid from M5Stackc Plus

go_chat: simple chat app using gin/melody/sqlite3
