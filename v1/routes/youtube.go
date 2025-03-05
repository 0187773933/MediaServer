package routes

import (
	"fmt"
	// "bytes"
	"strconv"
	"encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	server "github.com/0187773933/GO_SERVER/v1/server"
	bolt "github.com/boltdb/bolt"
	// circular "github.com/0187773933/BoltCircular/v1/circular"
	youtube "github.com/0187773933/MediaServer/v1/youtube"
)

type YoutubeVideo struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Position int `json:"position"`
}

type YoutubePlaylist struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Index int `json:"index"`
	Total int `json:"total"`
	Videos []YoutubeVideo `json:"videos"`
}
func YouTube_Playlist_Import( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		playlist_id := c.Params( "playlist_id" )
		playlist := youtube.GetVideosInPlaylist( s.Config.MiscMap[ "google_key" ] , playlist_id )
		var ytp YoutubePlaylist
		ytp.Id = playlist_id
		ytp.Name = playlist.Name
		ytp.Index = 0
		ytp.Total = len( playlist.Items )
		for _ , item := range playlist.Items {
			var v YoutubeVideo
			v.Id = item.Snippet.ResourceId.VideoId
			v.Name = item.Snippet.Title
			v.Position = 0
			ytp.Videos = append( ytp.Videos , v )
		}
		ytp_json , _ := json.Marshal( ytp )
		s.DB.Update( func( tx *bolt.Tx ) error {
			b , _ := tx.CreateBucketIfNotExists( []byte( "youtube-playlists" ) )
			b.Put( []byte( playlist_id ) , ytp_json )
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
			"playlist": ytp ,
		})
	}
}

func YouTube_Playlist_Get( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		playlist_id := c.Params( "playlist_id" )
		var ytp YoutubePlaylist
		s.DB.View( func( tx *bolt.Tx ) error {
			b := tx.Bucket( []byte( "youtube-playlists" ) )
			ytp_json := b.Get( []byte( playlist_id ) )
			json.Unmarshal( ytp_json , &ytp )
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
			"playlist": ytp ,
		})
	}
}

func YouTube_Playlist_GetAll( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		var playlists []YoutubePlaylist
		s.DB.View( func( tx *bolt.Tx ) error {
			b := tx.Bucket( []byte( "youtube-playlists" ) )
			b.ForEach( func( k , v []byte ) error {
				var ytp YoutubePlaylist
				json.Unmarshal( v , &ytp )
				playlists = append( playlists , ytp )
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
		var ytp YoutubePlaylist
		var next_id string
		var next_index int
		var next_position int
		s.DB.Update( func( tx *bolt.Tx ) error {
			b := tx.Bucket( []byte( "youtube-playlists" ) )
			ytp_json := b.Get( []byte( playlist_id ) )
			json.Unmarshal( ytp_json , &ytp )
			if ytp.Videos[ ytp.Index ].Position == -1 {
				ytp.Index++
				if ytp.Index >= ytp.Total {
					ytp.Index = 0
				}
				if ytp.Videos[ ytp.Index ] == -1 {
					next_index = ytp.Index
					next_position = 0
					break
				}
				next_id = ytp.Videos[ ytp.Index ].Id
				next_index = ytp.Index
				next_position = ytp.Videos[ ytp.Index ].Position
			} else {
				// continuing where you left off , not finished
				next_id = ytp.Videos[ ytp.Index ].Id
				next_index = ytp.Index
				next_position = ytp.Videos[ ytp.Index ].Position
				fmt.Println( "continuing where you left off" , next_id , next_index , next_position )
			}
			// restore
			ytp_json , _ = json.Marshal( ytp )
			b.Put( []byte( playlist_id ) , ytp_json )
			return nil
		})
		return c.JSON(fiber.Map{
			"result": true ,
			"playlist_id": playlist_id ,
			"next_id": next_id ,
			"next_index": next_index ,
			"next_position": next_position ,
		})
	}
}

func YouTube_Update_Position( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		playlist_id := c.Params( "playlist_id" )
		video_id := c.Params( "video_id" )
		position := c.Params( "position" )
		fmt.Println( "YouTube_Update_Position" , video_id , position )
		s.DB.Update( func( tx *bolt.Tx ) error {
			b , _ := tx.CreateBucketIfNotExists( []byte( "youtube-playlists" ) )
			ytp_json := b.Get( []byte( playlist_id ) )
			var ytp YoutubePlaylist
			json.Unmarshal( ytp_json , &ytp )
			for i , _ := range ytp.Videos {
				if ytp.Videos[ i ].Id != video_id { continue; }
				position_int , _ := strconv.Atoi( position )
				ytp.Videos[ i ].Position = position_int
				fmt.Println( "setting position" , video_id , i , position_int )
				if ytp.Index != i {
					fmt.Println( "resetting index" , i )
					ytp.Index = i
				}
				break
			}
			ytp_json , _ = json.Marshal( ytp )
			b.Put( []byte( playlist_id ) , ytp_json )
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
		})
	}
}