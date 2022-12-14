package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const weatherURL = "https://api.help.bj.cn/apis/weather/?id=101180101"

var weatherCode = map[string]int{"晴": 0, "阴": 1, "多云": 1, "雨": 3, "阵雨": 3, "雷阵雨": 3, "雷阵雨伴有冰雹": 3, "雨夹雪": 6, "小雨": 3, "中雨": 3, "大雨": 3, "暴雨": 3, "大暴雨": 3, "特大暴雨": 3, "阵雪": 5, "小雪": 5, "中雪": 5, "大雪": 5, "暴雪": 5, "雾": 2, "冻雨": 6, "沙尘暴": 2, "小雨转中雨": 3, "中雨转大雨": 3, "大雨转暴雨": 3, "暴雨转大暴雨": 3, "大暴雨转特大暴雨": 3, "小雪转中雪": 5, "中雪转大雪": 5, "大雪转暴雪": 5, "浮沉": 2, "扬沙": 2, "强沙尘暴": 2, "霾": 2}
var wcode = 0

type Pparam struct {
	ID string `json:"ID"`
}

type Temps struct {
	WTYPE string     `json:"type"`
	TEMPS []TempData `json:"temps"`
}

type TempData struct {
	ID   string `json:"id"`
	T    string `json:"t"`
	H    string `json:"h"`
	PM   string `json:"pm"`
	Time string `json:"time"`
}

type TemLData struct {
	ID   string `json:"id"`
	T    string `json:"t"`
	H    string `json:"h"`
	PM   string `json:"pm"`
	OT   string `json:"ot"`
	OH   string `json:"oh"`
	OPM  string `json:"opm"`
	Time string `json:"time"`
}

type ResultLResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []TemLData `json:"data"`
}

type ResultResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    []Temps `json:"data"`
}

var dbpool *sqlitex.Pool

func getTempList(w http.ResponseWriter, r *http.Request) {
	var temps []TemLData
	response := ResultLResponse{Code: 400, Message: "请求失败", Data: temps}
	var req Pparam
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("json.NewDecoder err:%s\n", err.Error())
		http.Error(w, "未知错误", http.StatusBadRequest)
		return
	}
	log.Printf("http.Request :%v\n", req)
	conn := dbpool.Get(r.Context())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)
	stmt := conn.Prep("SELECT id, t, h, pm25, time, ot, oh, opm25 FROM airnut WHERE id < $id order by id DESC limit 10;")
	stmt.SetText("$id", req.ID)
	for {
		if hasRow, err := stmt.Step(); err != nil {
			// ... handle error
		} else if !hasRow {
			break
		}
		temp := TemLData{ID: stmt.GetText("id"), T: stmt.GetText("t"), H: stmt.GetText("h"), Time: stmt.GetText("time"), PM: stmt.GetText("pm25"), OT: stmt.GetText("ot"), OH: stmt.GetText("oh"), OPM: stmt.GetText("opm25")}
		temps = append(temps, temp)
	}
	if len(temps) > 0 {
		response.Message = "请求成功"
		response.Data = temps
		response.Code = 200
	}
	w.Header().Set("Content-Type", "application/json") // and this
	json.NewEncoder(w).Encode(response)
}

func getTempA(w http.ResponseWriter, r *http.Request) {

	var temps []Temps
	var outtemps []TempData
	var intemps []TempData
	response := ResultResponse{Code: 400, Message: "请求失败", Data: temps}
	var req Pparam
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("json.NewDecoder err:%s\n", err.Error())
		http.Error(w, "未知错误", http.StatusBadRequest)
		return
	}
	log.Printf("http.Request :%v\n", req)
	conn := dbpool.Get(r.Context())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)
	stmt := conn.Prep("SELECT id, t, h, pm25, time, ot, oh, opm25 FROM airnut WHERE id < $id order by id DESC limit 10;")
	stmt.SetText("$id", req.ID)
	for {
		if hasRow, err := stmt.Step(); err != nil {
			// ... handle error
		} else if !hasRow {
			break
		}
		intemp := TempData{ID: stmt.GetText("id"), T: stmt.GetText("t"), H: stmt.GetText("h"), Time: stmt.GetText("time"), PM: stmt.GetText("pm25")}
		outtemp := TempData{ID: stmt.GetText("id"), T: stmt.GetText("ot"), H: stmt.GetText("oh"), Time: stmt.GetText("time"), PM: stmt.GetText("opm25")}
		outtemps = append(outtemps, outtemp)
		intemps = append(intemps, intemp)
	}
	outtempst := Temps{WTYPE: "out", TEMPS: outtemps}
	intempst := Temps{WTYPE: "in", TEMPS: intemps}
	temps = append(temps, outtempst)
	temps = append(temps, intempst)
	if len(temps) > 0 {
		response.Message = "请求成功"
		response.Data = temps
		response.Code = 200
	}
	w.Header().Set("Content-Type", "application/json") // and this
	json.NewEncoder(w).Encode(response)
}

func getTemp(w http.ResponseWriter, r *http.Request) {

	var temps []Temps
	var outtemps []TempData
	var intemps []TempData
	response := ResultResponse{Code: 400, Message: "请求失败", Data: temps}
	var req Pparam
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("json.NewDecoder err:%s\n", err.Error())
		http.Error(w, "未知错误", http.StatusBadRequest)
		return
	}
	log.Printf("http.Request :%v\n", req)
	conn := dbpool.Get(r.Context())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)
	//SELECT  t, time
	//   FROM airNut GROUP BY strftime('%H', datetime(time, 'unixepoch'))
	stmt := conn.Prep("SELECT id, t, h, pm25, time, ot, oh, opm25 FROM airnut WHERE id < $id AND t != '' AND ot != '' GROUP BY strftime('%Y%m%d%H', datetime(time, 'unixepoch')) order by id DESC limit 10;")
	stmt.SetText("$id", req.ID)
	for {
		if hasRow, err := stmt.Step(); err != nil {
			// ... handle error
		} else if !hasRow {
			break
		}
		intemp := TempData{ID: stmt.GetText("id"), T: stmt.GetText("t"), H: stmt.GetText("h"), Time: stmt.GetText("time"), PM: stmt.GetText("pm25")}
		outtemp := TempData{ID: stmt.GetText("id"), T: stmt.GetText("ot"), H: stmt.GetText("oh"), Time: stmt.GetText("time"), PM: stmt.GetText("opm25")}
		outtemps = append(outtemps, outtemp)
		intemps = append(intemps, intemp)
	}
	outtempst := Temps{WTYPE: "out", TEMPS: outtemps}
	intempst := Temps{WTYPE: "in", TEMPS: intemps}
	temps = append(temps, outtempst)
	temps = append(temps, intempst)
	if len(temps) > 0 {
		response.Message = "请求成功"
		response.Data = temps
		response.Code = 200
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	addr := "0.0.0.0:10512"
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("net.ResovleTCPAddr fail:%s", addr)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	} else {
		log.Println("rpc listening", addr)
	}
	//:memory:?mode=memory
	dbpool, err = sqlitex.Open("airnut.db", 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()
	http.HandleFunc("/getTemp", getTemp)
	http.HandleFunc("/getTempA", getTempA)
	http.HandleFunc("/getTempList", getTempList)

	go func() {
		log.Println("http.ListenAndServe")

		err := http.ListenAndServe(":9090", nil)
		if err != nil {
			log.Fatalf("http listen fail: %s", err)
		}
	}()

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
	write_buffer1 := []byte("{\"common\": {\"data\": {},\"code\": 0, \"protocol\": \"login\"}}")
	for {
		read_buffer := make([]byte, 1024)
		_, err1 := conn.Read(read_buffer)
		if err1 != nil {
			fmt.Println("ser Read failed:", err1)
			return
		}
		log.Println("read msg:", string(read_buffer))
		read := string(read_buffer)
		atype := gjson.Parse(read).Get("common.protocol").String()
		switch atype {
		case "login":
			_, err2 := conn.Write(write_buffer1)
			if err2 != nil {
				log.Println("ser send error:", err2)
				return
			}
			log.Println("send login msg:", string(write_buffer1))

		case "post":
			result := getWeather()
			var ot, oh, opm25 string
			if result != "" {
				wcode = weatherCode[gjson.Parse(result).Get("weather").String()]
				ot = gjson.Parse(result).Get("temp").String()
				oh = gjson.Parse(result).Get("humidity").String()
				opm25 = gjson.Parse(result).Get("pm25").String()
			}
			AddData(gjson.Parse(read).Get("param.t").String(), gjson.Parse(read).Get("param.h").String(), gjson.Parse(read).Get("param.pm25").String(), ot, oh, opm25)
			log.Println("battery:", gjson.Parse(read).Get("param.battery").String(), "t", gjson.Parse(read).Get("param.t").String(), "h:", gjson.Parse(read).Get("param.h").String())

		case "get_weather":
			ctime := time.Now().Add(time.Hour * 8).Unix()
			write_buffer := []byte("{\"common\": {\"code\": 0, \"protocol\": \"get_weather\"}, \"param\": {\"weather\":\"" + strconv.Itoa(wcode) + "\", \"time\":" + strconv.FormatInt(ctime, 10) + "}}")
			_, err2 := conn.Write(write_buffer)
			if err2 != nil {
				log.Println("ser get_weather error:", err2)
				return
			}
		case "heartbeat":
			log.Println("heartbeat")
		}
	}
}

func AddData(t string, h string, pm25 string, ot string, oh string, opm25 string) {
	ctime := time.Now().Unix()
	conn, err := sqlite.OpenConn("airnut.db", sqlite.OpenReadWrite|sqlite.OpenNoMutex)
	if err != nil {
		log.Println("sqlite.OpenConn: ", err.Error())
	}
	defer conn.Close()

	err = sqlitex.Execute(conn, "INSERT INTO airNut (t, h, pm25, time, ot, oh, opm25) VALUES (?, ?, ?, ?, ?, ?, ?);", &sqlitex.ExecOptions{
		Args: []interface{}{t, h, pm25, strconv.FormatInt(ctime, 10), ot, strings.Replace(oh, "%", "", -1), opm25},
	})
	if err != nil {
		log.Println("sqlite.Execute: ", err.Error())
	}

}

func getWeather() string {
	resp, err := http.Get(weatherURL)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	fmt.Println(string(b))
	return string(b)
}
