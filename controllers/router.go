package controllers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo"
)

type RouterController struct{}

func (RouterController) Post(c echo.Context) error {
	method, _, err := routeParse(c)
	if err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}
	switch method {
	case "wechat.pay":
		return WxApiController{}.Pay(c)
	case "wechat.query":
		return WxApiController{}.Query(c)
	case "wechat.reverse":
		return WxApiController{}.Reverse(c)
	case "wechat.refund":
		return WxApiController{}.Refund(c)
	case "wechat.prepay":
		return WxApiController{}.Prepay(c)

	case "wechat.notify":
		return WxApiController{}.Notify(c)
	case "wechat.prepay.easy":
		return WxApiController{}.PrepayEasy(c)
	case "wechat.prepay.openid":
		return WxApiController{}.PrepayOpenId(c)

	case "alipay.pay":
		return AlApiController{}.Pay(c)
	case "alipay.query":
		return AlApiController{}.Query(c)
	case "alipay.reverse":
		return AlApiController{}.Reverse(c)
	case "alipay.refund":
		return AlApiController{}.Refund(c)
	case "alipay.prepay":
		return AlApiController{}.Prepay(c)

	case "alipay.notify":
		return AlApiController{}.Notify(c)

	default:
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, errors.New("param 'method' is not validate"))
	}
}

func (RouterController) Get(c echo.Context) error {
	state := c.QueryParam("state")
	switch state {
	case "wechat.prepay.openid":
		return WxApiController{}.PrepayOpenId(c)
	// case "fruit.getone":
	// 	return FruitApiController{}.GetOne(c)
	default:
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, errors.New("param 'state' is missing"))
	}
}
