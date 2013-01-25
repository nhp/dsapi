package service

import (
	"code.google.com/p/gorest"
  "fmt"
	"database/sql"
  "encoding/xml"
	_ "github.com/Go-SQL-Driver/MySQL"
)
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

type PriceService struct {
	gorest.RestService `root:"/price/"`
	listPrice          gorest.EndPoint `method:"GET" path:"/list/token/{token:string}" output:"string"`
}
func (serv PriceService) ListPrice(token string) string {
	var query string
	user := GetUser(token)
	if user.Id > 0 && user.PriceList > 0 {
		query = fmt.Sprintf("SELECT product_id, price FROM price where price_list = '%d'", user.PriceList)

		db, e := sql.Open("mysql", Config.User+":"+Config.Pwd+"@unix(/var/run/mysqld/mysqld.sock)/"+Config.Database+"?charset=utf8")
		if e != nil {
			panic(e)
		}
		rows, err := db.Query(query)
		if err != nil {
			panic(e)
		}
		var pl PriceList
		pl.Id = user.PriceList
		for rows.Next() {
			var p Price
			rows.Scan(&p.Id, &p.Price)
			pl.List = append(pl.List, p)
		}
		t, e := xml.MarshalIndent(pl, "  ", "    ")
		if e != nil {
			panic(e)
		}
		return xml.Header + string(t)
	} else {
		return "Invalid path for request"
	}
	return ("No valid price list for your user-token")

}

