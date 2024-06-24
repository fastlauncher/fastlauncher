package log

import (
	baselog "log"
	"time"
)

func Println(params ...any) {
	baselog.Println(params...)
	time.Sleep(time.Second * 15)
}
