package service
import (
	"code.google.com/p/gorest"
	"database/sql"
  "encoding/xml"
	_ "github.com/Go-SQL-Driver/MySQL"
)
type Bestand struct {
	XMLName  xml.Name `xml:"product"`
	Id       int      `xml:"id"`
	Quantity string   `xml:"qty"`
}
type BestandsListe struct {
	XMLName     xml.Name `xml:"Bestand"`
	ListBestand []Bestand
}
type BestandService struct {
	gorest.RestService `root:"/bestand/"`
	listBestandFull    gorest.EndPoint `method:"GET" path:"/list/token/{token:string}" output:"string"`
}
func (serv BestandService) ListBestandFull(token string) string {
	user := GetUser(token)
	if user.Id == 0 {
		return "Invalid path for request"
	}
	db, e := sql.Open("mysql", Config.User+":"+Config.Pwd+"@unix(/var/run/mysqld/mysqld.sock)/"+Config.Database+"?charset=utf8")
	if e != nil {
		panic(e)
	}
	var BestandsSql string
	if user.FullStock {
		BestandsSql = "SELECT product_id, quantity FROM bestand"
	} else {
		BestandsSql = "SELECT product_id, if(quantity>0, 1, 0) FROM bestand"
	}
	rows, err := db.Query(BestandsSql)
	if err != nil {
		panic(e)
	}
	var bl BestandsListe

	for rows.Next() {
		var qty Bestand
		rows.Scan(&qty.Id, &qty.Quantity)
		bl.ListBestand = append(bl.ListBestand, qty)
	}
	t, e := xml.MarshalIndent(bl, "  ", "    ")
	if e != nil {
		panic(e)
	}
	return xml.Header + string(t)
}


