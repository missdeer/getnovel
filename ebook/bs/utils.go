package bs

import (
	"fmt"
	"strings"
)

func urlFix(uri, host string) string {
	if !strings.HasPrefix(uri, "/") {
		return uri
	}
	if !strings.HasPrefix(uri, "//") {
		return fmt.Sprintf("%s%s", host, uri)
	}
	return fmt.Sprintf("%s:%s", strings.Split(host, ":")[0], uri)
}
