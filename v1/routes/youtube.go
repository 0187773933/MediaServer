package routes

import (
	"fmt"
	"bytes"
	"strconv"
	fiber "github.com/gofiber/fiber/v2"
	server "github.com/0187773933/GO_SERVER/v1/server"
	bolt "github.com/boltdb/bolt"
	circular "github.com/0187773933/BoltCircular/v1/circular"
	youtube "github.com/0187773933/MediaServer/v1/youtube"
)

func YouTube_Playlist_Import( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		playlist_id := c.Params( "playlist_id" )
		playlist := youtube.GetVideosInPlaylist( s.Config.MiscMap[ "google_key" ] , playlist_id )
		cl_key := fmt.Sprintf( "youtube-playlist-%s" , playlist_id )
		cl := circular.Open( s.DB , cl_key )
		for _ , item := range playlist.Items {
			cl.AddNx( []byte( item.Snippet.ResourceId.VideoId ) )
		}
		s.DB.Update( func( tx *bolt.Tx ) error {
			youtube_playlist_title , _ := tx.CreateBucketIfNotExists( []byte( "youtube-titles-playlist" ) )
			youtube_playlist_title.Put( []byte( playlist_id ) , []byte( playlist.Name ) )
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

func YouTube_Playlist_Get( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		video_ids := []string{}
		playlist_id := c.Params( "playlist_id" )
		playlist_title := ""
		s.DB.View( func( tx *bolt.Tx ) error {
			cl_key := fmt.Sprintf( "youtube-playlist-%s" , playlist_id )
			b := tx.Bucket( []byte( cl_key ) )
			b.ForEach( func( k , v []byte ) error {
				fmt.Printf( "key=%s , value=%s\n" , k , v )
				video_ids = append( video_ids , string( v ) )
				return nil
			})
			title_b := tx.Bucket( []byte( "youtube-titles-playlist" ) )
			playlist_title = string( title_b.Get( []byte( playlist_id ) ) )
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
			"playlist_id": playlist_id ,
			"title": playlist_title ,
			"ids": video_ids ,
		})
	}
}

func YouTube_Playlist_GetAll( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		prefix := []byte( "youtube-playlist-" )
		playlists := []string{}
		s.DB.View( func( tx *bolt.Tx ) error {
			tx.ForEach( func( bucket_name []byte , _ *bolt.Bucket ) error {
				if bytes.HasPrefix( bucket_name , prefix ) == false { return nil }
				id := string( bucket_name )[ len( prefix ) : ]
				playlists = append( playlists , id )
				fmt.Println( id )
				return nil
			})
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
			"playlists": playlists ,
		})
	}
}

func YouTube_Playlist_Next( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		playlist_id := c.Params( "playlist_id" )
		playlist_title := ""
		next_id := ""
		next_title := ""
		next_index := ""
		next_position := ""
		s.DB.View( func( tx *bolt.Tx ) error {
			playlist_title_b := tx.Bucket( []byte( "youtube-titles-playlist" ) )
			position_b := tx.Bucket( []byte( "youtube-id-position" ) )
			title_b := tx.Bucket( []byte( "youtube-id-title" ) )
			playlist_title = string( playlist_title_b.Get( []byte( playlist_id ) ) )
			cl_key := fmt.Sprintf( "youtube-playlist-%s" , playlist_id )
			cl := circular.Open( s.DB , cl_key )
			current , ci , _ := cl.Current()
			current_id := string( current )
			current_position := string( position_b.Get( current ) )
			current_position_int , _ := strconv.Atoi( current_position )
			if current_position_int == 0 {
				next_id = current_id
				next_index = string( ci )
				next_title = string( title_b.Get( current ) )
				next_position = "0"
				return nil
			}
			next := cl.Next()
			next_id = string( next )
			_ , ni , _ := cl.Current()
			next_index = string( ni )
			next_title = string( title_b.Get( next ) )
			next_position = "0"
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
			"playlist_id": playlist_id ,
			"playlist_title": playlist_title ,
			"next_id": next_id ,
			"next_index": next_index ,
			"next_title": next_title ,
			"next_position": next_position ,
		})
	}
}
