package main

import (
	"github.com/heirko/iris-contrib/middleware/logrus-logger/example/backend/api"
	"github.com/heirko/iris-contrib/middleware/logrus-logger/example/backend/routes"
	"github.com/Sirupsen/logrus"
	"github.com/iris-contrib/middleware/logger"
	"github.com/iris-contrib/template/html"
	"github.com/kataras/iris"
	"github.com/spf13/viper"
	"fmt"
	_ "github.com/gogap/logrus_mate/hooks/file"
	"github.com/heirko/go-contrib/logrusHelper"
)

func main() {
	// set the template engine
	iris.UseTemplate(html.New(html.Config{Layout: "layout.html"})).Directory("../frontend/templates", ".html")
	// set the favicon
	iris.Favicon("../frontend/public/images/favicon.ico")

	// set static folder(s)
	iris.Static("/public", "../frontend/public", 1)

	// set the global middlewares
	iris.Use(logger.New(iris.Logger))

	// set the custom errors
	iris.OnError(iris.StatusNotFound, func(ctx *iris.Context) {
		ctx.Render("errors/404.html", iris.Map{"Title": iris.StatusText(iris.StatusNotFound)})
	})

	iris.OnError(iris.StatusInternalServerError, func(ctx *iris.Context) {
		ctx.Render("errors/500.html", nil, iris.RenderOptions{"layout": iris.NoLayout})
	})

	// register the routes & the public API
	registerRoutes()
	registerAPI()
	initLogger()
	// start the server
	iris.Listen("127.0.0.1:8080")
}

func initLogger() {

	// ########## Init Viper
	var viper = viper.New()

	viper.SetConfigName("mate") // name of config file (without extension), here we use some logrus_mate sample
	viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	// ########### End Init Viper

	// Read configuration
	var c = logrusHelper.UnmarshalConfiguration(viper) // Unmarshal configuration from Viper
	logrusHelper.SetConfig(logrus.StandardLogger(), c) // for e.g. apply it to logrus default instance

	// ### End Read Configuration

	// ### Use logrus as normal
	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
	}).Error("A walrus appears")
}

func registerRoutes() {
	// register index using a 'Handler'
	iris.Handle("GET", "/", routes.Index())

	// this is other way to declare a route
	// using a 'HandlerFunc'
	iris.Get("/about", routes.About)

	// Dynamic route

	iris.Get("/profile/:username", routes.Profile)("user-profile")
	// user-profile is the custom,optional, route's Name: with this we can use the {{ url "user-profile" $username}} inside userlist.html

	iris.Get("/all", routes.UserList)
}

func registerAPI() {
	// this is other way to declare routes using the 'API'
	iris.API("/users", api.UserAPI{})
}

