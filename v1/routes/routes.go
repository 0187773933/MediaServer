package routes

import (
	"fmt"
	"time"
	// hex "encoding/hex"
	// base64 "encoding/base64"
	fiber "github.com/gofiber/fiber/v2"
	server "github.com/0187773933/GO_SERVER/v1/server"
	// encryption "github.com/0187773933/encryption/v1/enscryption"
	rate_limiter "github.com/gofiber/fiber/v2/middleware/limiter"
)

func PublicMaxedOut( c *fiber.Ctx ) error {
	ip_address := c.IP()
	log_message := fmt.Sprintf( "%s === %s === %s === PUBLIC RATE LIMIT REACHED !!!" , ip_address , c.Method() , c.Path() );
	fmt.Println( log_message )
	c.Set( "Content-Type" , "text/html" )
	return c.SendString( "<html><h1>loading ...</h1><script>setTimeout(function(){ window.location.reload(1); }, 6000);</script></html>" )
}
var PublicLimter = rate_limiter.New( rate_limiter.Config{
	Max: 3 ,
	Expiration: 1 * time.Second ,
	KeyGenerator: func( c *fiber.Ctx ) string {
		return c.Get( "x-forwarded-for" )
	} ,
	LimitReached: PublicMaxedOut ,
	LimiterMiddleware: rate_limiter.SlidingWindow{} ,
})

func SetupPublicRoutes( s *server.Server ) {
	prefix_string := "/"
	if s.Config.URLS.Prefix != "" {
		prefix_string = s.Config.URLS.Prefix
	}
	prefix := s.FiberApp.Group( prefix_string )

	prefix.Get( "/test/:filename" , func( c *fiber.Ctx ) error {
		fmt.Println( c.Params( "filename" ) )
		file := "./v1/cdn/" + c.Params("filename")
		c.Set( "Content-Type" , "application/javascript" )
		fmt.Println( file )
		return c.SendFile( file )
	})

	prefix.Get( "/" , PublicLimter , func( c *fiber.Ctx ) error {
		return c.JSON( fiber.Map{
			"result": true ,
			"url": "/" ,
		})
	})
	// s.FiberApp.Get( "/twitch" , func( c *fiber.Ctx ) error {
	// 	c.Set( "Content-Type" , "text/html" )
	// 	return c.SendFile( "./v1/html/twitch.html" )
	// })
	// s.FiberApp.Post( "/update_position" , PublicLimter , UpdatePosition( s ) )
	// prefix := s.FiberApp.Group( s.Config.URLS.Prefix )
	// prefix.Get( "/:uuid.:ext" , UUIDFileLimter , ServeFile( s ) )

	// // youtube
	// youtube := prefix.Group( "/youtube" )
	// youtube.Get( "/:library_key/:session_id" , YouTubeSessionHTMLPlayer( s ) )

	// // local library
	// library := prefix.Group( "/library" )
	// library.Get( "/get/entries" , LibraryGetEntries( s ) )
	// // library-session
	// prefix.Use( PublicLimter )
	// prefix.Get( "/:library_key/:session_id/reset" , SessionReset( s ) )
	// prefix.Get( "/:library_key/:session_id/total" , SessionTotal( s ) )
	// prefix.Get( "/:library_key/:session_id/index" , SessionIndex( s ) )
	// prefix.Get( "/:library_key/:session_id/set/index/:index" , SessionSetIndex( s ) )
	// prefix.Get( "/:library_key/:session_id/previous" , SessionPrevious( s ) )
	// prefix.Get( "/:library_key/:session_id/next" , SessionNext( s ) )
	// prefix.Get( "/:library_key/:session_id" , SessionHTMLPlayer( s ) ) // HTML Player
	// prefix.Get( "/:library_key/:session_id/:index" , SessionHTMLPlayerAtIndex( s ) ) // HTML Player at Session Index ?
}

func SetupAdminRoutes( s *server.Server ) {
	prefix_string := "/"
	if s.Config.URLS.AdminPrefix != "" {
		prefix_string = s.Config.URLS.AdminPrefix
	}
	admin := s.FiberApp.Group( prefix_string )
	admin.Get( "/" , PublicLimter , func( c *fiber.Ctx ) error {
		return c.JSON( fiber.Map{
			"result": true ,
			"url": "/" ,
		})
	})
	admin.Use( s.ValidateAdminMW )
	admin.Get( "/import" , HTML_Serve( s , "import" ) )
	// youtube
	youtube := admin.Group( "/youtube" )
	// youtube.Get( "/playlist/:playlist_id" , HTML_Serve( s , "youtube-playlist" ) )
	youtube.Get( "/import/playlist/:playlist_id" , YouTube_Playlist_Import( s ) )
	youtube.Get( "/get/playlist/:playlist_id" , YouTube_Playlist_Get( s ) )
	youtube.Get( "/get/playlists" , YouTube_Playlist_GetAll( s ) )
	youtube.Get( "/playlist/:playlist_id" , YouTube_Playlist_Next_New_Session( s ) )
	// youtube.Get( "/playlist/:playlist_id/:session_id/next" , YouTube_Playlist_Next( s ) ).Name( "youtube_playlist_session_next" )
	youtube.Get( "/:session_id" , YouTube_Session_Next( s ) ).Name( "youtube_session" )
	youtube.Get( "/:session_id/next" , YouTube_Session_Next( s ) )
	// youtube.Get( "/playlist/:playlist_id/:session_id/next" , YouTube_Playlist_Next( s ) )
	// youtube.Get( "/playlist/:playlist_id/previous" , YouTube_Playlist_Previous( s ) )
	// youtube.Get( "/playlist/:playlist_id/update/position/:video_id/:position" , YouTube_Update_Position( s ) )
}