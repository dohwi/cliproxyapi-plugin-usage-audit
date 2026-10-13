package main

import (
	"fmt"
	"os"

	"github.com/dohwi/cliproxyapi-usage-audit/internal/store"
)

func main() {
	recs, err := store.ReadAll(os.Args[1])
	if err != nil {
		panic(err)
	}
	for _, r := range recs {
		fmt.Println(r.Model)
	}
}
