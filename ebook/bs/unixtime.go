package bs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UnixTime convert from JSON value to time type
type UnixTime time.Time

// UnmarshalJSON convert from JSON value
func (t *UnixTime) UnmarshalJSON(data []byte) (err error) {
	r := strings.Replace(string(data), `"`, ``, -1)
	ti, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(ti/1000, 0)
	return
}

// MarshalJSON convert time type to JSON value
func (t UnixTime) MarshalJSON() ([]byte, error) {
	ts := fmt.Sprintf("%v", time.Time(t).Unix()*1000)
	return []byte(ts), nil
}
