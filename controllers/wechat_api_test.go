package controllers

import "testing"

func Test_Wechat_Pay(t *testing.T) {
	bodyStr := `
	{
		"e_id":10001,
		"auth_code":"135298324463700425",
		"body":"xiaoxinmiao test",
		"total_fee":1
	}`
	ReqCommon(t, bodyStr)
}

func Test_Wechat_Query(t *testing.T) {
	bodyStr := `
	{
		"e_id":10001,
		"out_trade_no":"14201711085205823413229775520"
	}`
	ReqCommon(t, bodyStr)
}

func Test_Wechat_Refund(t *testing.T) {
	bodyStr := `
	{
		"e_id":10001,
		"out_trade_no":"147688874645492354650",
		"refund_fee":1
	}`
	ReqCommon(t, bodyStr)
}

func Test_Wechat_Reverse(t *testing.T) {
	bodyStr := `
	{
		"e_id":10001,
		"out_trade_no":"143420620288156126697"
	}`
	ReqCommon(t, bodyStr)
}

func Test_Wechat_RefundQuery(t *testing.T) {
	bodyStr := `
	{
		"e_id":10001,
		"out_trade_no":"144650782494807835413"
	}`
	ReqCommon(t, bodyStr)
}

func Test_Wechat_PrePay(t *testing.T) {
	bodyStr := `
	{
		"e_id":10001,
		"body":"xiaomiao test",
		"total_fee":1,
		"trade_type":"JSAPI",
		"notify_url":"http://xiao.xinmiao.com",
		"openid":"os2u9uPKLkCKL08FwCM6hQAQ_LtI"
	}`
	ReqCommon(t, bodyStr)
}

func Test_Wechat_Notify(t *testing.T) {
	bodyStr := `<xml>
	<appid><![CDATA[wx2421b1c4370ec43b]]></appid>
	<attach><![CDATA[{"sub_notify_url":"https://baidu.com","e_id":10001}]]></attach>
	<bank_type><![CDATA[CFT]]></bank_type>
	<fee_type><![CDATA[CNY]]></fee_type>
	<is_subscribe><![CDATA[Y]]></is_subscribe>
	<mch_id><![CDATA[10000100]]></mch_id>
	<nonce_str><![CDATA[5d2b6c2a8db53831f7eda20af46e531c]]></nonce_str>
	<openid><![CDATA[oUpF8uMEb4qRXf22hE3X68TekukE]]></openid>
	<out_trade_no><![CDATA[1409811653]]></out_trade_no>
	<result_code><![CDATA[SUCCESS]]></result_code>
	<return_code><![CDATA[SUCCESS]]></return_code>
	<sign><![CDATA[7D24E7B803ED7574785872A50105046D]]></sign>
	<sub_mch_id><![CDATA[10000100]]></sub_mch_id>
	<time_end><![CDATA[20140903131540]]></time_end>
	<total_fee>1</total_fee>
	<trade_type><![CDATA[JSAPI]]></trade_type>
	<transaction_id><![CDATA[B2AE05C99B9C81A640472406AA3C2710]]></transaction_id>
 </xml>`
	ReqCommon(t, bodyStr)
}

func Test_Wechat_Notify_ServiceProvider(t *testing.T) {
	bodyStr := `<xml><appid><![CDATA[wx856df5e42a345096]]></appid>
	<attach><![CDATA[e_id||||10001&sub_notify_url||||https://baidu.com]]></attach>
	<bank_type><![CDATA[CMB_CREDIT]]></bank_type>
	<cash_fee><![CDATA[1]]></cash_fee>
	<fee_type><![CDATA[CNY]]></fee_type>
	<is_subscribe><![CDATA[Y]]></is_subscribe>
	<mch_id><![CDATA[1294997801]]></mch_id>
	<nonce_str><![CDATA[1240648768328515708]]></nonce_str>
	<openid><![CDATA[os2u9uBHeJRPtCkisjVf-kWZWjKQ]]></openid>
	<out_trade_no><![CDATA[169126120915612414200792892307]]></out_trade_no>
	<result_code><![CDATA[SUCCESS]]></result_code>
	<return_code><![CDATA[SUCCESS]]></return_code>
	<sign><![CDATA[DACC65EAD9461590F693C08EAB2F0A10]]></sign>
	<sub_appid><![CDATA[wx38db2bfbb79a3cea]]></sub_appid>
	<sub_is_subscribe><![CDATA[Y]]></sub_is_subscribe>
	<sub_mch_id><![CDATA[1464381802]]></sub_mch_id>
	<sub_openid><![CDATA[o2-sBj3ozQQ6gxiyYKI2JzJFcUhY]]></sub_openid>
	<time_end><![CDATA[20171209235456]]></time_end>
	<total_fee>1</total_fee>
	<trade_type><![CDATA[JSAPI]]></trade_type>
	<transaction_id><![CDATA[4200000031201712091054287297]]></transaction_id>
	</xml>`
	ReqCommon(t, bodyStr)
}
