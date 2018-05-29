package controllers

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"ipay-api/factory"
	"ipay-api/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/relax-space/go-kitt/auth"

	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/relax-space/go-kitt/random"

	"github.com/relax-space/lemon-wxmp-sdk/mpAuth"

	"github.com/relax-space/go-kit/base"
	"github.com/relax-space/go-kit/data"
	"github.com/relax-space/go-kit/sign"

	wxsdk "github.com/relax-space/lemon-wxpay-sdk"

	"github.com/labstack/echo"
)

type WxApiController struct {
}

func (WxApiController) Pay(c echo.Context) error {

	reqDto := WxReqPayDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}
	account, err := models.WxAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &wxsdk.ReqBaseDto{
		AppId:    account.AppId,
		SubAppId: account.SubAppId,
		MchId:    account.MchId,
		SubMchId: account.SubMchId,
	}
	customDto := wxsdk.ReqCustomerDto{
		Key: account.Key,
	}

	statusCode, _, result, err := wxsdk.Pay(reqDto.ReqPayDto, &customDto)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Wechat",
		Method:         "POST",
		Uri:            wxsdk.URLPAY,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqPayDto),
		Extra:          "Pay",
		Err:            err,
	})
	if err != nil {
		if err.Error() == wxsdk.MESSAGE_PAYING {
			outTradeNo := result["out_trade_no"].(string)
			queryDto := wxsdk.ReqQueryDto{
				ReqBaseDto: reqDto.ReqBaseDto,
				OutTradeNo: outTradeNo,
			}
			statusCode, _, result, err = wxsdk.LoopQuery(&queryDto, &customDto, 40, 2)
			PrintApiBehaviorError(c.Request().Context(), UrlInfo{
				ApiName:        "Wechat",
				Method:         "POST",
				Uri:            wxsdk.URLQUERY,
				ResponseStatus: statusCode,
				Struct:         structs.New(queryDto),
				Extra:          "Query",
				Err:            err,
			})
			if err == nil {
				return ReturnApiSucc(c, http.StatusOK, result)
			} else {
				reverseDto := wxsdk.ReqReverseDto{
					ReqBaseDto: reqDto.ReqBaseDto,
					OutTradeNo: outTradeNo,
				}
				statusCode, _, _, err = wxsdk.Reverse(&reverseDto, &customDto, 10, 10)
				PrintApiBehaviorError(c.Request().Context(), UrlInfo{
					ApiName:        "Wechat",
					Method:         "POST",
					Uri:            wxsdk.URLREVERSE,
					ResponseStatus: statusCode,
					Struct:         structs.New(&reverseDto),
					Extra:          "Reverse",
					Err:            err,
				})
				if err != nil {
					return ReturnApiFail(c, http.StatusOK, ApiErrorWechat, err)
				} else {
					return ReturnApiFail(c, http.StatusOK, ApiErrorWechat, errors.New("reverse success"))
				}
			}
		} else {
			return ReturnApiFail(c, http.StatusOK, ApiErrorWechat, err)
		}
	}
	return ReturnApiSucc(c, http.StatusOK, result)
}

func (d WxApiController) Query(c echo.Context) error {
	reqDto := WxReqQueryDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}

	account, err := models.WxAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	result, err := d.queryByAccount(c.Request().Context(), account, reqDto.OutTradeNo)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorWechat, err)
	}
	return ReturnApiSucc(c, http.StatusOK, result)
}
func (WxApiController) Refund(c echo.Context) error {
	reqDto := WxReqRefundDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}
	account, err := models.WxAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &wxsdk.ReqBaseDto{
		AppId:    account.AppId,
		SubAppId: account.SubAppId,
		MchId:    account.MchId,
		SubMchId: account.SubMchId,
	}
	custDto := wxsdk.ReqCustomerDto{
		Key:          account.Key,
		CertPathName: account.CertName,
		CertPathKey:  account.CertKey,
		RootCa:       account.RootCa,
	}
	statusCode, _, result, err := wxsdk.Refund(reqDto.ReqRefundDto, &custDto)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Wechat",
		Method:         "POST",
		Uri:            wxsdk.URLREFUND,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqRefundDto),
		Extra:          "Refund",
		Err:            err,
	})
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorWechat, err)
	}
	return ReturnApiSucc(c, http.StatusOK, result)

}
func (WxApiController) Reverse(c echo.Context) error {
	reqDto := WxReqReverseDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}
	account, err := models.WxAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &wxsdk.ReqBaseDto{
		AppId:    account.AppId,
		SubAppId: account.SubAppId,
		MchId:    account.MchId,
		SubMchId: account.SubMchId,
	}
	custDto := wxsdk.ReqCustomerDto{
		Key:          account.Key,
		CertPathName: account.CertName,
		CertPathKey:  account.CertKey,
		RootCa:       account.RootCa,
	}
	statusCode, _, result, err := wxsdk.Reverse(reqDto.ReqReverseDto, &custDto, 10, 10)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Wechat",
		Method:         "POST",
		Uri:            wxsdk.URLREVERSE,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqReverseDto),
		Extra:          "Reverse",
		Err:            err,
	})
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorWechat, err)
	}
	return ReturnApiSucc(c, http.StatusOK, result)
}

func (WxApiController) RefundQuery(c echo.Context) error {
	reqDto := WxReqRefundQueryDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}

	account, err := models.WxAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &wxsdk.ReqBaseDto{
		AppId:    account.AppId,
		SubAppId: account.SubAppId,
		MchId:    account.MchId,
		SubMchId: account.SubMchId,
	}
	customDto := wxsdk.ReqCustomerDto{
		Key: account.Key,
	}
	statusCode, _, result, err := wxsdk.RefundQuery(reqDto.ReqRefundQueryDto, &customDto)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Wechat",
		Method:         "POST",
		Uri:            wxsdk.URLREFUNDQUERY,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqRefundQueryDto),
		Extra:          "RefundQuery",
		Err:            err,
	})
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorWechat, err)
	}
	return ReturnApiSucc(c, http.StatusOK, result)
}

func (d WxApiController) Prepay(c echo.Context) error {
	reqDto := WxReqPrepayDto{}
	if err := c.Bind(&reqDto); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}

	account, err := models.WxAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorDB, err)
	}
	reqDto.ReqBaseDto = &wxsdk.ReqBaseDto{
		AppId:    account.AppId,
		SubAppId: account.SubAppId,
		MchId:    account.MchId,
		SubMchId: account.SubMchId,
	}
	reqDto.TimeStart = ChinaDatetime().Format("20060102150405")
	reqDto.TimeExpire = ChinaDatetime().Add(10 * time.Minute).Format("20060102150405")
	customDto := wxsdk.ReqCustomerDto{
		Key: account.Key,
	}
	/*
		1.set customer notify_url into attach
		2.set ipay united notify_url to reqDto
	*/
	reqDto.ReqPrepayDto.Attach = d.setNotifyAttach(c.Request().Context(), reqDto.NotifyUrl, reqDto.Attach, reqDto.EId)
	reqDto.ReqPrepayDto.NotifyUrl = fmt.Sprintf("%v/%v", factory.ConfigString(c.Request().Context(), "IPAY_HOST"), "notify")

	reqDto.ReqPrepayDto.TimeStart = ChinaDatetime().Format("20060102150405")
	reqDto.ReqPrepayDto.TimeExpire = ChinaDatetime().Add(10 * time.Minute).Format("20060102150405")
	statusCode, _, result, err := wxsdk.Prepay(reqDto.ReqPrepayDto, &customDto)
	PrintApiBehaviorError(c.Request().Context(), UrlInfo{
		ApiName:        "Wechat",
		Method:         "POST",
		Uri:            wxsdk.URLPREPAY,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqPrepayDto),
		Extra:          "Prepay",
		Err:            err,
	})
	if err != nil {
		return ReturnApiFail(c, http.StatusOK, ApiErrorWechat, err)
	}

	prePayParam := make(map[string]interface{}, 0)
	prePayParam["package"] = "prepay_id=" + base.ToString(result["prepay_id"])
	prePayParam["timeStamp"] = base.ToString(ChinaDatetime().Unix())
	prePayParam["nonceStr"] = result["nonce_str"]
	prePayParam["signType"] = "MD5"
	prePayParam["appId"] = result["appid"]
	prePayParam["paySign"] = sign.MakeMd5Sign(base.JoinMapObject(prePayParam), account.Key)

	return ReturnApiSucc(c, http.StatusOK, prePayParam)
}

func (d WxApiController) Notify(c echo.Context) error {

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return ReturnApiWxPushFail(c, http.StatusBadRequest, err)
	}
	xmlBody := string(body)
	if len(xmlBody) == 0 {
		return ReturnApiWxPushFail(c, http.StatusBadRequest, errors.New("xml is empty"))
	}
	fmt.Println(string(xmlBody))
	//1.get dto data
	var notifyDto models.WxNotify
	err = xml.Unmarshal([]byte(xmlBody), &notifyDto)
	if err != nil {
		return ReturnApiWxPushFail(c, http.StatusBadRequest, err)
	}
	//1.1 get mapData
	wxData := data.New()
	err = wxData.FromXml(xmlBody)
	if err != nil {
		return ReturnApiWxPushFail(c, http.StatusBadRequest, err)
	}
	//2.valid
	if err = d.notifyValid(c.Request().Context(), notifyDto.Attach, notifyDto.Sign, notifyDto.OutTradeNo, notifyDto.TotalFee, wxData); err != nil {
		return ReturnApiWxPushFail(c, http.StatusOK, err)
	}

	//3.save into data base
	err = (&notifyDto).InsertOne(c.Request().Context())
	if err != nil {
		return ReturnApiWxPushFail(c, http.StatusOK, err)
	}

	successResult := struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
	}{xml.Name{}, "SUCCESS", "OK"}
	return c.XML(http.StatusOK, successResult)
}

const (
	IPAY_WECHAT_PREPAY       = "IPAY_WECHAT_PREPAY"
	IPAY_WECHAT_PREPAY_INNER = "IPAY_WECHAT_PREPAY_INNER"
	IPAY_WECHAT_PREPAY_ERROR = "IPAY_WECHAT_PREPAY_ERROR"
)

/*
PrepayEasy Part1: redirect to "https:/xxxx/v3/prepayopenid" for get openid
*/
func (d WxApiController) PrepayEasy(c echo.Context) error {

	prepay_param := c.QueryParam("prepay_param")

	reqDto := WxReqPrepayEasyDto{}
	err := json.Unmarshal([]byte(prepay_param), &reqDto)
	if err != nil {
		d.prepayErrCookie(c, err)
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}
	urlStr, err := d.prepayPageUrl(reqDto.PageUrl)
	if err != nil {
		d.prepayErrCookie(c, err)
		return ReturnApiFail(c, http.StatusBadRequest, ApiErrorParameter, err)
	}
	reqDto.PageUrl = fmt.Sprintf(urlStr, random.Uuid("")) + "?" + random.Uuid("")
	account, err := models.WxAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return d.returnPrepayFailAndRedirect(c, reqDto.PageUrl, err)
	}

	openIdUrlParam := &mpAuth.ReqDto{
		AppId:       account.AppId,
		State:       "state",
		RedirectUrl: fmt.Sprintf("%v?state=%v", factory.ConfigString(c.Request().Context(), "IPAY_HOST"), "wechat.prepay.openid"),
		PageUrl:     reqDto.PageUrl,
	}
	SetCookie(IPAY_WECHAT_PREPAY_INNER, prepay_param, c)
	SetCookie(IPAY_WECHAT_PREPAY, "", c)
	SetCookie(IPAY_WECHAT_PREPAY_ERROR, "", c)
	return c.Redirect(http.StatusFound, mpAuth.GetUrlForAccessToken(openIdUrlParam))
}

/*
PrepayEasy Part2:
	1.redirect back to origin request url
*/
func (d WxApiController) PrepayOpenId(c echo.Context) error {
	code := c.QueryParam("code")
	reqUrl := c.QueryParam("reurl")
	reqDto, err := d.prepayReqParam(c)
	if err != nil {
		return d.returnPrepayFailAndRedirect(c, reqUrl, err)
	}
	//1.get account
	account, err := models.WxAccount{}.Get(c.Request().Context(), reqDto.EId)
	if err != nil {
		return d.returnPrepayFailAndRedirect(c, reqUrl, err)
	}
	//2.get openId
	respDto, err := mpAuth.GetAccessTokenAndOpenId(code, account.AppId, account.Secret)
	if err != nil {
		return d.returnPrepayFailAndRedirect(c, reqUrl, err)
	}
	reqDto.OpenId = respDto.OpenId
	//3.get prepay param
	prePayParam, err := d.prepayRespParam(c.Request().Context(), reqDto, account)
	if err != nil {
		return d.returnPrepayFailAndRedirect(c, reqUrl, err)
	}
	SetCookieObj(IPAY_WECHAT_PREPAY, prePayParam, c)
	SetCookie(IPAY_WECHAT_PREPAY_ERROR, "", c)
	SetCookie(IPAY_WECHAT_PREPAY_INNER, "", c)
	return c.Redirect(http.StatusFound, reqUrl)
}

func (WxApiController) queryByAccount(ctx context.Context, account *models.WxAccount, outTradeNo string) (result map[string]interface{}, err error) {
	var reqDto wxsdk.ReqQueryDto
	reqDto.ReqBaseDto = &wxsdk.ReqBaseDto{
		AppId:    account.AppId,
		SubAppId: account.SubAppId,
		MchId:    account.MchId,
		SubMchId: account.SubMchId,
	}
	customDto := &wxsdk.ReqCustomerDto{
		Key: account.Key,
	}
	reqDto.OutTradeNo = outTradeNo
	statusCode, _, result, err := wxsdk.Query(&reqDto, customDto)
	PrintApiBehaviorError(ctx, UrlInfo{
		ApiName:        "Wechat",
		Method:         "POST",
		Uri:            wxsdk.URLQUERY,
		ResponseStatus: statusCode,
		Struct:         structs.New(&reqDto),
		Extra:          "Query",
		Err:            err,
	})
	return
}

func (WxApiController) notifyBodyParse(ctx context.Context, body string) (bodyMap map[string]interface{}, eId int64, err error) {
	bodyMap = base.ParseMapObject(body, factory.ConfigString(ctx, "NOTIFY_BODY_SEP1"), factory.ConfigString(ctx, "NOTIFY_BODY_SEP2"))
	err = errors.New("e_id is not existed in body or format is not correct")
	eIdObj, ok := bodyMap["e_id"]
	if !ok {
		return
	}
	if eId, err = strconv.ParseInt(eIdObj.(string), 10, 64); err != nil {
		return
	}
	err = nil
	return
}

func (d WxApiController) notifyValid(ctx context.Context, body, signParam, outTradeNo string, totalAmount int64, dataParam *data.Data) (err error) {
	subNotifyUrl, rawAttach, eId, err := d.getNotifyAttach(ctx, body)
	if err != nil {
		return
	}
	account, err := models.WxAccount{}.Get(ctx, eId)
	if err != nil {
		return
	}

	//1.valid sign
	signStr := signParam
	mapParam := dataParam.DataAttr
	delete(mapParam, "sign")
	if !sign.CheckMd5Sign(base.JoinMapObject(mapParam), account.Key, signStr) {
		err = errors.New("sign valid failure")
		return
	}
	mapParam["attach"] = rawAttach
	//2.valid data
	queryMap, err := d.queryByAccount(ctx, account, outTradeNo)
	if err != nil {
		return
	}
	if !(queryMap["total_fee"].(string) == base.ToString(totalAmount)) {
		err = errors.New("amount is exception")
		return
	}
	mapParam["sign"] = signParam
	//3.send data to sub_mch
	if len(subNotifyUrl) != 0 {
		go func(signParam, subNotifyUrl string, dataParam *data.Data) {
			d.subNotify(subNotifyUrl, dataParam.ToXml())
		}(signStr, subNotifyUrl, dataParam)
	}
	return
}

func (WxApiController) subNotify(subNotifyUrl, xmlParam string) (result interface{}) {
	var successResult struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
	}
	token, err := getToken()
	if err != nil {
		return
	}
	resp, err := POSTXml(token, subNotifyUrl, xmlParam, &successResult)
	result = successResult
	if err == nil && resp != nil &&
		resp.StatusCode == http.StatusOK && successResult.ReturnCode == "SUCCESS" {
		return
	}
	return
}

func (d WxApiController) returnPrepayFailAndRedirect(c echo.Context, reqUrl string, err error) error {
	logContext := behaviorlog.FromCtx(c.Request().Context())
	if logContext != nil {
		if err != nil {
			logContext.WithError(err)
		}
	}
	if err != nil {
		d.prepayErrCookie(c, err)
	}
	return c.Redirect(http.StatusFound, reqUrl)
}

func (WxApiController) prepayErrCookie(c echo.Context, err error) {
	if err != nil {
		SetCookie(IPAY_WECHAT_PREPAY_ERROR, err.Error(), c)
	}
	SetCookie(IPAY_WECHAT_PREPAY_INNER, "", c)
	SetCookie(IPAY_WECHAT_PREPAY, "", c)
}
func (WxApiController) prepayPageUrl(pageUrl string) (result string, err error) {
	result, err = url.QueryUnescape(pageUrl)
	if err != nil {
		return
	}
	if len(result) == 0 {
		err = errors.New("page_url miss")
		return
	}
	indexTag := strings.Index(result, "#")
	result = result[0:indexTag] + "%v?" + result[indexTag:]
	return
}

func (WxApiController) prepayReqParam(c echo.Context) (reqDto *WxReqPrepayEasyDto, err error) {
	cookie, err := c.Cookie(IPAY_WECHAT_PREPAY_INNER)
	if err != nil {
		return
	}
	param, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return
	}
	reqDto = &WxReqPrepayEasyDto{}
	err = json.Unmarshal([]byte(param), reqDto)
	if err != nil {
		return
	}
	return
}

func (WxApiController) prepayRespParam(ctx context.Context, reqDto *WxReqPrepayEasyDto, account *models.WxAccount) (prePayParam map[string]interface{}, err error) {
	reqDto.ReqBaseDto = &wxsdk.ReqBaseDto{
		AppId:    account.AppId,
		SubAppId: account.SubAppId,
		MchId:    account.MchId,
		SubMchId: account.SubMchId,
	}
	customDto := wxsdk.ReqCustomerDto{
		Key: account.Key,
	}
	statusCode, _, result, err := wxsdk.Prepay(reqDto.ReqPrepayDto, &customDto)
	PrintApiBehaviorError(ctx, UrlInfo{
		ApiName:        "Wechat",
		Method:         "POST",
		Uri:            wxsdk.URLPREPAY,
		ResponseStatus: statusCode,
		Struct:         structs.New(reqDto.ReqPrepayDto),
		Extra:          "Prepay",
		Err:            err,
	})
	if err != nil {
		return
	}

	prePayParam = make(map[string]interface{}, 0)
	prePayParam["package"] = "prepay_id=" + base.ToString(result["prepay_id"])
	prePayParam["timeStamp"] = base.ToString(ChinaDatetime().Unix())
	prePayParam["nonceStr"] = result["nonce_str"]
	prePayParam["signType"] = "MD5"
	prePayParam["appId"] = result["appid"]
	prePayParam["paySign"] = sign.MakeMd5Sign(base.JoinMapObject(prePayParam), account.Key)
	prePayParam["jwtToken"], _ = auth.NewToken(map[string]interface{}{"type": "ticket"})
	return
}

/*
e_id,sub_notify_url,attach
*/
func (WxApiController) setNotifyAttach(ctx context.Context, subNotifyUrl, attach string, eId int64) (newAttach string) {
	newAttach = strconv.FormatInt(eId, 10) + factory.ConfigString(ctx, "NOTIFY_BODY_SEP2") +
		url.QueryEscape(subNotifyUrl) + factory.ConfigString(ctx, "NOTIFY_BODY_SEP2") +
		url.QueryEscape(attach)
	return
}

/*
e_id,sub_notify_url,attach
*/
func (WxApiController) getNotifyAttach(ctx context.Context, attach string) (subNotifyUrl, rawAttach string, eId int64, err error) {
	vs := strings.Split(attach, factory.ConfigString(ctx, "NOTIFY_BODY_SEP2"))
	if len(vs) != 3 {
		err = errors.New("attach param is missing[e_id,sub_notify_url,attach]")
		return
	}
	eId, err = strconv.ParseInt(vs[0], 10, 64)
	if err != nil {
		return
	}
	subNotifyUrl, err = url.QueryUnescape(vs[1])
	rawAttach, err = url.QueryUnescape(vs[2])
	return
}
