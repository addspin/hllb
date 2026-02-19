package checks

import (
	"hllb/utils"
	"log"
)

var ValidPoolHost []string

func TCPCheck() {
	checkItems, err := utils.ReadCheckConfig("check.yaml")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("checkItems:", checkItems)

}
