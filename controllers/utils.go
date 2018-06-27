package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/relax-space/go-kitt/auth"

	"github.com/mitchellh/mapstructure"
	"github.com/pangpanglabs/goutils/behaviorlog"

	"github.com/fatih/structs"
	"github.com/labstack/echo"
)

const (
	FlashName      = "flash"
	FlashSeparator = ";"
)

type ApiResult struct {
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
	Error   ApiError    `json:"error"`
}

type ApiError struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type ArrayResult struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"totalCount"`
}

var (
	// System Error
	ApiErrorSystem             = ApiError{Code: 10001, Message: "System Error"}
	ApiErrorServiceUnavailable = ApiError{Code: 10002, Message: "Service unavailable"}
	ApiErrorRemoteService      = ApiError{Code: 10003, Message: "Remote service error"}
	ApiErrorIPLimit            = ApiError{Code: 10004, Message: "IP limit"}
	ApiErrorPermissionDenied   = ApiError{Code: 10005, Message: "Permission denied"}
	ApiErrorIllegalRequest     = ApiError{Code: 10006, Message: "Illegal request"}
	ApiErrorHTTPMethod         = ApiError{Code: 10007, Message: "HTTP method is not suported for this request"}
	ApiErrorParameter          = ApiError{Code: 10008, Message: "Parameter error"}
	ApiErrorMissParameter      = ApiError{Code: 10009, Message: "Miss required parameter"}
	ApiErrorDB                 = ApiError{Code: 10010, Message: "DB error, please contact the administator"}
	ApiErrorTokenInvaild       = ApiError{Code: 10011, Message: "Token invaild"}
	ApiErrorMissToken          = ApiError{Code: 10012, Message: "Miss token"}
	ApiErrorVersion            = ApiError{Code: 10013, Message: "API version %s invalid"}
	ApiErrorNotFound           = ApiError{Code: 10014, Message: "Resource not found"}
	// Business Error
	ApiErrorUserNotExists = ApiError{Code: 20001, Message: "User does not exists"}
	ApiErrorPassword      = ApiError{Code: 20002, Message: "Password error"}
	ApiErrorAlipay        = ApiError{Code: 20003, Message: "Request alipay error"}
	ApiErrorWechat        = ApiError{Code: 20004, Message: "Request wechat error"}
	ApiErrorIpayCheck     = ApiError{Code: 20005, Message: "sign or data invalid"}
)

func ReturnApiFail(c echo.Context, status int, apiError ApiError, err error, v ...map[string]interface{}) error {
	logContext := behaviorlog.FromCtx(c.Request().Context())
	if logContext != nil {
		if err != nil {
			logContext.WithError(err)
		}
		if len(v) > 0 {
			logContext.WithBizAttrs(v[0])
		}
	}

	str := ""
	if err != nil {
		str = err.Error()
	}
	return c.JSON(status, ApiResult{
		Success: false,
		Error: ApiError{
			Code:    apiError.Code,
			Message: apiError.Message,
			Details: str,
		},
	})
}
func ReturnApiAliPushFail(c echo.Context, status int, err error, v ...map[string]interface{}) error {
	logContext := behaviorlog.FromCtx(c.Request().Context())
	if logContext != nil {
		if err != nil {
			logContext.WithError(err)
		}
		if len(v) > 0 {
			logContext.WithBizAttrs(v[0])
		}
	}
	return c.String(status, "failure")
}

func ReturnApiWxPushFail(c echo.Context, status int, err error, v ...map[string]interface{}) error {
	logContext := behaviorlog.FromCtx(c.Request().Context())
	if logContext != nil {
		if err != nil {
			logContext.WithError(err)
		}
		if len(v) > 0 {
			logContext.WithBizAttrs(v[0])
		}
	}
	errResult := struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
	}{xml.Name{}, "FAIL", ""}
	if err != nil {
		errResult.ReturnMsg = err.Error()
	}
	return c.XML(status, errResult)
}

func NotifyError(c echo.Context, errMsg string) error {
	errResult := struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
	}{xml.Name{}, "FAIL", ""}
	errResult.ReturnMsg = errMsg
	return c.XML(http.StatusBadRequest, errResult)

}

func ReturnApiSucc(c echo.Context, status int, result interface{}) error {
	if status == 204 {
		return c.NoContent(status)
	}

	return c.JSON(status, ApiResult{
		Success: true,
		Result:  result,
	})
}
func ReturnApiListSucc(c echo.Context, status int, totalCount int64, items interface{}) error {
	if status == 204 {
		return c.NoContent(status)
	}
	return c.JSON(status, ApiResult{
		Success: true,
		Result:  ArrayResult{TotalCount: totalCount, Items: items},
	})
}

func setFlashMessage(c echo.Context, m map[string]string) {
	var flashValue string
	for key, value := range m {
		flashValue += "\x00" + key + "\x23" + FlashSeparator + "\x23" + value + "\x00"
	}

	c.SetCookie(&http.Cookie{
		Name:  FlashName,
		Value: url.QueryEscape(flashValue),
	})
}
func getFlashMessage(c echo.Context) map[string]string {
	cookie, err := c.Cookie(FlashName)
	if err != nil {
		return nil
	}

	m := map[string]string{}

	v, _ := url.QueryUnescape(cookie.Value)
	vals := strings.Split(v, "\x00")
	for _, v := range vals {
		if len(v) > 0 {
			kv := strings.Split(v, "\x23"+FlashSeparator+"\x23")
			if len(kv) == 2 {
				m[kv[0]] = kv[1]
			}
		}
	}
	//read one time then delete it
	c.SetCookie(&http.Cookie{
		Name:   FlashName,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	return m
}

type UrlInfo struct {
	ControllerName string
	ApiName        string //spring,sf,best
	Method         string //GET,POST
	Uri            string
	ResponseStatus int
	Struct         *structs.Struct
	Extra          interface{}
	Err            error
}

func PrintApiBehaviorError(c context.Context, urlInfo UrlInfo) {
	logContext := behaviorlog.FromCtx(c)
	if logContext != nil {
		logClone := logContext.Clone()
		if urlInfo.Err != nil {
			logClone.WithError(urlInfo.Err)
		}
		logClone.Controller = urlInfo.ControllerName
		logClone.Params = map[string]interface{}{}
		urlInfo.Struct.TagName = "json"
		logClone.WithCallURLInfo(
			urlInfo.Method,
			urlInfo.Uri,
			urlInfo.Struct.Map(),
			urlInfo.ResponseStatus,
		).WithBizAttr("Extra", urlInfo.Extra).Log(urlInfo.ApiName)
		logContext.Params = map[string]interface{}{}
	}
}
func routeParse(c echo.Context) (method, version string, err error) {
	b, err := ioutil.ReadAll(c.Request().Body)
	c.Request().Body.Close()
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(b))
	var reqDto struct {
		Method string `json:"method"`
	}
	err = json.Unmarshal(b, &reqDto)
	if err != nil {
		return
	}
	method = reqDto.Method
	return
}

func ChinaDatetime() (date time.Time) {
	date = time.Now().UTC().Add(8 * time.Hour)
	return
}

func SetCookie(key, value string, c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = key
	value = url.QueryEscape(value)
	cookie.Value = value
	index := strings.Index(c.Request().Host, ".")
	cookie.Domain = c.Request().Host[index+1:]
	cookie.Path = "/"
	c.SetCookie(cookie)
}

func SetCookieObj(key string, value interface{}, c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = key
	b, _ := json.Marshal(value)
	cookie.Value = url.QueryEscape(string(b))
	index := strings.Index(c.Request().Host, ".")
	cookie.Domain = c.Request().Host[index+1:]
	cookie.Path = "/"
	c.SetCookie(cookie)
}

func getToken() (token string, err error) {
	m := make(map[string]interface{}, 0)
	m["key"] = "value"
	token, err = auth.NewToken(m)
	return
}

func POSTXml(token, url, param string, v interface{}) (resp *http.Response, err error) {
	b := []byte(param)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		err = fmt.Errorf("HTTP New Request Error: %s", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/xml")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err = (&http.Client{}).Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("[%d %s]%s", resp.StatusCode, resp.Status, string(b))
		return
	}
	if v != nil {
		dec := xml.NewDecoder(resp.Body)
		if err = dec.Decode(&v); err != nil {
			return
		}
	}

	return
}

const (
	NOTIFY_BODY_SEP1 = "&"
	NOTIFY_BODY_SEP2 = "||||"
)

// Decode takes a map and uses reflection to convert it into the
// given Go native structure. val must be a pointer to a struct.
func Decode(m interface{}, rawVal interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   rawVal,
		TagName:  "mapstruct",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}
