package wapi

import (
	"strconv"
	"strings"
	"time"
)

func Wash(log, method string) string {
	switch method {
	case "nginx": //nginx功能说明：从nginx的日志流中筛选出有效日志，并拼接成kafka的字符串，发送至kafka服务器
		if strings.LastIndex(log, "GET") > 0 {
			if strings.LastIndex(log, "WebAudio-1.0-SNAPSHOT") > 0 {
				s := strings.Split(log, "[")
				m := strings.Split(s[1], "]")[0]
				w, _ := time.Parse("02/Jan/2006:15:04:05 -0700", m)

				msg := "{\"V\":\"2\",\"T\":\"StatisticsMsg\",\"M\":\"{\\\"c\\\":" + strconv.FormatInt(w.Unix()*1000, 10) + ",\\\"cid\\\":0,\\\"t\\\":\\\"m\\\",\\\"ts\\\":0,\\\"ec\\\":0}\"}"
				//fmt.Println(msg)
				return msg
			} else {
				return ""
			}
		} else {
			return ""
		}
	}
	return ""
}
