package models

import (
	"context"
	"fmt"
	"os"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/pangpanglabs/goutils/echomiddleware"
)

var ctx context.Context

func init() {
	runtime.GOMAXPROCS(1)
	fmt.Println(os.Getenv("IPAY_CONN"))
	xormEngine, err := xorm.NewEngine("mysql", os.Getenv("IPAY_CONN"))

	if err != nil {
		panic(err)
	}
	xormEngine.ShowSQL(true)
	ctx = context.WithValue(context.Background(), echomiddleware.ContextDBName, xormEngine.NewSession())
}
