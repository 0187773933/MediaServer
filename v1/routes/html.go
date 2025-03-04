package routes

import (
	"io"
	"fmt"
	fiber "github.com/gofiber/fiber/v2"
	server "github.com/0187773933/GO_SERVER/v1/server"
)

func HTML_Serve( s *server.Server , file_name string ) fiber.Handler {
	if s.Config.MiscMap[ "mode" ] == "dev" {
		return func( c *fiber.Ctx ) error {
			c.Set( "Content-Type" , "text/html" )
			return c.SendFile( fmt.Sprintf( "./v1/embed/html/%s.html" , file_name ) )
		}
	} else {
		return func( c *fiber.Ctx ) error {
			c.Set( "Content-Type" , "text/html" )
			file_path := fmt.Sprintf( "html/%s.html" , file_name )
			file , _ := s.EMBEDED.Open( file_path )
			defer file.Close()
			content , _ := io.ReadAll( file )
			return c.SendString( string( content ) )
		}
	}
}