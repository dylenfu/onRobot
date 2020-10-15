package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

type Config struct {
	Ip   string
	Port uint16
	User string
	Pwd  string
	Db   string
}

func NewDao(c *Config) {
	if db != nil {
		return
	}

	var err error

	url := fmt.Sprintf("%s:%s@(%s:%d)/%s?", c.User, c.Pwd, c.Ip, c.Port, c.Db)
	if db, err = gorm.Open("mysql", url); err != nil {
		panic("failed to connect database")
	}

	var initTable = func(tb interface{}) {
		// db.DropTableIfExists(tb)
		// db.CreateTable(tb)
		db.AutoMigrate(tb)
	}

	initTable(&Stat{})
	initTable(&Subnet{})
}

type Stat struct {
	Ip   string
	Port uint16
	Hash string
	Send uint64
	Recv uint64
}

func InsertStat(ip string, port uint16, hash string, send, recv uint64) error {
	return db.Create(&Stat{Ip: ip, Port: port, Hash: hash, Send: send, Recv: recv}).Error
}

type Subnet struct {
	Pubkey string
}

func CheckSubnet(id string) bool {
	data := new(Subnet)
	affected := db.Where("pubkey = ?", id).First(data).RowsAffected
	return affected > 0
}

func GetAllSubnet() []string {
	list := []string{}
	ids := []*Subnet{}
	db.Find(&ids)
	for _, v := range ids {
		list = append(list, v.Pubkey)
	}
	return list
}
