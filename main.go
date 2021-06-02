package main

import (
	"github.com/cgghui/exchange_subscribe/exchange/huobi"
)

func main() {
	huobi.NewConnect("6c30af8a-07852cb8-dbuqg6hkte-397ec", "9f3c006d-de28b80c-c230e555-2bcf2")
	select {}
}
