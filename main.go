package main

import (
	"encoding/json"
	"fmt"

	"github.com/thisisnttheway/viac-wealth-getter/wealth"
)

func main() {
	wealth, err := wealth.GetWealth()
	if err != nil {
		panic(err)
	}

	o, _ := json.Marshal(wealth)
	fmt.Println(string(o))
}
