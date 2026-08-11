package ordermanager

import (
	"fmt"
	"strings"
)

func GenCollectionListedKey(chain string, addr string) string {
	return fmt.Sprintf("cache:%s:collection:listed:%s", strings.ToLower(chain), strings.ToLower(addr))
}
