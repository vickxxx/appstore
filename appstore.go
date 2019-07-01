package appstore

import (
	"compress/gzip"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// P8PATH 苹果api p8密钥路径
const (
	// DLDir 报表下载路径
	DLDir      = "applestore_dl"
	SalesURL   = "https://api.appstoreconnect.apple.com/v1/salesReports"
	FinanceURL = "https://api.appstoreconnect.apple.com/v1/financeReports"
)

var (
	ErrAuthKeyNotPem   = errors.New("token: AuthKey must be a valid .p8 PEM file")
	ErrAuthKeyNotECDSA = errors.New("token: AuthKey must be of type ecdsa.PrivateKey")
	ErrAuthKeyNil      = errors.New("token: AuthKey was nil")

	PstLoc, _ = time.LoadLocation("America/Los_Angeles")
)

// AppleStatData 苹果数据汇总结果
type AppleStatData map[string](map[string]int)

// AppleClient 苹果接口结构
type AppleClient struct {
	KeyID    string
	IssID    string
	PrivKey  string
	VendorNO string
	JwtToken string
	// P8Path  string
}

// Init 获取jwt token等初始化操作
func (ac *AppleClient) Init() error {
	payload := jwt.StandardClaims{
		Audience:  "appstoreconnect-v1",
		Issuer:    ac.IssID,
		ExpiresAt: time.Now().Unix() + 600, // 10分钟有效期
	}

	token := jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": "ES256",
			"kid": ac.KeyID,
		},
		Claims: payload,
		Method: jwt.SigningMethodES256,
	}

	key, err := ParseP8PrivKey("", ac.PrivKey)
	if err != nil {
		log.Fatal("parse p8 priv key fail")
		return err
	}
	secretStr, err := token.SignedString(key)
	ac.JwtToken = secretStr
	return err
}

// New 生成appleClient
func (ac *AppleClient) New() *AppleClient {
	return nil
}

// ParseP8PrivKey 从p8文件读取私钥
func ParseP8PrivKey(path, txt string) (*ecdsa.PrivateKey, error) {
	var rawByte []byte
	if path == "" {
		rawByte = []byte(txt)
	} else {
		var err error
		rawByte, err = ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("获取密钥错误")
			return nil, err
		}
	}

	block, _ := pem.Decode(rawByte)
	if block == nil {
		log.Fatal("密钥文件不是pem格式")
		return nil, ErrAuthKeyNotPem
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pk := key.(type) {
	case *ecdsa.PrivateKey:
		return pk, nil
	default:
		return nil, ErrAuthKeyNotECDSA
	}
}

// SaveFile 保存苹果销售日报数据为文件
func SaveFile(filename string, txt []byte) {
	if !DLDirReady(DLDir) {
		return
	}
	ioutil.WriteFile(filename, txt, 0666)
}

// DLDirReady 判断所给路径文件/文件夹是否存在
func DLDirReady(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if err := os.Mkdir(DLDir, os.ModePerm); err != nil {
			log.Fatal("创建dl路径失败")
			return false
		}
		return true
	}
	return true
}

// ReqApple 请求苹果api接口
func ReqApple(url string, token string, paras map[string]string) ([]byte, error) {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("Accept", "application/a-gzip")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Authorization", "Bearer "+token)

	qs := req.URL.Query()
	for k, v := range paras {
		qs.Add(k, v)
	}

	req.URL.RawQuery = qs.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("http error")
		return []byte(""), err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		log.Fatalf("req not 200, is %d\n%s\n%v", resp.StatusCode, body, err)
		// log.Fatal(string(body))
		return []byte(""), errors.New("http not 200")
	}

	zr, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer zr.Close()
	ss, err := ioutil.ReadAll(zr)
	return ss, err
}

// ResetTZ 重置时间时区,当loc为nil时，采用pst时区
func ResetTZ(d time.Time, loc *time.Location) time.Time {
	rawstr := d.Format("2006-01-02")
	newDate, _ := time.ParseInLocation("2006-01-02", rawstr, PstLoc)
	// fmt.Println(d, newDate)
	return newDate
}
