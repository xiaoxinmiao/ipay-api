package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/relax-space/go-kit/test"
)

func Test_WxAccount_Get(t *testing.T) {
	result, err := WxAccount{}.Get(ctx, 10001)
	test.Ok(t, err)
	fmt.Printf("%+v", result)
}

func Test_WxNotify_Get(t *testing.T) {
	result, err := WxNotify{}.Get(ctx, "wx2421b1c4370ec43b",
		"10000100", "1409811653")
	test.Ok(t, err)
	fmt.Printf("%+v", result)
}

func Test_WxNotify_InsertOne(t *testing.T) {
	d := &WxNotify{
		AppId:      "Test",
		MchId:      time.Now().Format("2006-01-02 15:04:05"),
		OutTradeNo: "test",
	}
	err := d.InsertOne(ctx)
	test.Ok(t, err)
}
