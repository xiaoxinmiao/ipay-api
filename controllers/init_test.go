package controllers

import (
	"context"
	"fmt"
	"ipay-api/factory"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/asaskevich/govalidator"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/relax-space/go-kit/test"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
)

var (
	echoApp          *echo.Echo
	handleWithFilter func(handlerFunc echo.HandlerFunc, c echo.Context) error
	ctx              context.Context
)

func init() {
	db, err := initDB("mysql", os.Getenv("IPAY_CONN"))
	if err != nil {
		panic(err)
	}
	fmt.Println(os.Getenv("IPAY_CONN"))

	echoApp = echo.New()
	echoApp.Validator = &Validator{}
	configMap := map[string]interface{}{
		"NOTIFY_BODY_SEP1": "&",
		"NOTIFY_BODY_SEP2": "||||",
		"IPAY_HOST":        os.Getenv("IPAY_HOST"),
		"JWT_SECRET":       os.Getenv("JWT_SECRET"),
	}
	setContextValueMiddleware := setContextValue(&configMap, db)
	handleWithFilter = func(handlerFunc echo.HandlerFunc, c echo.Context) error {
		return setContextValueMiddleware(handlerFunc)(c)
	}
}

func setContextValue(configMap *map[string]interface{}, db *xorm.Engine) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			reqContext := context.WithValue(req.Context(), echomiddleware.ContextDBName, db)
			reqContext = context.WithValue(reqContext, factory.ContextConfigName, configMap)
			c.SetRequest(req.WithContext(reqContext))
			return next(c)
		}
	}
}

func initDB(driver, connection string) (*xorm.Engine, error) {
	db, err := xorm.NewEngine(driver, connection)
	if err != nil {
		return nil, err
	}
	db.ShowSQL(true)
	return db, nil
}

type Validator struct{}

func (v *Validator) Validate(i interface{}) error {
	_, err := govalidator.ValidateStruct(i)
	return err
}

func ReqCommon(t *testing.T, bodyStr string) {
	req, err := http.NewRequest(echo.POST, "/", strings.NewReader(bodyStr))
	test.Ok(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	test.Ok(t, handleWithFilter(RouterController{}.Post, echoApp.NewContext(req, rec)))
	fmt.Println(string(rec.Body.Bytes()))
}
