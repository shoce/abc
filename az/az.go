/*
GoFmt
GoBuildNull
GoBuild
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	alfavit []string = []string{
		"А", "A", "а", "a",
		"Б", "B", "б", "b",
		"В", "V", "в", "v",
		"Г", "G", "г", "g",
		"Д", "D", "д", "d",
		"Е", "E", "е", "e",
		"Ё", "IO", "ё", "io",
		"Ж", "J", "ж", "j",
		"З", "Z", "з", "z",
		"И", "I", "и", "i",
		"Й", "I", "й", "i",
		"К", "K", "к", "k",
		"Л", "L", "л", "l",
		"М", "M", "м", "m",
		"Н", "N", "н", "n",
		"О", "O", "о", "o",
		"П", "P", "п", "p",
		"Р", "R", "р", "r",
		"С", "S", "с", "s",
		"Т", "T", "т", "t",
		"У", "U", "у", "u",
		"Ф", "F", "ф", "f",
		"Х", "H", "х", "h",
		"Ц", "C", "ц", "c",
		"Ч", "X", "ч", "x",
		"Ш", "W", "ш", "w",
		"Щ", "WX", "щ", "wx",
		"Ъ", "-", "ъ", "-",
		"Ы", "Y", "ы", "y",
		"Ь", "I", "ь", "i",
		"Э", "E", "э", "e",
		"Ю", "IU", "ю", "iu",
		"Я", "Q", "я", "q",
	}
	replacer *strings.Replacer
)

func getReplacer() *strings.Replacer {
	if replacer == nil {
		replacer = strings.NewReplacer(alfavit...)
	}
	return replacer
}

func main() {
	msgbb, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(getReplacer().Replace(string(msgbb)))

	return
}
