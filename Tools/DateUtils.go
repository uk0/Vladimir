package Tools

import (
	"fmt"
	"time"
)

func Before(t1s string , t2s string)  bool{
	//先把时间字符串格式化成相同的时间类型
	t1, err := time.Parse("2006-01-02", t1s)
	t2, err := time.Parse("2006-01-02", t2s)
	if err == nil && t1.Before(t2) {
		//处理逻辑
		return true
	}
	return false
}

func GetNowDate(day int, formatStr string) string {
	nTime := time.Now()
	yesTime := nTime.AddDate(0, 0, day)
	ymdPath := yesTime.Format(fmt.Sprintf("2006%s01%s02", formatStr, formatStr))
	return ymdPath;
}

func GetNowTime() string {
	return time.Now().Format("2006-01-02 15:04:05");
}



func main() {
	fmt.Println(Before("2019-04-08","2019-04-09"))
}