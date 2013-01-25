package service

import (
	"code.google.com/p/gorest"
	"database/sql"
  "encoding/xml"
	_ "github.com/Go-SQL-Driver/MySQL"
)

type ProductList struct {
	XMLName  xml.Name `xml:"Products"`
	Products []ListProduct
}

type ListProduct struct {
	XMLName     xml.Name `xml:"product"`
	Id          int      `xml:"id"`
	Sku         string   `xml:"sku"`
	Name        string   `xml:"title"`
}

type Product struct {
	XMLName     xml.Name `xml:"product"`
	Id          int      `xml:"id"`
	Sku         string   `xml:"sku"`
	Name        string   `xml:"title"`
	ShortDesc   string   `xml:"short_description"`
	Description string   `xml:"description"`
	Color       string   `xml:"color"`
	Size        string   `xml:"size"`
	Uvp         string   `xml:"uvp"`
	Status      int      `xml:"status"`
	StdPrice    string   `xml:"std_price"`
	Ean         string   `xml:"ean"`
	ProductType int      `xml:"product_type"`
}

type ProductService struct {
	gorest.RestService `root:"/product/"`
	productDetailsFull gorest.EndPoint `method:"GET" path:"/detail/id/{id:string}/token/{token:string}" output:"string"`
	listProduct        gorest.EndPoint `method:"GET" path:"/list/token/{token:string}" output:"string"`
}
func (serv ProductService) ListProduct(token string) string {
	user := GetUser(token)
	if user.Id == 0 {
		return "Invalid path for request"
	}
	db, e := sql.Open("mysql", Config.User+":"+Config.Pwd+"@unix(/var/run/mysqld/mysqld.sock)/"+Config.Database+"?charset=utf8")
	if e != nil {
		panic(e)
	}
	rows, err := db.Query("SELECT id, sku, name FROM products")
	if err != nil {
		panic(e)
	}
	var pl ProductList

	for rows.Next() {
		var p ListProduct
		rows.Scan(&p.Id, &p.Sku, &p.Name)
		pl.Products = append(pl.Products, p)
	}
	t, e := xml.MarshalIndent(pl, "  ", "    ")
	if e != nil {
		panic(e)
	}
	return xml.Header + string(t)
}

func (serv ProductService) ProductDetailsFull(id string, token string) string {

	var query string
	user := GetUser(token)
	if user.Id > 0 && user.FullDescription {
		query = "SELECT id, sku, name, ean, color, size, product_type, status, uvp, standardprice, shortdescription, longdescription FROM products where id = '" + id + "'"
	} else if user.Id > 0 && !user.FullDescription {
		query = "SELECT id, sku, name, ean, color, size, product_type, status, uvp, standardprice, shortdescription = NULL, longdescription = NULL FROM products where id = '" + id + "'"
	} else {
		return "Invalid path for request"
	}

	db, e := sql.Open("mysql", Config.User+":"+Config.Pwd+"@unix(/var/run/mysqld/mysqld.sock)/"+Config.Database+"?charset=utf8")

	if e != nil {
		panic(e)
	}

	rows := db.QueryRow(query)

		var p Product
		rows.Scan(&p.Id, &p.Sku, &p.Name, &p.Ean, &p.Color, &p.Size, &p.ProductType, &p.Status, &p.Uvp, &p.StdPrice, &p.ShortDesc, &p.Description)
		p.Name = CData(p.Name)
		p.Description = CData(p.Description)
		p.ShortDesc = CData(p.ShortDesc)
	t, e := xml.MarshalIndent(p, "  ", "    ")
	if e != nil {
		panic(e)
	}
	return xml.Header + string(t)
}
