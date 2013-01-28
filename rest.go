package main

import (
	"code.google.com/p/gorest"
	"net/http"
  "bitbucket.org/nhp/dsapi/service"
)

func main() {
	service.Config.ConfigFrom("config.json")
	gorest.RegisterService(new(service.ProductService))
	gorest.RegisterService(new(service.BestandService))
	gorest.RegisterService(new(service.PriceService))
	http.Handle("/", gorest.Handle())
	http.ListenAndServe(":8787", nil)
}
