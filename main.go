package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

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
	Int      *int64   `| ("+" | "-")? @Int`
	Bool     *bool    `| (@"true" | "false")`
	List     []*Value `| ("{" @@ (("," | ";") @@)* ("," | ";")? "}" | "{" "}")`
	Property *Value   `| "[" @@ "]" "=" @@`
	Enum     *string  `| ("ROUGHNESS_METALLIC" | "DECAL" | "RROUGHNESS_METALLIC" | "DIFFUSE" | "default_diff" | "FROM_PATHS" | "SPECULAR")`
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

	for _, plane := range planes {
		liveries, err := ioutil.ReadDir(liveriesFolder + "\\" + plane.Name())
		if err != nil {
			log.Fatal(err)
		}
		for _, liveryFolder := range liveries {
			if liveryFolder.IsDir() {
				livery := parseLivery(liveriesFolder + "\\" + plane.Name() + "\\" + liveryFolder.Name() + "\\description.lua")

				countriesFieldExists := false

				for _, field := range livery.Properties {
					if field.Key == "countries" {
						countriesFieldExists = true
					}
				}

				if countriesFieldExists {
					addCountriesToLivery(livery)
					rebuildLiveryFile(liveriesFolder+"\\"+plane.Name()+"\\"+liveryFolder.Name()+"\\description.lua", livery)
				}
			}
		}
	}
}

func parseLivery(fileLocation string) *LUA {
	fileBytes, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		log.Fatal(err)
	}
	fileString := string(fileBytes)

	commentlessFile := removeComments(fileString)

	parser, err := participle.Build(&LUA{})
	if err != nil {
		log.Fatal(err)
	}

	lua := &LUA{}
	err = parser.ParseString(fileLocation, commentlessFile, lua)
	if err != nil {
		log.Fatal(err)
	}

	return lua
}

func removeComments(fileContents string) string {
	lines := strings.Split(fileContents, "\n")
	newFileContents := ""
	multiLineComment := false

	for _, line := range lines {
		if strings.Contains(line, "--[[") {
			multiLineComment = true
		} else if strings.Contains(line, "--]]") {
			multiLineComment = false
		}

		if !strings.Contains(line, "--") && !multiLineComment && !strings.Contains(line, "local") {
			newFileContents += line + "\n"
		}
	}

	return newFileContents
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

func rebuildLiveryFile(fileName string, livery *LUA) {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	fileString := string(fileBytes)

	countriesIndex := strings.Index(fileString, "countries")

	countriesEnd := 0

	for i := countriesIndex; i < len(fileBytes); i++ {
		if string(fileBytes[i]) == "}" {
			countriesEnd = i + 1
		}
	}

	countriesArray := buildCountriesByteArray(livery)

	writer, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}

	defer writer.Close()

	_, err = writer.Write(fileBytes[0:countriesIndex])
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write(countriesArray)
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write(fileBytes[countriesEnd:len(fileBytes)])
	if err != nil {
		log.Fatal(err)
	}
}

func buildCountriesByteArray(livery *LUA) []byte {
	countries := []byte("countries = {\n")
	for _, field := range livery.Properties {
		if field.Key == "countries" {
			for _, country := range field.Value.List {
				for _, letter := range []byte(*country.String) {
					countries = append(countries, letter)
				}
				countries = append(countries, byte(','))
				countries = append(countries, byte(' '))
			}
		}
	}

	countries = append(countries, byte('\n'))
	countries = append(countries, byte('}'))
	countries = append(countries, byte('\n'))

	return countries
}
