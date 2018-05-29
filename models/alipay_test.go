package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/relax-space/go-kit/test"
)

func Test_AlAccount_Get(t *testing.T) {
	result, err := AlAccount{}.Get(ctx, 10001)
	test.Ok(t, err)
	fmt.Printf("%+v", result)
}

func Test_AlNotify_Get(t *testing.T) {
	result, err := AlNotify{}.Get(ctx, "2015081700219294", "131712072074192779392994717")
	test.Ok(t, err)
	fmt.Printf("%+v", result)
}

func Test_AlNotify_InsertOne(t *testing.T) {
	d := &AlNotify{
		AppId:      "Test",
		OutTradeNo: time.Now().Format("2006-01-02 15:04:05"),
	}
	err := d.InsertOne(ctx)
	test.Ok(t, err)
}
