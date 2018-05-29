package controllers

import (
	"context"
	"errors"
	"io/ioutil"
	"ipay-api/factory"
	"ipay-api/models"
	"net/http"
	"strconv"

	"github.com/fatih/structs"
	"github.com/labstack/echo"
	"github.com/relax-space/go-kit/base"
	"github.com/relax-space/go-kit/sign"
	alsdk "github.com/relax-space/lemon-alipay-sdk"
)

type AlApiController struct {
}

func (AlApiController) Pay(c echo.Context) error {
	reqDto := AliReqPayDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}

	account, err := models.AlAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &alsdk.ReqBaseDto{
		AppId:        account.AppId,
		AppAuthToken: account.AuthToken,
	}
	if len(account.SysServiceProviderId) != 0 {
		reqDto.ExtendParams = &alsdk.ExtendParams{
			SysServiceProviderId: account.SysServiceProviderId,
		}
	}
	customDto := &alsdk.ReqCustomerDto{
		PriKey: account.PriKey,
		PubKey: account.PubKey,
	}

	statusCode, _, result, err := alsdk.Pay(reqDto.ReqPayDto, customDto)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Alipay",
		Method:         "POST",
		Uri:            alsdk.OPENAPIURL,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqPayDto),
		Extra:          "Pay",
		Err:            err,
	})
	if err != nil {
		if err.Error() == alsdk.MESSAGE_PAYING {
			outTradeNo := result.OutTradeNo
			queryDto := alsdk.ReqQueryDto{
				ReqBaseDto: reqDto.ReqBaseDto,
				OutTradeNo: outTradeNo,
			}
			statusCode, _, result, err = alsdk.LoopQuery(&queryDto, customDto, 40, 2)
			PrintApiBehaviorError(c.Request().Context(), UrlInfo{
				ApiName:        "Alipay",
				Method:         "POST",
				Uri:            alsdk.OPENAPIURL,
				ResponseStatus: statusCode,
				Struct:         structs.New(&queryDto),
				Extra:          "Query",
				Err:            err,
			})
			if err == nil {
				return ReturnApiSucc(c, http.StatusOK, result)
			} else {
				reverseDto := alsdk.ReqReverseDto{
					ReqBaseDto: reqDto.ReqBaseDto,
					OutTradeNo: outTradeNo,
				}
				statusCode, _, _, err = alsdk.Reverse(&reverseDto, customDto, 10, 10)
				PrintApiBehaviorError(c.Request().Context(), UrlInfo{
					ApiName:        "Alipay",
					Method:         "POST",
					Uri:            alsdk.OPENAPIURL,
					ResponseStatus: statusCode,
					Struct:         structs.New(&reverseDto),
					Extra:          "Reverse",
					Err:            err,
				})
				if err != nil {
					return ReturnApiFail(c, http.StatusOK, ApiErrorAlipay, err)
				} else {
					return ReturnApiFail(c, http.StatusOK, ApiErrorAlipay, errors.New("reverse success"))
				}
			}
		} else {
			return ReturnApiFail(c, http.StatusOK, ApiErrorAlipay, err)
		}
	}
	return ReturnApiSucc(c, http.StatusOK, result)
}

func (d AlApiController) Query(c echo.Context) error {
	reqDto := AliReqQueryDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}

	account, err := models.AlAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	result, err := d.queryByAccount(c.Request().Context(), account, reqDto.OutTradeNo)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorAlipay, err)
	}
	return ReturnApiSucc(c, http.StatusOK, result)
}
func (AlApiController) Refund(c echo.Context) error {
	reqDto := AliReqRefundDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}
	account, err := models.AlAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &alsdk.ReqBaseDto{
		AppId:        account.AppId,
		AppAuthToken: account.AuthToken,
	}

	customDto := &alsdk.ReqCustomerDto{
		PriKey: account.PriKey,
		PubKey: account.PubKey,
	}
	statusCode, _, result, err := alsdk.Refund(reqDto.ReqRefundDto, customDto)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Alipay",
		Method:         "POST",
		Uri:            alsdk.OPENAPIURL,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqRefundDto),
		Extra:          "Refund",
		Err:            err,
	})
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorAlipay, err)
	}
	return ReturnApiSucc(c, http.StatusOK, result)

}
func (AlApiController) Reverse(c echo.Context) error {
	reqDto := AliReqReverseDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}
	account, err := models.AlAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &alsdk.ReqBaseDto{
		AppId:        account.AppId,
		AppAuthToken: account.AuthToken,
	}

	customDto := &alsdk.ReqCustomerDto{
		PriKey: account.PriKey,
		PubKey: account.PubKey,
	}
	statusCode, _, result, err := alsdk.Reverse(reqDto.ReqReverseDto, customDto, 10, 10)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Alipay",
		Method:         "POST",
		Uri:            alsdk.OPENAPIURL,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqReverseDto),
		Extra:          "Reverse",
		Err:            err,
	})
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorAlipay, err)

	}
	return ReturnApiSucc(c, http.StatusOK, result)
}

func (AlApiController) Prepay(c echo.Context) error {
	reqDto := AliReqPrepayDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}

	account, err := models.AlAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &alsdk.ReqBaseDto{
		AppId:        account.AppId,
		AppAuthToken: account.AuthToken,
	}
	if len(account.SysServiceProviderId) != 0 {
		reqDto.ExtendParams = &alsdk.ExtendParams{
			SysServiceProviderId: account.SysServiceProviderId,
		}
	}
	customDto := &alsdk.ReqCustomerDto{
		PriKey: account.PriKey,
		PubKey: account.PubKey,
	}
	statusCode, _, result, err := alsdk.Prepay(reqDto.ReqPrepayDto, customDto)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Alipay",
		Method:         "POST",
		Uri:            alsdk.OPENAPIURL,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqPrepayDto),
		Extra:          "Prepay",
		Err:            err,
	})
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorAlipay, err)
	}
	return ReturnApiSucc(c, http.StatusOK, result)
}

func (d AlApiController) Notify(c echo.Context) error {
	sbody, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return ReturnApiAliPushFail(c, http.StatusBadRequest, err)
	}
	formParam := string(sbody)
	if len(formParam) == 0 {
		return ReturnApiAliPushFail(c, http.StatusBadRequest, errors.New("request param is required"))
	}
	var reqDto models.AlNotify
	mapParam := base.ParseMapObjectEncode(formParam, "&", "=")
	err = Decode(mapParam, &reqDto)
	if err != nil {
		return ReturnApiAliPushFail(c, http.StatusBadRequest, err)
	}

	//1.validate
	if err = d.notifyValid(c.Request().Context(), reqDto.Body, reqDto.Sign, reqDto.OutTradeNo, reqDto.TotalAmount, mapParam); err != nil {
		return ReturnApiAliPushFail(c, http.StatusOK, err)
	}

	//2.save notify info
	err = (&reqDto).InsertOne(c.Request().Context())
	if err != nil {
		return ReturnApiAliPushFail(c, http.StatusOK, err)
	}
	return c.String(http.StatusOK, "success")
}

func (d AlApiController) notifyValid(ctx context.Context, body, signParam, outTradeNo, totalAmount string, mapParam map[string]interface{}) (err error) {

	//0.get account info
	bodyMap := base.ParseMapObject(body, factory.ConfigString(ctx, "NOTIFY_BODY_SEP1"), factory.ConfigString(ctx, "NOTIFY_BODY_SEP2"))
	var eId int64
	var flag bool
	if eIdObj, ok := bodyMap["e_id"]; ok {
		if eId, err = strconv.ParseInt(eIdObj.(string), 10, 64); err == nil {
			flag = true
		}
	}
	if !flag {
		err = errors.New("e_id(int64) is not existed in param(param name:body) or format is not correct")
		return
	}

	account, err := models.AlAccount{}.Get(ctx, eId)
	if err != nil {
		return
	}

	//1.valid sign
	signStr := signParam
	delete(mapParam, "sign")
	delete(mapParam, "sign_type")

	if !sign.CheckSha1Sign(base.JoinMapObject(mapParam), signStr, account.PubKey) {
		err = errors.New("sign valid failure")
		return
	}

	//2.valid data
	queryDto, err := d.queryByAccount(ctx, account, outTradeNo)
	if err != nil {
		return
	}
	if !(queryDto.TotalAmount == totalAmount) {
		err = errors.New("request amount valid failure")
		return
	}
	return
}

func (AlApiController) queryByAccount(ctx context.Context, account *models.AlAccount, outTradeNo string) (result *alsdk.RespQueryDto, err error) {
	var reqDto alsdk.ReqQueryDto
	reqDto.ReqBaseDto = &alsdk.ReqBaseDto{
		AppId:        account.AppId,
		AppAuthToken: account.AuthToken,
	}

	customDto := &alsdk.ReqCustomerDto{
		PriKey: account.PriKey,
		PubKey: account.PubKey,
	}
	reqDto.OutTradeNo = outTradeNo
	statusCode, _, result, err := alsdk.Query(&reqDto, customDto)
	PrintApiBehaviorError(ctx, UrlInfo{
		ApiName:        "Alipay",
		Method:         "POST",
		Uri:            alsdk.OPENAPIURL,
		ResponseStatus: statusCode,
		Struct:         structs.New(&reqDto),
		Extra:          "Query",
		Err:            err,
	})
	return
}
