package domain

import (
	"database/sql/driver"
	"errors"
	"time"
)

// 存datetime 显示datetime
type Time time.Time

// 存datetime 显示date
type Date time.Time

const (
	timeFormart = "2006-01-02 15:04:05"
	dateFormart = "2006-01-02"
)

func (t *Time) UnmarshalJSON(data []byte) error {
	//if string(data) == "null" {
	//	return nil
	//}
	//var err error
	////前端接收的时间字符串
	//str := string(data)
	////去除接收的str收尾多余的"
	//timeStr := strings.Trim(str, "\"")
	//t1, err := time.Parse("2006-01-02 15:04:05", timeStr)
	//*t = Time(t1)

	now, err := time.ParseInLocation(`"`+timeFormart+`"`, string(data), time.Local)
	*t = Time(now)
	return err
}

func (t Time) MarshalJSON() ([]byte, error) {
	//formatted := fmt.Sprintf("\"%v\"", time.Time(t).Format(timeFormart))
	//return []byte(formatted), nil
	b := make([]byte, 0, len(timeFormart)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormart)
	b = append(b, '"')
	return b, nil
}

func (t Time) Value() (driver.Value, error) {
	// Time 转换成 time.Time 类型
	tTime := time.Time(t)
	return tTime.Format(timeFormart), nil
}

func (t *Time) Scan(v interface{}) error {
	switch vt := v.(type) {
	case time.Time:
		// 字符串转成 time.Time 类型
		*t = Time(vt)
	default:
		return errors.New("类型处理错误")
	}
	return nil
}

func (t Time) String() string {
	return time.Time(t).Format(timeFormart)
}

func (t *Date) UnmarshalJSON(data []byte) error {
	now, err := time.ParseInLocation(`"`+timeFormart+`"`, string(data), time.Local)
	*t = Date(now)
	return err
}

func (t Date) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(dateFormart)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, dateFormart)
	b = append(b, '"')
	return b, nil
}

func (t Date) Value() (driver.Value, error) {
	// Time 转换成 time.Time 类型
	tTime := time.Time(t)
	return tTime.Format(timeFormart), nil
}

func (t *Date) Scan(v interface{}) error {
	switch vt := v.(type) {
	case time.Time:
		// 字符串转成 time.Time 类型
		*t = Date(vt)
	default:
		return errors.New("类型处理错误")
	}
	return nil
}

func (t Date) String() string {
	return time.Time(t).Format(dateFormart)
}
