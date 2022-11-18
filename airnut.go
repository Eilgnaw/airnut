package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
)

// func InitDB(){
//     dbFileName := "airnut.db"
// 	db, err := sql.Open("sqlite3", dbFileName)
// 	if err != nil {
// 		fmt.Println(err)
// 		fmt.Scanln()
// 	}
// 	defer db.Close()

// 	sqlStmt := `
// 	create table airNut (id integer not null primary key, t text, h text, pm25 text, time text);
// 	`
// 	_, err = db.Exec(sqlStmt)
// 	if err != nil {
// 		log.Printf("%q: %s\n", err, sqlStmt)
// 		fmt.Scanln()
// 		return
// 	}
// }

func AddData(t string, h string, pm25 string){
	ctime := time.Now().Add(time.Hour * 8).Unix()
	dbFileName := "airnut.db"
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		fmt.Println(err)
		fmt.Scanln()
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		fmt.Scanln()
	}
	stmt, err := tx.Prepare("insert into airNut(t, h, pm25 , time) values(?, ?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		fmt.Scanln()
	}
	defer stmt.Close()
	_, err = stmt.Exec(t, h, pm25, strconv.FormatInt(ctime,10))
	if err != nil {
		fmt.Println(err)
		fmt.Scanln()
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		fmt.Scanln()
	}
}


func main() {
	// InitDB()
	addr := "0.0.0.0:10512"
	tcpAddr, err := net.ResolveTCPAddr("tcp",addr)
	if err != nil {
	log.Fatalf("net.ResovleTCPAddr fail:%s", addr) //等价于print err后，再os.Exit(1)
	}
	
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	}else {
	
		log.Println("rpc listening", addr)	
	}
  
	for {
		conn, err := listener.Accept()
		if err != nil {
		log.Println("listener.Accept error:", err)
		continue
	}
	
	go handle_Client(conn)
	}
} 
  
func handle_Client(conn net.Conn) {
	defer conn.Close()
	var write_buffer1 []byte = []byte("{\"common\": {\"data\": {},\"code\": 0, \"protocol\": \"login\"}}")
	
	hbtimes := 0
	for  {
		read_buffer := make([]byte, 1024)
		_, err1 := conn.Read(read_buffer)
		if err1 != nil {
			fmt.Println("ser Read failed:", err1)
			break
		}
		log.Println("read msg:", string(read_buffer))
		read := string(read_buffer)
		atype := gjson.Parse(read).Get("common.protocol").String()
		switch atype {
		case "login":
			_, err2 := conn.Write(write_buffer1)
			if err2 != nil {
				log.Println("ser send error:", err2)
				break
			}
			log.Println("send login msg:", string(write_buffer1)) 
			
		case "post":
			// {"common":{"source":"air","protocol":"post","device":"Fun_pm25"},"param":{"pm25":"53","ap_mac":"88C397D13EF5","mac":"C89346AA881F","charge":"1","t":"22.520000","h":"57.990002","battery":"61","manual":1}}
			hbtimes = 0  
			AddData(gjson.Parse(read).Get("param.t").String(), gjson.Parse(read).Get("param.h").String(),gjson.Parse(read).Get("param.pm25").String())
			log.Println("battery:", gjson.Parse(read).Get("param.battery").String(),"t", gjson.Parse(read).Get("param.t").String(), "h:",gjson.Parse(read).Get("param.h").String()) 
			
		case "heartbeat":
			hbtimes += 1
			if hbtimes % 15 == 0 {
				var write_buffer []byte = []byte("{\"common\": {\"device\": \"Fun_pm25\", \"protocol\": \"detect\"}, \"param\": {\"fromport\": 8023, \"airid\": 1010695,\"fromhost\": \"one\"}}")
				_, err2 := conn.Write(write_buffer)
				if err2 != nil {
					log.Println("ser send error:", err2)
					break
				}
				
				log.Println("send detect msg:", string(write_buffer)) 
				
			} else {
				ctime := time.Now().Add(time.Hour * 8).Unix()
				var write_buffer []byte = []byte("{\"common\": {\"code\": 0, \"protocol\": \"get_weather\"}, \"param\": {\"weather\": \"1\", \"time\": "+strconv.FormatInt(ctime,10)+"}}")
				_, err2 := conn.Write(write_buffer)
				if err2 != nil {
					log.Println("ser send error:", err2)
					break
				}
				log.Println("send weather msg:", string(write_buffer)) 
			}
		}
	}
	
}