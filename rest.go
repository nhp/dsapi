package main
import (
    "code.google.com/p/gorest"
    "net/http"
    "fmt"
    "database/sql"
    _ "github.com/Go-SQL-Driver/MySQL"
    "encoding/xml"
    "encoding/json"
    "io/ioutil"
)

type Cfg struct{
  User string
  Pwd string
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

type ProductList struct {
  XMLName xml.Name `xml:"Products"`
  Products []Product
}


type Product struct {
  XMLName xml.Name `xml:"product"`
  Id int `xml:"id"`
  Sku string `xml:"sku"`
  Name string `xml:"title,omitempty"`
  ShortDesc string `xml:"short_description,omitempty"`
  Description string `xml:"description,omitempty"`
  Color string `xml:"color,omitempty"`
  Size string `xml:"size,omitempty"`
  Uvp string `xml:"uvp,omitempty"`
  Status int `xml:"status,omitempty"`
  StdPrice string `xml:"std_price,omitempty"`
  Ean string `xml:"ean,omitempty"`
  ProductType int `xml:"product_type,omitempty"`
}

type Bestand struct {
  XMLName xml.Name `xml:"product"`
  Id int `xml:"id"`
  Quantity string `xml:"qty"`
}
type BestandsListe struct {
  XMLName xml.Name `xml:"Bestand"`
  ListBestand []Bestand
}

type PriceList struct {
  XMLName xml.Name `xml:"Prices"`
  Id int `xml:"id,attr"`
  List []Price
}
type Price struct {
  XMLName xml.Name `xml:"product"`
  Id int `xml:"id"`
  Price string `xml:"price"`
}

func main() {
    Config.ConfigFrom("config.json")
    gorest.RegisterService(new(ProductService))
    gorest.RegisterService(new(BestandService))
    gorest.RegisterService(new(PriceService))
    http.Handle("/",gorest.Handle())
    http.ListenAndServe(":8787",nil)
}

type User struct {
  Id int
  //CustomerId int
  Token string
  PriceList int
  FullStock bool
  FullDescription bool
}

type BestandService struct {
  gorest.RestService `root:"/bestand/"`
  listBestandFull gorest.EndPoint `method:"GET" path:"/list/token/{token:string}" output:"string"`
}
type PriceService struct {
  gorest.RestService `root:"/price/"`
  listPrice gorest.EndPoint `method:"GET" path:"/list/token/{token:string}" output:"string"`
}

type ProductService struct {
  gorest.RestService `root:"/product/"`
  productDetailsFull gorest.EndPoint `method:"GET" path:"/detail/id/{id:string}/token/{token:string}" output:"string"`
  listProduct gorest.EndPoint `method:"GET" path:"/list/token/{token:string}" output:"string"`
}


func(serv PriceService) ListPrice(token string) string {
  var query string
  user := GetUser(token)
  if user.Id > 0 && user.PriceList > 0 {
    query = fmt.Sprintf("SELECT product_id, price FROM price where price_list = '%d'", user.PriceList)

    db, e := sql.Open("mysql", Config.User + ":" + Config.Pwd + "@unix(/var/run/mysqld/mysqld.sock)/" + Config.Database + "?charset=utf8")
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
    t,e := xml.MarshalIndent(pl, "  ", "    ")
    if e != nil {
      panic(e)
    }
    return xml.Header + string(t)
  } else {
    return "Invalid path for request"
  }
  return ("No valid price list for your user-token")

}

func(serv BestandService) ListBestandFull(token string) string {
  user := GetUser(token)
  if user.Id == 0 {
      return "Invalid path for request"
  }
  db, e := sql.Open("mysql", Config.User + ":" + Config.Pwd + "@unix(/var/run/mysqld/mysqld.sock)/" + Config.Database + "?charset=utf8")
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
    bl.ListBestand= append(bl.ListBestand, qty)
  }
  t,e := xml.MarshalIndent(bl, "  ", "    ")
  if e != nil {
    panic(e)
  }
  return xml.Header + string(t)
}

func(serv ProductService) ListProduct(token string) string {
  user := GetUser(token)
  if user.Id == 0 {
    return "Invalid path for request"
  }
  db, e := sql.Open("mysql", Config.User + ":" + Config.Pwd + "@unix(/var/run/mysqld/mysqld.sock)/" + Config.Database + "?charset=utf8")
  if e != nil {
    panic(e)
  }
  rows, err := db.Query("SELECT id, sku, name FROM products")
  if err != nil {
    panic(e)
  }
  var pl ProductList

  for rows.Next() {
    var p Product
    rows.Scan(&p.Id, &p.Sku, &p.Name)
    pl.Products = append(pl.Products, p)
  }
  t,e := xml.MarshalIndent(pl, "  ", "    ")
  if e != nil {
    panic(e)
  }
  return xml.Header + string(t)
}

func CData(s string) string {
  if s != "" {
    return fmt.Sprintf("<![CDATA[%s]]>", s)
  }
  return s
}

func GetUser(s string) User {
  db, e := sql.Open("mysql", Config.User + ":" + Config.Pwd + "@unix(/var/run/mysqld/mysqld.sock)/" + Config.Database + "?charset=utf8")
  if e != nil {
    panic(e)
  }
  row := db.QueryRow("SELECT id, price_list, full_stock, full_description FROM user WHERE token = '" + s +"'")

  var user User
  user.Token = s
  row.Scan(&user.Id, &user.PriceList, &user.FullStock, &user.FullDescription)
  return user

}


func(serv ProductService) ProductDetailsFull(id string, token string) string {

  var query string
  user := GetUser(token)
  if user.Id > 0 && user.FullDescription {
    query = "SELECT id, sku, name, ean, color, size, product_type, status, uvp, standardprice, shortdescription, longdescription FROM products where id = '" + id + "'"
  } else if user.Id > 0 && !user.FullDescription {
    query = "SELECT id, sku, name, ean, color, size, product_type, status, uvp, standardprice, shortdescription = NULL, longdescription = NULL FROM products where id = '" + id + "'"
  } else {
    return "Invalid path for request"
  }

  db, e := sql.Open("mysql", Config.User + ":" + Config.Pwd + "@unix(/var/run/mysqld/mysqld.sock)/" + Config.Database + "?charset=utf8")

  if e != nil {
    panic(e)
  }

  rows, err := db.Query(query)
  if err != nil {
    panic(e)
  }
  var pl ProductList

  for rows.Next() {
    var p Product
    rows.Scan(&p.Id, &p.Sku, &p.Name, &p.Ean, &p.Color, &p.Size, &p.ProductType, &p.Status, &p.Uvp, &p.StdPrice, &p.ShortDesc, &p.Description)
    p.Name = CData(p.Name)
    p.Description = CData(p.Description)
    p.ShortDesc = CData(p.ShortDesc)
    pl.Products = append(pl.Products, p)
  }
  t,e := xml.MarshalIndent(pl, "  ", "    ")
  if e != nil {
    panic(e)
  }
  return xml.Header + string(t)
}
