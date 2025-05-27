package utils

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
)

var Host = "smtp.qq.com"
var Port = 25
var UserName = "861214959@qq.com"
var Password = "wvrubjyqgtjjbfgd"
var d *gomail.Dialer

func init() {
	d = gomail.NewDialer(Host, Port, UserName, Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
}

func SendRegisterMail(user string, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", UserName)
	m.SetHeader("To", user)
	m.SetHeader("Subject", "GodQQ注册验证码")
	m.SetBody("text/html", fmt.Sprintf("您的验证码为：%s。<br>过期时间:五分钟<br>", code))
	err := d.DialAndSend(m)
	return err
}
