package bs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type UnixTime time.Time

func (t *UnixTime) UnmarshalJSON(data []byte) (err error) {
	r := strings.Replace(string(data), `"`, ``, -1)
	ti, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(ti/1000, 0)
	return
}

func (t UnixTime) MarshalJSON() ([]byte, error) {
	ts := fmt.Sprintf("%v", time.Time(t).Unix()*1000)
	return []byte(ts), nil
}
