package dao

import (
	"fmt"

	"github.com/btcsuite/goleveldb/leveldb"
)

var instance *DaoImpl

type DaoImpl struct {
	db   *leveldb.DB
	name string
}

func NewDao(dir string) {
	d := &DaoImpl{}
	d.name = dir
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("open leveldb %s\r\n", dir)
	d.db = db
	instance = d
}

func (d *DaoImpl) Name() string {
	return d.name
}

func SavePwd(typ byte, k, v []byte) error {
	key := formatKey(typ, k)
	return instance.db.Put(key, v, nil)
}

func GetPwd(typ byte, k []byte) ([]byte, error) {
	key := formatKey(typ, k)
	return instance.db.Get(key, nil)
}

func formatKey(typ byte, k []byte) []byte {
	key := make([]byte, 0)
	key = append(key, typ)
	key = append(key, k...)
	return key
}
