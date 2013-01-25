package main

import (
	"code.google.com/p/gorest"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	_ "github.com/Go-SQL-Driver/MySQL"
	"io/ioutil"
	"net/http"
  "service"
)

type Cfg struct {
	User     string
	Pwd      string
	Database string `json:"db"`
}

var Config Cfg

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


type Bestand struct {
	XMLName  xml.Name `xml:"product"`
	Id       int      `xml:"id"`
	Quantity string   `xml:"qty"`
}
type BestandsListe struct {
	XMLName     xml.Name `xml:"Bestand"`
	ListBestand []Bestand
}

type PriceList struct {
	XMLName xml.Name `xml:"Prices"`
  Id      int      `xml:"-"`
	List    []Price
}
type Price struct {
	XMLName xml.Name `xml:"product"`
	Id      int      `xml:"id"`
	Price   string   `xml:"price"`
}

func main() {
	Config.ConfigFrom("config.json")
	gorest.RegisterService(new(service.ProductService))
	gorest.RegisterService(new(BestandService))
	gorest.RegisterService(new(PriceService))
	http.Handle("/", gorest.Handle())
	http.ListenAndServe(":8787", nil)
}

type User struct {
	Id int
	//CustomerId int
	Token           string
	PriceList       int
	FullStock       bool
	FullDescription bool
}

type BestandService struct {
	gorest.RestService `root:"/bestand/"`
	listBestandFull    gorest.EndPoint `method:"GET" path:"/list/token/{token:string}" output:"string"`
}
type PriceService struct {
	gorest.RestService `root:"/price/"`
	listPrice          gorest.EndPoint `method:"GET" path:"/list/token/{token:string}" output:"string"`
}

}



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

