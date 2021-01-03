package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/alecthomas/participle/v2"
)

type LUA struct {
	Properties []*Property `@@*`
}

type Property struct {
	Key   string `@Ident "="`
	Value *Value `@@`
}

type Value struct {
	String   *string  `@String`
	Float    *float64 `| @Float`
	Int      *int64   `| @Int`
	Bool     *bool    `| (@"true" | "false")`
	List     []*Value `| "{" @@ (("," | ";") @@)* ("," | ";")? "}"`
	Property *Value   `| "[" @@ "]" "=" @@`
}

// CountryCodes - Country codes for DCS
var CountryCodes = [...]string{`"RUS"`, `"UKR"`, `"USA"`, `"TUR"`, `"UK"`, `"FRA"`, `"GER"`, `"AUSAF"`, `"CAN"`, `"SPN"`, `"NETH"`, `"BEL"`, `"NOR"`, `"DEN"`, `"ISR"`, `"GRG"`, `"INS"`, `"ABH"`,
	`"RSO"`, `"ITA"`, `"AUS"`, `"SUI"`, `"AUT"`, `"BLR"`, `"BGR"`, `"CZE"`, `"CHN"`, `"HRV"`, `"EGY"`, `"FIN"`, `"GRC"`, `"HUN"`, `"IND"`, `"IRN"`, `"IRQ"`, `"JPN"`, `"KAZ"`, `"PRK"`,
	`"PAK"`, `"POL"`, `"ROU"`, `"SAU"`, `"SRB"`, `"SVK"`, `"KOR"`, `"SWE"`, `"SYR"`, `"YEM"`, `"VNM"`, `"VEN"`, `"TUN"`, `"THA"`, `"SDN"`, `"PHL"`, `"MAR"`, `"MEX"`, `"MYS"`,
	`"LBY"`, `"JOR"`, `"IDN"`, `"HND"`, `"ETH"`, `"CHL"`, `"BRA"`, `"BHR"`, `"NZG"`, `"YUG"`, `"SUN"`, `"RSI"`, `"DZA"`, `"KWT"`, `"QAT"`, `"OMN"`, `"ARE"`, `"CUB"`, `"RSA"`}

func main() {
	var dcsFolder = "C:\\Program Files\\Eagle Dynamics\\DCS World OpenBeta"
	var liveriesFolder = dcsFolder + "\\Bazar\\Liveries"
	planes, err := ioutil.ReadDir(liveriesFolder)
	if err != nil {
		log.Fatal(err)
	}

	// var su25tFolder = liveriesFolder + "\\su-25t"

	// liveries, err := ioutil.ReadDir(su25tFolder)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	for _, plane := range planes {
		liveries, err := ioutil.ReadDir(liveriesFolder + "\\" + plane.Name())
		if err != nil {
			log.Fatal(err)
		}
		for _, livery := range liveries {
			if livery.IsDir() {
				fmt.Println(livery.Name())
				livery := parseLivery(liveriesFolder + "\\" + plane.Name() + "\\" + livery.Name() + "\\description.lua")
				printCountries(livery)
				fmt.Println()
				addCountriesToLivery(livery)
				printCountries(livery)
			}
		}
	}
}

func parseLivery(fileLocation string) *LUA {
	parser, err := participle.Build(&LUA{})
	if err != nil {
		log.Fatal(err)
	}

	r, err := os.Open(fileLocation)
	if err != nil {
		log.Fatal(err)
	}

	lua := &LUA{}
	err = parser.Parse(fileLocation, r, lua)
	if err != nil {
		log.Fatal(err)
	}

	return lua
}

func addCountriesToLivery(livery *LUA) {
	for _, field := range livery.Properties {
		if field.Key == "countries" {
			for _, code := range CountryCodes {
				if checkCodeUniqueness(code, field.Value.List) {
					tempValue := &Value{}
					tempString := code
					tempValue.String = &tempString
					field.Value.List = append(field.Value.List, tempValue)
				}
			}
		}
	}
}

func checkCodeUniqueness(code string, list []*Value) bool {
	for _, cc := range list {
		if *cc.String == code {
			return false
		}
	}

	return true
}

func printCountries(livery *LUA) {
	for _, field := range livery.Properties {
		if field.Key == "countries" {
			for _, country := range field.Value.List {
				fmt.Println(*country.String)
			}
		}
	}
}
