package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/asaskevich/govalidator"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	configutil "github.com/pangpanglabs/goutils/config"
	"github.com/pangpanglabs/goutils/echomiddleware"

	"ipay-api/controllers"
	"ipay-api/factory"
	"ipay-api/models"
)

var (
	handleWithFilter func(handlerFunc echo.HandlerFunc, c echo.Context) error
)

func main() {
	appEnv := flag.String("app-env", os.Getenv("APP_ENV"), "app env")
	ipayConnEnv := flag.String("IPAY_CONN", os.Getenv("IPAY_CONN"), "IPAY_CONN")
	ipayConnDemoEnv := flag.String("DEMO_IPAY_CONN", os.Getenv("DEMO_IPAY_CONN"), "DEMO_IPAY_CONN")
	hostUrl := flag.String("IPAY_HOST", os.Getenv("IPAY_HOST"), "IPAY_HOST")
	jwtEnv := flag.String("JWT_SECRET", os.Getenv("JWT_SECRET"), "JWT_SECRET")

	flag.Parse()
	var c Config
	if err := configutil.Read(*appEnv, &c); err != nil {
		panic(err)
	}

	fmt.Println(c)
	connEnv := *ipayConnEnv
	if *appEnv == "demo" {
		connEnv = *ipayConnDemoEnv
	}
	db, err := initDB("mysql", connEnv)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.GET("/swagger", func(c echo.Context) error {
		return c.File("./swagger.yml")
	})
	e.Static("/docs", "./swagger-ui")
	apiVersion := "/v3"
	e.GET(apiVersion, controllers.RouterController{}.Get)
	e.POST(apiVersion, controllers.RouterController{}.Post)
	e.GET(apiVersion+"/jwt", controllers.RouterController{}.Get)
	e.POST(apiVersion+"/jwt", controllers.RouterController{}.Post)

	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(*jwtEnv),
		Skipper: func(c echo.Context) bool {
			ignore := []string{
				apiVersion + "/jwt",
			}

			for _, i := range ignore {
				if strings.HasPrefix(c.Request().URL.Path, i) {
					return false
				}
			}
			return true
		},
	}))

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Use(middleware.RequestID())
	e.Use(echomiddleware.ContextLogger())
	e.Use(echomiddleware.ContextDB(c.Service, db, echomiddleware.KafkaConfig(c.Logger.Kafka)))
	e.Use(echomiddleware.BehaviorLogger(c.Service, echomiddleware.KafkaConfig(c.BehaviorLog.Kafka)))

	e.Validator = &Validator{}

	e.Debug = c.Debug

	configMap := map[string]interface{}{
		"NOTIFY_BODY_SEP1": "&",
		"NOTIFY_BODY_SEP2": "||||",
		"IPAY_HOST":        *hostUrl,
		"JWT_SECRET":       *jwtEnv,
	}
	setContextValueMiddleware := setContextValue(&configMap)
	handleWithFilter = func(handlerFunc echo.HandlerFunc, c echo.Context) error {
		return setContextValueMiddleware(handlerFunc)(c)
	}

	if err := e.Start(":" + c.HttpPort); err != nil {
		log.Println(err)
	}
}

func setContextValue(configMap *map[string]interface{}) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			reqContext := context.WithValue(req.Context(), factory.ContextConfigName, configMap)
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
	db.Sync(new(models.AlAccount),
		new(models.AlNotify),
		new(models.WxAccount),
		new(models.WxNotify),
	)
	return db, nil
}

type Config struct {
	Logger struct {
		Kafka echomiddleware.KafkaConfig
	}
	BehaviorLog struct {
		Kafka echomiddleware.KafkaConfig
	}
	Trace struct {
		Zipkin echomiddleware.ZipkinConfig
	}

	Debug    bool
	Service  string
	HttpPort string
}

type Validator struct{}

func (v *Validator) Validate(i interface{}) error {
	_, err := govalidator.ValidateStruct(i)
	return err
}
