package routes

import (
	"fmt"
	fiber "github.com/gofiber/fiber/v2"
	server "github.com/0187773933/GO_SERVER/v1/server"
	bolt "github.com/boltdb/bolt"
	circular "github.com/0187773933/BoltCircular/v1/circular"
	youtube "github.com/0187773933/MediaServer/v1/youtube"
)


func YouTube_ImportPlaylist( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		playlist_id := c.Params( "playlist_id" )
		playlist := youtube.GetVideosInPlaylist( s.Config.MiscMap[ "google_key" ] , playlist_id )
		cl_key := fmt.Sprintf( "youtube-playlist-%s" , playlist_id )
		cl := circular.Open( s.DB , cl_key )
		for _ , item := range playlist.Items {
			cl.AddNx( []byte( item.Snippet.ResourceId.VideoId ) )
		}
		s.DB.Update( func( tx *bolt.Tx ) error {
			youtube_id_title , _ := tx.CreateBucketIfNotExists( []byte( "youtube-id-title" ) )
			youtube_id_position , _ := tx.CreateBucketIfNotExists( []byte( "youtube-id-position" ) )
			for _ , item := range playlist.Items {
				id := item.Snippet.ResourceId.VideoId
				title := item.Snippet.Title
				youtube_id_title.Put( []byte( id ) , []byte( title ) )
				youtube_id_position.Put( []byte( id ) , []byte( "0" ) )
			}
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
		})
	}
}