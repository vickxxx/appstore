package appstore

import (
	"strconv"
	"time"
)

// 苹果财年相关操作
// 日历采用utc时区，避免夏令时切换带来的错位

// Period 财务月结构
type Period struct {
	Year      int       // 财务年
	P         int       // 财务月序号，1-12
	WeekCount int       // 本财务月周数
	BeginDate time.Time // 开始日期
	EndDate   time.Time // 结束日期
}

// FinanceYear 财务年，包含12个财务月
type FinanceYear []Period

// FYMap 财年映射表
var (
	FYMap = map[int]FinanceYear{}
)

// GetPeriod 获取财务月
func GetPeriod(year, period int) Period {
	if FYMap[year] == nil {
		// fymap := map[int]FinanceYear{
		// 	2019: GenFY(year),
		// }
		FYMap[year] = GenFY(year)
	}
	return FYMap[year][period-1]
}

func (p Period) String() string {
	return "FY " + strconv.Itoa(p.Year) +
		"Period-" + strconv.Itoa(p.P) +
		" weeks: " + strconv.Itoa(p.WeekCount) +
		"\n=======\nStartDate:\t" + string(p.BeginDate.Format("2006-01-02")) +
		"\nEndDate:\t" + string(p.EndDate.Format("2006-01-02")) + "\n\n"
}

// IsLeapYear 是否闰年
func IsLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

// DaysInMonth 当前月份有多少天
func DaysInMonth(year, month int) int {
	monthHead := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	return monthHead.Day()
}

// GetFYBeginDay 获取财务日历开始第一天日期
func GetFYBeginDay(year int) time.Time {
	oct := time.Date(year-1, time.October, 1, 0, 0, 0, 0, time.UTC)
	dayInWeek := time.Duration(oct.Weekday())
	beginDay := oct.Add(time.Hour * -24 * dayInWeek)
	return beginDay
}

// GenFY 生成苹果财务日历
func GenFY(year int) FinanceYear {
	financeYear := make([]Period, 12)
	beginDay := GetFYBeginDay(year) // 总是从周日开始，10月不够，9月来凑
	crtDay := beginDay
	for quarters := 1; quarters < 5; quarters++ { // 四个季度Q1，Q2, Q3, Q4
		for p := 1; p < 4; p++ { // 3个财务月
			period := (quarters-1)*3 + p

			weekCount := 4 //本财务月包含几周，其他基本为4周
			if p == 1 {
				weekCount = 5 // 每季度第一个月为5周
			}

			if period == 3 { // 第3财务月
				newYearDay := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
				if IsLeapYear(year) && newYearDay.Weekday() == time.Sunday { // 闰年(财务年）,同时下一年元旦为周日，则当月为五周
					weekCount = 5
				}
				if (year-2)%5 == 0 { // 千年是5的倍数，需条调整为5周
					weekCount = 5
				}

			}
			periodStuct := Period{
				Year:      year,
				P:         period,
				WeekCount: weekCount,
				BeginDate: crtDay,
			}

			nHours := time.Duration(weekCount * 7 * 24) // 跳跃4-5周，闭区间
			crtDay = crtDay.Add(nHours * time.Hour)     // 更新当前日期
			periodStuct.EndDate = crtDay.Add(-24 * time.Hour)
			financeYear[period-1] = periodStuct
		}
	}
	return financeYear
}
