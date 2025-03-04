package main

import (
	// "fmt"
	"os"
	fs "io/fs"
	"embed"
	"os/signal"
	"syscall"
	"time"
	"strings"
	"flag"
	bolt "github.com/boltdb/bolt"
	fiber "github.com/gofiber/fiber/v2"
	cors "github.com/gofiber/fiber/v2/middleware/cors"
	logger "github.com/0187773933/Logger/v1/logger"
	server "github.com/0187773933/GO_SERVER/v1/server"
	server_utils "github.com/0187773933/GO_SERVER/v1/utils"
	routes "github.com/0187773933/MediaServer/v1/routes"
)

//go:embed v1/embed/*
var EMBED_FILES embed.FS

var s server.Server
var DB *bolt.DB

func SetupCloseHandler() {
	c := make( chan os.Signal )
	signal.Notify( c , os.Interrupt , syscall.SIGTERM , syscall.SIGINT )
	go func() {
		<-c
		logger.Log.Println( "\r- Ctrl+C pressed in Terminal" )
		DB.Close()
		logger.Log.Printf( "Shutting Down %s Server" , s.Config.Name )
		s.FiberApp.Shutdown()
		os.Exit( 0 )
	}()
}

func main() {
	var GenConfig bool
	flag.BoolVar( &GenConfig , "gen-config" , false , "Generate configuration file" )
	flag.BoolVar( &GenConfig , "g" , false , "Generate configuration file (shorthand)" )
	flag.Parse()
	config := server_utils.GetConfig()
	if GenConfig {
		server_utils.GenerateNewKeysWrite( &config )
		panic( "Generated Config" )
		return
	}
	// server_utils.GenerateNewKeysWrite( &config )
	defer server_utils.SetupStackTraceReport()
	logger.New( &config.Log )
	logger.SetLocation( config.TimeZone )
	DB , _ = bolt.Open( config.Bolt.Path , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	s = server.New( &config , logger.Log , DB )
	allow_origins_string := strings.Join( config.AllowOrigins , "," )
	s.FiberApp.Use( cors.New( cors.Config{
		AllowOrigins: allow_origins_string ,
		// AllowMethods: "GET, POST, PUT, DELETE, OPTIONS" ,
		AllowMethods: "GET, POST" ,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, k" ,
	}))
	s.FiberApp.Use( func( c *fiber.Ctx ) error {
		c.Set( "Cache-Control" , "no-store, no-cache, must-revalidate, proxy-revalidate" )
		c.Set( "Pragma" , "no-cache" )
		c.Set( "Expires" , "0" )
		return c.Next()
	})
	s.EMBEDED , _ = fs.Sub( EMBED_FILES , "v1/embed" )
	routes.SetupPublicRoutes( &s )
	routes.SetupAdminRoutes( &s )
	SetupCloseHandler()
	s.Start()
}