package main

import (
	"crypto/subtle"
	"flag"
	"sync"
	"time"

	"github.com/izuc/asset-registry/pkg/registryservice"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	// defaultMongoDBOpTimeout defines the default MongoDB operation timeout.
	defaultMongoDBOpTimeout = 5 * time.Second
)

var (
	clientDB   *mongo.Client
	dbOnce     sync.Once
	server     *echo.Echo
	serverOnce sync.Once
	log        *zap.SugaredLogger
)

func main() {
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()

	service := registryservice.NewService(mongoDB())
	httpHandler := registryservice.NewHTTPHandler(service, log)

	Server()

	// configure the server
	server.HideBanner = true
	server.HidePort = true

	server.Group("/admin")
	adminGroup := middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Be careful to use constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(*basicAuthUser)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(*basicAuthPassword)) == 1 {
			return true, nil
		}
		return false, nil
	})

	server.GET("/", IndexRequest)
	server.POST("/registries/:network/assets", httpHandler.SaveAsset)
	server.GET("/registries/:network/assets", httpHandler.LoadAssets)
	server.GET("/registries/:network/assets/:ID", httpHandler.LoadAsset)

	server.DELETE("/admin/:network/assets/byID/:ID", httpHandler.DeleteAssetByID, adminGroup)
	server.DELETE("/admin/:network/assets/byName/:name", httpHandler.DeleteAssetByName, adminGroup)
	server.POST("/admin/filters/:word", httpHandler.AddFilter, adminGroup)
	server.DELETE("/admin/filters/:word", httpHandler.DeleteFilter, adminGroup)
	server.GET("/admin/filters", httpHandler.LoadFilter, adminGroup)

	log.Infof("Starting server ...")

	log.Fatal(server.Start(*httpBindAddr))
}
