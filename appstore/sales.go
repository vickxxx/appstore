package appstore

import (
	"bytes"
	"fmt"
	"path"
	"time"
)

// GetAppleSales 获取苹果销售数据
// freq: 报告粒度 DAILY , WEEKLY, MONTHLY, YEARLY
func (ac *AppleClient) GetAppleSales(date string, freq string) ([]byte, error) {
	qs := make(map[string]string)
	qs["filter[reportDate]"] = date
	qs["filter[reportSubType]"] = "SUMMARY"
	qs["filter[reportType]"] = "SALES"
	qs["filter[frequency]"] = freq
	qs["filter[vendorNumber]"] = ac.VendorNO

	return ReqApple(SalesURL, ac.JwtToken, qs)
}

// GetSalesStatInFYP 按照财务月获取销售数据
func (ac *AppleClient) GetSalesStatInFYP(fy time.Time) (AppleStatData, error) {
	// fy 格式 2018,02
	// t, _ := time.Parse("2006,01", fy)
	fYear := fy.Year()
	fPeriod := int(fy.Month())

	period := GetPeriod(fYear, fPeriod)
	beginDate := period.BeginDate
	endDate := period.EndDate

	var appleRetLst [][]byte
	fmt.Println(beginDate, "daily")
	firstday, _ := ac.GetAppleSales(beginDate.Format("2006-01-02"), "DAILY")

	appleRetLst = append(appleRetLst, firstday)

	beginDate = beginDate.Add(24 * time.Hour)
	for week := 0; week < period.WeekCount-1; week++ {
		beginDate = beginDate.Add(6 * 24 * time.Hour)
		fmt.Println(beginDate, "weekly")
		var weekRet []byte
		weekRet, _ = ac.GetAppleSales(beginDate.Format("2006-01-02"), "WEEKLY")
		appleRetLst = append(appleRetLst, weekRet)

		beginDate = beginDate.Add(24 * time.Hour)
	}

	for {
		fmt.Println(beginDate, "daily")
		var dayRet []byte
		dayRet, _ = ac.GetAppleSales(beginDate.Format("2006-01-02"), "DAILY")
		appleRetLst = append(appleRetLst, dayRet)

		beginDate = beginDate.Add(24 * time.Hour)
		if beginDate.After(endDate) {
			break
		}
	}

	var buffer bytes.Buffer
	for _, i := range appleRetLst {
		buffer.Write(i)
	}
	reporter, err := StatReporter(buffer.Bytes(), "sales")
	// fmt.Println(reporter)
	fileName := fmt.Sprintf("S_FY_%s_%s.txt", ac.VendorNO, fy.Format("2006,01"))
	SaveFile(path.Join(DLDir, fileName), buffer.Bytes())
	return reporter, err
}

// GetSalesStat 获取苹果销售分类统计结果
// date 为PST时区日期
func (ac *AppleClient) GetSalesStat(date string) (AppleStatData, error) {

	ss, _ := ac.GetAppleSales(date, "DAILY")
	fileName := fmt.Sprintf("S_D_%s_%s.txt", ac.VendorNO, date)
	SaveFile(path.Join(DLDir, fileName), ss)
	return StatReporter(ss, "sales")
}
