package main
import (
  "database/sql"
  _ "github.com/Go-SQL-Driver/MySQL"
  "fmt"
  "strings"
  "os"
  "encoding/xml"
  "io/ioutil"
  "encoding/json"
)


type Cfg struct {
	User     string
	Pwd      string
	Database string `json:"db"`
  Path     string `json:"xmlpath"`
}

var Config Cfg

func CData(s string) string {
	if s != "" {
		return fmt.Sprintf("<![CDATA[%s]]>", s)
	}
	return s
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

type Product struct {
  ID int
  Sku string `xml:"sku"`
  Name string `xml:"title"`
  Color string `xml:"product_color"`
  Size string `xml:"product_size"`
  Pt int  `xml:"product_type"`
  Status int `xml:"product_status"`
  Uvp string `xml:"price"`
  Standardprice string `xml:"uvp"`
  Ean string `xml:"ean"`
  Shortdescription string `xml:"shortdescription"`
  Longdescription string `xml:"longescription"`


}

type Bestand struct {
  ProductId int `xml:"product_id"`
  Quantity string `xml:"quantity"`
}
func (b Bestand) String() string {
  return fmt.Sprintf("ProductId:%d Quantity:%s", b.ProductId, b.Quantity)
}
type Price struct {
  ProductId int `xml:"product_id"`
  Price string `xml:"product_price"`
  PriceGroup int `xml:"price_group"`
}
func (p Price) String() string {
  return fmt.Sprintf("ProductId:%d Price:%s PriceGroup:%d", p.ProductId, p.Price, p.PriceGroup)
}

func (p Product) String() string {
  return fmt.Sprintf("ID:%d-pt:%s-SKU:%s-uvp:%s", p.ID, p.Color, p.Sku, p.Shortdescription)
}



func main() {
  Config.ConfigFrom("config.json")
  found := 0
  xmlFile, err := os.Open(Config.Path + "Artikeldaten.xml")
  var productList []Product
  var priceList []Price
  var bestandList []Bestand
  if err != nil {
    fmt.Println("Error opening file:", err)
  } else {
    found = 1
    decoder := xml.NewDecoder(xmlFile)
    total := 0
    var inElement string
    for {
      t, _ := decoder.Token()
      if t == nil {
        break
      }
      switch se := t.(type) {
        case xml.StartElement:
          // If we just read a StartElement token
          inElement = se.Name.Local
          if inElement == "Artikel" {
            var p Product
            // decode a whole chunk of following XML into the
            // variable p which is a Page (se above)
            decoder.DecodeElement(&p, &se)

            // Do some stuff with the page.
            productList = append(productList, p)
            fmt.Printf("\t%s\n", p)
            total++
          }
          default:
          }
    }
  }
  xmlFile2, err := os.Open(Config.Path + "EKPreise.xml")
  if err != nil {
    fmt.Println("Error opening file:", err)
  } else {
    found = 2
    decoder := xml.NewDecoder(xmlFile2)
    total := 0
    for {
      t, _ := decoder.Token()
      if t == nil {
        break
      }
      switch se := t.(type) {
        case xml.StartElement:
          inElement := se.Name.Local
          if inElement == "Preis" {
            var pr Price
            decoder.DecodeElement(&pr, &se)
            priceList = append(priceList, pr)
            total++
          }
        default:
      }
    }
  }
  xmlFile, err = os.Open(Config.Path + "Bestand.xml")
  if err != nil {
    fmt.Println("Error opening file:", err)
  } else {
    found = 3
    decoder := xml.NewDecoder(xmlFile)
    total := 0
    for {
      t, _ := decoder.Token()
      if t == nil {
        break
      }
      switch se := t.(type) {
        case xml.StartElement:
          inElement := se.Name.Local
          if inElement == "Bestand" {
            var b Bestand
            decoder.DecodeElement(&b, &se)
            bestandList = append(bestandList, b)
            total++
          }
        default:
      }
    }
  }

  if found == 0 {
    fmt.Printf("No xml file found, aborting")
    return
  }

  db,e := sql.Open("mysql", Config.User + ":" + Config.Pwd + "@unix(/var/run/mysqld/mysqld.sock)/" + Config.Database + "?charset=utf8")

  if e != nil {
    panic(e)
  }
  trans, te := db.Begin()
  if te != nil {
    panic(te)
  }

  for _, s := range productList {
    sql := fmt.Sprintf("REPLACE INTO products (id, sku, name, color, size, product_type,status, uvp, standardprice, ean, shortdescription, longdescription) VALUES('%d', '%s', '%s', '%s', '%s', '%d', '%d', '%s', '%s', '%s', '%s', '%s')",
     s.ID, s.Sku, strings.Replace(s.Name, "'", `"`, -1), s.Color, s.Size, s.Pt, s.Status, s.Uvp, s.Standardprice, s.Ean, strings.Replace(s.Shortdescription, "'", `"`, -1), strings.Replace(s.Longdescription, "'", `'`, -1))
     _, tee := trans.Exec(sql)
    if tee != nil {
      fmt.Printf(sql)
      panic(tee)
    }
  }
  te = trans.Commit()
  if te != nil {
    panic(te)
  }
  trans, te = db.Begin()

  for _, s := range priceList {
    sql := fmt.Sprintf("REPLACE INTO price (price_list, product_id, price) VALUES('%d', '%d', '%s')",
     s.PriceGroup, s.ProductId, s.Price)
     _, tee := trans.Exec(sql)
    if tee != nil {
      fmt.Printf(sql)
      panic(tee)
    }
  }
  te = trans.Commit()
  if te != nil {
    panic(te)
  }
  trans, te = db.Begin()

  for i, s := range bestandList {
    sql := fmt.Sprintf("REPLACE INTO bestand (id, product_id, quantity) VALUES('%d', '%d', '%s')",
      i, s.ProductId, s.Quantity)
     _, tee := trans.Exec(sql)
    if tee != nil {
      fmt.Printf(sql)
      panic(tee)
    }
  }
  te = trans.Commit()
  if te != nil {
    panic(te)
  }

}

