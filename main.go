package main

import (
	"fmt"
	"time"

	"github.com/vickxxx/appstore/appstore"
)

func main() {
	fmt.Println("xx")
	fy2019 := appstore.GenFY(2019)
	fmt.Println(fy2019)

	ac := appstore.AppleClient{
		KeyID:    "KeyID",
		IssID:    "IssID",
		VendorNO: "VendorNO",
		PrivKey:  "PrivKey",
	}
	ac.Init()

	fy201905 := time.Date(2019, time.May, 1, 0, 0, 0, 0, time.UTC)

	ac.GetFinanceStat(fy201905) // 2019财年p5财务数据

	ac.GetSalesStatInFYP(fy201905) //  2019财年p5销售数据

	ac.GetAppleSales("20190622", "DAILY") // 2019年6月22日销售数据

	// appstore.GenFYDetail()
}
