package appstore

import (
	"bytes"
	"log"
	"strconv"

	tsv "github.com/valyala/tsvreader"
)

// Item TSV文件每一行的数据格式
type Item map[string]string

// Rows TSV文件的所有行数据
type Rows []Item

// 对苹果数据进行分析汇总

// StatReporter 统计销售报表或者财务报表
// skuKey: 商品的识别字段，销售报表为SKU字段，财务报表为 Vendor Identifier
// countKey: 计数字段，销售报表为Units字段，财务报表为 Quantity
// reportType: 报告类型，sales 或者 finance
func StatReporter(raw []byte, reportType string) (AppleStatData, error) {
	var skuKey, countKey string
	switch reportType {
	case "sales":
		skuKey = "SKU"
		countKey = "Units"
	case "finance":
		skuKey = "Vendor Identifier"
		countKey = "Quantity"
	}

	rawReader := bytes.NewReader(raw)
	tsvReader := tsv.New(rawReader)
	var titles []string
	var dataLst Rows
	for tsvReader.Next() {
		var tmpTitle []string
		item := make(Item)
		index := 0
		for tsvReader.HasCols() {
			s := tsvReader.String()
			if len(titles) == 0 { // 提取标题
				tmpTitle = append(tmpTitle, s)
				continue
			}
			if s == "Total_Rows" {
				// sku分列部分完成，不再解析以下数据
				goto FINISH
			}
			item[titles[index]] = s
			index++
		}

		if len(tmpTitle) > 0 { // 提取标题
			titles = tmpTitle
			continue
		}
		dataLst = append(dataLst, item)
	}
FINISH:

	statSku := make(AppleStatData)
	for _, row := range dataLst {
		sku := row[skuKey]
		countStr := row[countKey]

		count, err := strconv.Atoi(countStr)
		if err != nil {
			log.Fatal("转换数字出错")
			return nil, err
		}
		if statSku[sku] == nil {
			statSku[sku] = map[string]int{
				"sales":  0,
				"refund": 0,
			}
		}
		if count > 0 {
			statSku[sku]["sales"] = statSku[sku]["sales"] + count
		} else {
			statSku[sku]["refund"] = statSku[sku]["refund"] + count
		}
	}
	return statSku, nil
}
