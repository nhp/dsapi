package service

import (
	"io/ioutil"
  "encoding/json"
  "fmt"
	"database/sql"
	_ "github.com/Go-SQL-Driver/MySQL"

)
type User struct {
	Id int
	//CustomerId int
	Token           string
	PriceList       int
	FullStock       bool
	FullDescription bool
}


type Cfg struct {
	User     string
	Pwd      string
	Database string `json:"db"`
}

var Config Cfg

func CData(s string) string {
	if s != "" {
		return fmt.Sprintf("<![CDATA[%s]]>", s)
	}
	return s
}

func GetUser(s string) User {
	db, e := sql.Open("mysql", Config.User+":"+Config.Pwd+"@unix(/var/run/mysqld/mysqld.sock)/"+Config.Database+"?charset=utf8")
	if e != nil {
		panic(e)
	}
	row := db.QueryRow("SELECT id, price_list, full_stock, full_description FROM user WHERE token = '" + s + "'")

	var user User
	user.Token = s
	row.Scan(&user.Id, &user.PriceList, &user.FullStock, &user.FullDescription)
	return user

}
func (l *Cfg) ConfigFrom(path string) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &l)
	if err != nil {
		fmt.Print("bad json ", err)
	}
	return
}

