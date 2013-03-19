package gogi

import (
	"strings"
)

func CamelCase(str string) (name string) {
	words := strings.Split(str, "_")
	for _, word := range words {
		name += strings.Title(word)
	}
	return
}

// convert an interface name to its implementation name,
// e.g. Window -> windowImpl
// it's made lower-case in order to hide it
func GetImplName(name string) string {
	return strings.ToLower(name[0:1]) + name[1:] + "Impl"
}
