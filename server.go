package main

import (
	"FunGoLoginCenter/database"
	"FunGoLoginCenter/utils"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/badoux/checkmail"
	"github.com/gin-gonic/gin"
	ZSC "github.com/zuishabi/ServiceCenter/src"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LoginInfo struct {
	Email    string `json:"email"`
	Password string `json:"pwd"`
	Service  string `json:"service"` // 记录需要登录哪个服务器，SharGo 或 FunGo
}

type RegisterInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserName string `json:"user_name"`
	Code     string `json:"code"`
}

var ServiceClient *ZSC.Client

func main() {
	if err := database.Start(); err != nil {
		panic(err)
	}
	database.Update()
	initBF()
	initServiceCenterClient()
	r := gin.Default()
	r.POST("/login", login)
	r.POST("/register", register)
	r.GET("/mail_code", getMailCode)
	r.Run(":8888")
}

func login(c *gin.Context) {
	loginInfo := LoginInfo{}
	if err := c.BindJSON(&loginInfo); err != nil {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}
	//计算密码的md5值
	md5Password := utils.StringMD5(loginInfo.Password)
	userInfo := database.UserInfo{}
	database.Db.Where("user_email = ?", loginInfo.Email).Where("password = ?", md5Password).First(&userInfo)
	if userInfo.UserName == "" {
		c.JSON(http.StatusOK, gin.H{"status": "用户名账号或密码错误"})
		return
	}
	//计算当前用户的邮箱的md5
	hash := md5.New()
	_, _ = io.WriteString(hash, userInfo.UserEmail)
	key := hex.EncodeToString(hash.Sum(nil))
	//database.RDB.SetEx(context.Background(), key, userInfo.UID, 10*time.Second)
	// 将用户的信息发送到valkey中
	if err := database.Client.Do(
		context.Background(),
		//database.Client.B().Setex().Key(key).Seconds(10).Value(strconv.Itoa(int(userInfo.UID))).Build(),
		database.Client.B().Hmset().Key(key).FieldValue().
			FieldValue("uid", strconv.Itoa(int(userInfo.UID))).FieldValue("name", userInfo.UserName).Build(),
	).Error(); err != nil {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}
	// 设置过期时间
	if err := database.Client.Do(
		context.Background(),
		database.Client.B().Expire().Key(key).Seconds(10).Build(),
	).Error(); err != nil {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}
	if loginInfo.Service == "FunGo" && FunGo.Status == 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": "login successful",
			"ip":     FunGo.IP,
			"port":   FunGo.Port,
			"key":    key,
		})
	} else if loginInfo.Service == "SharGo" && SharGo.Status == 1 {
		c.JSON(http.StatusOK, gin.H{
			"status": "login successful",
			"ip":     SharGo.IP,
			"port":   SharGo.Port,
			"key":    key,
		})
	} else {
		c.JSON(http.StatusForbidden, gin.H{})
	}
}

func register(c *gin.Context) {
	//检查验证码是否存在
	registerInfo := RegisterInfo{}
	if err := c.BindJSON(&registerInfo); err != nil {
		return
	}
	//s := database.RDB.Get(context.Background(), "code_"+registerInfo.Email)
	s, err := database.Client.Do(context.Background(), database.Client.B().Get().Key("code_"+registerInfo.Email).Build()).ToString()
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"success": 0,
			"msg":     "邮箱错误",
		})
		return
	}
	if s != registerInfo.Code {
		fmt.Println(registerInfo.Code, " ", s)
		c.JSON(200, gin.H{
			"success": 0,
			"msg":     "验证码错误",
		})
		return
	}
	//检查用户名是否被使用
	userInfo := database.UserInfo{}
	if UserNameBF.GetItem(registerInfo.UserName) {
		// 名称可能重复
		database.Db.Where("user_name = ?", registerInfo.UserName).First(&userInfo)
		if userInfo.UID != 0 {
			c.JSON(200, gin.H{
				"success": 0,
				"msg":     "用户名重复",
			})
			return
		}
	}
	//删除原来的验证码
	//database.RDB.Del(context.Background(), "code_"+registerInfo.Email)
	database.Client.Do(context.Background(), database.Client.B().Del().Key("code_"+registerInfo.Email).Build())
	userInfo.UserEmail = registerInfo.Email
	userInfo.UserName = registerInfo.UserName
	userInfo.Password = utils.StringMD5(registerInfo.Password)
	database.Db.Create(&userInfo)
	c.JSON(200, gin.H{
		"success": 1,
		"msg":     "注册成功",
	})
}

func initServiceCenterClient() {
	interestingServices := []string{"SharGo", "FunGo"}
	ServiceClient = ZSC.NewClient("127.0.0.1", 8888, "LoginCenter", interestingServices, "service-center:9999")
	if err := ServiceClient.Start(); err != nil {
		panic(err)
	}
	ServiceClient.SetTimeoutTime(10000)
	ServiceClient.RegisterOnlineFunc(onServiceConnect)
	ServiceClient.RegisterOfflineFunc(onServiceDisconnect)
	time.Sleep(50 * time.Millisecond)
	status, err := ServiceClient.GetServiceStatus("FunGo")
	if err != nil {
		fmt.Println("get MainServer status error,err = ", err)
		return
	}
	FunGo = status
	fmt.Println(FunGo)
	status, err = ServiceClient.GetServiceStatus("SharGo")
	if err != nil {
		fmt.Println("get MainServer status error,err = ", err)
		return
	}
	SharGo = status
	fmt.Println(SharGo)
}

var FunGo ZSC.ServiceStatus
var SharGo ZSC.ServiceStatus

func onServiceConnect(info *ZSC.ServiceStatus) {
	if info.Name == "FunGo" {
		FunGo = *info
	} else if info.Name == "SharGo" {
		SharGo = *info
	} else {
		fmt.Println("not interesting service connect : ", info.Name)
	}
}

func onServiceDisconnect(name string) {
	if name == "FunGo" {
		FunGo.Status = 2
	} else if name == "SharGo" {
		SharGo.Status = 2
	} else {
		fmt.Println("not interesting service disconnect : ", name)
	}
}

func getMailCode(c *gin.Context) {
	email := c.Query("email")
	err := checkmail.ValidateFormat(email)
	if err != nil {
		c.JSON(200, gin.H{
			"success": 0,
			"msg":     "邮件格式错误",
		})
		return
	}
	err = checkmail.ValidateHost(email)
	if err != nil {
		c.JSON(200, gin.H{
			"success": 0,
			"msg":     "无法解析的邮箱",
		})
		return
	}
	//检查邮箱是否被注册
	userInfo := database.UserInfo{}
	if UserEmailBF.GetItem(email) {
		// 当前邮箱可能被注册了
		database.Db.Where("user_email = ?", email).First(&userInfo)
		if userInfo.UserEmail != "" {
			c.JSON(http.StatusOK, gin.H{
				"success": 0,
				"msg":     "当前邮箱已被注册",
			})
			return
		}
	}
	//生成一个验证码，同时存储进redis中
	code := createCode()
	//s := database.RDB.SetEx(context.Background(), "code_"+email, code, 300*time.Second)
	err = database.Client.Do(
		context.Background(),
		database.Client.B().Setex().Key("code_"+email).Seconds(300).Value(code).Build(),
	).Error()
	if err != nil {
		fmt.Println("send register email err = ", err)
		c.JSON(http.StatusOK, gin.H{
			"success": 0,
			"msg":     "邮件发送错误: " + err.Error(),
		})
		return
	}
	if err := utils.SendRegisterMail(email, code); err != nil {
		fmt.Println("send register email err = ", err)
		c.JSON(http.StatusOK, gin.H{
			"success": 0,
			"msg":     "邮件发送错误: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": 1,
		"msg":     "邮件发送成功",
	})
}

var codeLen = 6
var codeList = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

// 生成一个验证码
func createCode() string {
	r := len(codeList)
	var sb strings.Builder
	for i := 0; i < codeLen; i++ {
		_, _ = fmt.Fprintf(&sb, "%d", codeList[rand.Intn(r)])
	}
	return sb.String()
}

var UserEmailBF *utils.BloomFilter
var UserNameBF *utils.BloomFilter

// 创建名称和email的布隆过滤器
func initBF() {
	UserEmailBF = utils.InitBF(1000)
	UserNameBF = utils.InitBF(1000)
	offset := 0
	users := make([]database.UserInfo, 5)
	for len(users) != 0 {
		database.Db.Offset(offset).Limit(5).Find(&users)
		offset += len(users)
		for _, v := range users {
			UserNameBF.SetItem(v.UserName)
			UserEmailBF.SetItem(v.UserEmail)
		}
	}
}
