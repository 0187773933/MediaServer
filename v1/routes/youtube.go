package routes

import (
	"fmt"
	// "bytes"
	"strconv"
	// "strings"
	"time"
	"encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	server "github.com/0187773933/GO_SERVER/v1/server"
	bolt "github.com/boltdb/bolt"
	// circular "github.com/0187773933/BoltCircular/v1/circular"
	youtube "github.com/0187773933/MediaServer/v1/youtube"
	utils "github.com/0187773933/MediaServer/v1/utils"
)

type YoutubeVideo struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Position int `json:"position"`
	Accessed int64 `json:"accessed"`
	Description string `json:"description"`
	Thumbnail string `json:"thumbnail"`
	Tags []string `json:"tags"`
	ChannelId string `json:"channel_id"`
	ChannelTitle string `json:"channel_title"`
}

type YoutubePlaylist struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Tags []string `json:"tags"`
	Accessed int64 `json:"accessed"`
	Index int `json:"index"`
	Total int `json:"total"`
	Videos []YoutubeVideo `json:"videos"`
}

type YoutubeSession struct {
	Id string `json:"id"`
	Tags []string `json:"tags"`
	Accessed int64 `json:"accessed"`
	Playlist YoutubePlaylist `json:"playlist"`
}

func YouTube_Playlist_Import( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		playlist_id := c.Params( "playlist_id" )
		playlist := youtube.GetVideosInPlaylist( s.Config.MiscMap[ "google_key" ] , playlist_id )
		now := time.Now().UnixMilli()
		var ytp YoutubePlaylist
		ytp.Id = playlist_id
		ytp.Name = playlist.Name
		ytp.Index = 0
		ytp.Accessed = now
		ytp.Total = len( playlist.Items )
		for _ , item := range playlist.Items {
			// https://pkg.go.dev/google.golang.org/api@v0.228.0/youtube/v3#PlaylistItemSnippet
			// https://pkg.go.dev/google.golang.org/api@v0.228.0/youtube/v3#ThumbnailDetails
			var v YoutubeVideo
			v.Id = item.Snippet.ResourceId.VideoId
			v.Name = item.Snippet.Title
			v.Description = item.Snippet.Description
			if item.Snippet.Thumbnails.Maxres != nil {
				v.Thumbnail = item.Snippet.Thumbnails.Maxres.Url
			} else if item.Snippet.Thumbnails.High != nil {
				v.Thumbnail = item.Snippet.Thumbnails.High.Url
			} else {
				v.Thumbnail = item.Snippet.Thumbnails.Default.Url
			}
			v.Position = 0
			v.Accessed = now
			v.ChannelId = item.Snippet.VideoOwnerChannelId
			v.ChannelTitle = item.Snippet.VideoOwnerChannelTitle
			ytp.Videos = append( ytp.Videos , v )
		}
		ytp_json , _ := json.Marshal( ytp )
		s.DB.Update( func( tx *bolt.Tx ) error {
			b , _ := tx.CreateBucketIfNotExists( []byte( "youtube-playlists" ) )
			b.Put( []byte( playlist_id ) , ytp_json )
			videos_b , _ := tx.CreateBucketIfNotExists( []byte( "youtube-videos" ) )
			for _ , v := range ytp.Videos {
				v_json , _ := json.Marshal( v )
				videos_b.Put( []byte( v.Id ) , v_json )
			}
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

func YouTube_Playlist_Next_New_Session( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		s.LOG.Debug( "YouTube_Playlist_Next_New_Session()" )
		playlist_id := c.Params( "playlist_id" )
		session_id := utils.GenULID()
		// route_info := c.Route()
		// prefix_parts := strings.Split( route_info.Path , "/playlist/:playlist_id" )
		// prefix := prefix_parts[ 0 ]
		// url := fmt.Sprintf( "%s/%s/next" , prefix , session_id )
		// s.LOG.Debug( url )
		now := time.Now().UnixMilli()
		s.DB.Update( func( tx *bolt.Tx ) error {
			sb , _ := tx.CreateBucketIfNotExists( []byte( "youtube-sessions" ) )
			var session YoutubeSession
			session.Id = session_id
			session.Playlist.Id = playlist_id
			session.Accessed = now
			playlist_b := tx.Bucket( []byte( "youtube-playlists" ) )
			playlist_json := playlist_b.Get( []byte( playlist_id ) )
			json.Unmarshal( playlist_json , &session.Playlist )
			s.LOG.Debug( session )
			session_json , _ := json.Marshal( session )
			sb.Put( []byte( session_id ) , []byte( session_json ) )
			return nil
		})

		// return c.Redirect( url )
		return c.RedirectToRoute( "youtube_session", fiber.Map{
			"session_id": session_id ,
		})

		// fiber v3 only
		// return c.Redirect().Route( "youtube_playlist_session_next" , fiber.RedirectConfig{
		// 	Params: fiber.Map{
		// 		"playlist_id": playlist_id ,
		// 		"session_id":  session_id ,
		// 	} ,
		// })
	}
}

func YouTube_Session_Next( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		s.LOG.Debug( "YouTube_Playlist_Next()" )
		session_id := c.Params( "session_id" )
		now := time.Now().UnixMilli()
		// var ytp YoutubePlaylist
		var next_id string
		var next_index int
		var next_position int
		var session YoutubeSession
		s.DB.Update( func( tx *bolt.Tx ) error {
			b := tx.Bucket( []byte( "youtube-sessions" ) )
			session_json := b.Get( []byte( session_id ) )
			if session_json == nil {
				s.LOG.Debug( "session json is empty ??" )
				return nil
			}
			json.Unmarshal( session_json , &session )
			session.Accessed = now
			if session.Playlist.Videos[ session.Playlist.Index ].Position == -1 {
				session.Playlist.Index++
				if session.Playlist.Index >= session.Playlist.Total { // look for off by 1
					session.Playlist.Index = 0
				}
				if session.Playlist.Videos[ session.Playlist.Index ].Position == -1 {
					next_index = session.Playlist.Index
					next_position = 0
				}
				next_id = session.Playlist.Videos[ session.Playlist.Index ].Id
				next_index = session.Playlist.Index
				next_position = session.Playlist.Videos[ session.Playlist.Index ].Position
			} else {
				// continuing where you left off , not finished
				next_id = session.Playlist.Videos[ session.Playlist.Index ].Id
				next_index = session.Playlist.Index
				next_position = session.Playlist.Videos[ session.Playlist.Index ].Position
				s.LOG.Debug( fmt.Sprintf( "continuing where you left off : %s : %d : %d" , next_id , next_index , next_position ) )
			}
			// restore
			session_json , _ = json.Marshal( session )
			b.Put( []byte( session_id ) , session_json )
			return nil
		})
		return c.JSON(fiber.Map{
			"result": true ,
			"session_id": session_id ,
			"playlist_id": session.Playlist.Id ,
			"next_id": next_id ,
			"next_index": next_index ,
			"next_position": next_position ,
		})
	}
}

func YouTube_Session_Update_Position( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		now := time.Now().UnixMilli()
		session_id := c.Params( "session_id" )
		video_id := c.Params( "video_id" )
		position := c.Params( "position" )
		position_int , _ := strconv.Atoi( position )

		s.LOG.Debug( fmt.Sprintf( "YouTube_Update_Position( %s , %s , %s )" , session_id , video_id , position ) )

		var session YoutubeSession
		s.DB.Update( func( tx *bolt.Tx ) error {
			b := tx.Bucket( []byte( "youtube-sessions" ) )
			session_json := b.Get( []byte( session_id ) )
			if session_json == nil {
				s.LOG.Debug( "session json is empty ??" )
				return nil
			}
			json.Unmarshal( session_json , &session )
			session.Accessed = now
			for i , _ := range session.Playlist.Videos {
				if session.Playlist.Videos[ i ].Id != video_id { continue; }
				session.Playlist.Videos[ i ].Position = position_int
				if session.Playlist.Index != i {
					s.LOG.Debug( fmt.Sprintf( "resetting index === %d" , i ) )
					session.Playlist.Index = i
				}
				break
			}
			session_json , _ = json.Marshal( session )
			b.Put( []byte( session_id ) , []byte( session_json ) )
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
		})
	}
}

func YouTube_Session_Get( s *server.Server ) fiber.Handler {
	return func( c *fiber.Ctx ) error {
		session_id := c.Params( "session_id" )
		var session YoutubeSession
		s.DB.View( func( tx *bolt.Tx ) error {
			b := tx.Bucket( []byte( "youtube-sessions" ) )
			session_json := b.Get( []byte( session_id ) )
			if session_json == nil {
				s.LOG.Fatal( "session json is empty ??" )
				return nil
			}
			json.Unmarshal( session_json , &session )
			return nil
		})
		return c.JSON( fiber.Map{
			"result": true ,
			"session": session ,
		})
	}
}