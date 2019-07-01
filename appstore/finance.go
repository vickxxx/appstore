package appstore

import (
	"fmt"
	"log"
	"path"
	"time"
)

// 处理苹果财务报表数据

// GetAppleFinance 获取苹果销售数据
func (ac *AppleClient) GetAppleFinance(date string) ([]byte, error) {

	qs := make(map[string]string)
	qs["filter[reportDate]"] = date
	qs["filter[reportType]"] = "FINANCIAL"
	qs["filter[regionCode]"] = "ZZ" // 所有时区
	qs["filter[vendorNumber]"] = ac.VendorNO

	return ReqApple(FinanceURL, ac.JwtToken, qs)
}

// GetFinanceStat 获取苹果财务报表，按月汇总，时间T+35
// date 为PST时区日期
func (ac *AppleClient) GetFinanceStat(fy time.Time) (AppleStatData, error) {
	fmt.Println(fy)
	d := fy.Format("2006-01")
	ss, err := ac.GetAppleFinance(d)
	if err != nil {
		log.Fatal("获取苹果财务数据失败")
		return nil, err
	}
	// fmt.Println(string(ss))
	fileName := fmt.Sprintf("F_D_%s_%s.txt", ac.VendorNO, d)
	SaveFile(path.Join(DLDir, fileName), ss)
	// return StatSales(ss)
	financeSummary, err := StatReporter(ss, "finance")
	fmt.Println(financeSummary)
	return financeSummary, err
}
