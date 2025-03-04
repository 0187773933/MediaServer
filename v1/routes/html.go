package routes

import (
	"io"
	fiber "github.com/gofiber/fiber/v2"
	server "github.com/0187773933/GO_SERVER/v1/server"
)

func HTML_ImportPage( s *server.Server ) fiber.Handler {
	if s.Config.MiscMap[ "mode" ] == "dev" {
		return func( c *fiber.Ctx ) error {
			c.Set( "Content-Type" , "text/html" )
			return c.SendFile( "./v1/embed/html/import.html" )
		}
	} else {
		return func( c *fiber.Ctx ) error {
			c.Set( "Content-Type" , "text/html" )
			file_path := "html/import.html"
			file , _ := s.EMBEDED.Open( file_path )
			defer file.Close()
			content , _ := io.ReadAll( file )
			return c.SendString( string( content ) )
		}
	}
}