package controllers

import (
	alpay "github.com/relax-space/lemon-alipay-sdk"
	wxpay "github.com/relax-space/lemon-wxpay-sdk"
)

const (
	DefaultMaxResultCount = 30
)

type SearchInput struct {
	Sortby         []string `query:"sortby"`
	Order          []string `query:"order"`
	SkipCount      int      `query:"skipCount"`
	MaxResultCount int      `query:"maxResultCount"`
}

type AliReqPayDto struct {
	*alpay.ReqPayDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}
type AliReqQueryDto struct {
	*alpay.ReqQueryDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}
type AliReqRefundDto struct {
	*alpay.ReqRefundDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}
type AliReqReverseDto struct {
	*alpay.ReqReverseDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}

type AliReqPrepayDto struct {
	*alpay.ReqPrepayDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}

type WxReqPayDto struct {
	*wxpay.ReqPayDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}
type WxReqQueryDto struct {
	*wxpay.ReqQueryDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}
type WxReqRefundDto struct {
	*wxpay.ReqRefundDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}
type WxReqReverseDto struct {
	*wxpay.ReqReverseDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}
type WxReqRefundQueryDto struct {
	*wxpay.ReqRefundQueryDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}
type WxReqPrepayDto struct {
	*wxpay.ReqPrepayDto
	EId     int64  `json:"e_id"`
	Method  string `json:"method"`
	Version string `json:"version"`
}

type WxReqPrepayEasyDto struct {
	*wxpay.ReqPrepayDto
	EId   int64  `json:"e_id" query:"e_id"`
	AppId string `json:"app_id" query:"app_id"` //required
	Scope string `json:"scope" query:"scope"`   //option
	State string `json:"state" query:"state"`   //option

	RedirectUrl string `json:"redirect_url" query:"redirect_url"`
	PageUrl     string `json:"page_url" query:"page_url"` //option
	Method      string `json:"method"`
	Version     string `json:"version"`
}
