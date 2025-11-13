package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPersonalOperationsStory(t *testing.T) {
	t.Parallel()

	user1 := UserT{
		Email:    "robertzane@gmail.com",
		Username: "robertzane",
		Name:     "Robert Zane",
		Password: "robert_laura",
		Birthday: bday("1993-11-07"),
		Bio:      "Whatever!",
	}

	user2 := UserT{
		Email:    "samanthawheeler@gmail.com",
		Username: "samantha_wheel",
		Name:     "Samantha Wheeler",
		Password: "wheel_saman",
		Birthday: bday("1999-11-07"),
		Bio:      "Whatever!",
	}

	user3 := UserT{
		Email:    "katrinabennet@gmail.com",
		Username: "katrina",
		Name:     "Katrina Bennet",
		Password: "katrina_ben",
		Birthday: bday("1999-11-07"),
		Bio:      "Whatever!",
	}

	{
		t.Log("Setup: create new account for users")

		for _, user := range []*UserT{&user1, &user2, &user3} {
			{
				reqBody, err := makeReqBody(map[string]any{"email": user.Email})
				require.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/request_new_account", reqBody)
				require.NoError(t, err)
				req.Header.Add("Content-Type", "application/json")

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(t, err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(t, err)

				td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
					"msg": fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", user.Email),
				}, nil))

				user.SessionCookie = res.Header.Get("Set-Cookie")
			}

			{
				verfCode := os.Getenv("DUMMY_TOKEN")

				reqBody, err := makeReqBody(map[string]any{"code": verfCode})
				require.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
				require.NoError(t, err)
				req.Header.Set("Cookie", user.SessionCookie)
				req.Header.Add("Content-Type", "application/json")

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(t, err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(t, err)

				td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
					"msg": fmt.Sprintf("Your email, %s, has been verified!", user.Email),
				}, nil))

				user.SessionCookie = res.Header.Get("Set-Cookie")
			}

			{
				reqBody, err := makeReqBody(map[string]any{
					"username": user.Username,
					"name":     user.Name,
					"password": user.Password,
					"birthday": user.Birthday,
					"bio":      user.Bio,
				})
				require.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
				require.NoError(t, err)
				req.Header.Set("Cookie", user.SessionCookie)
				req.Header.Add("Content-Type", "application/json")

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(t, err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(t, err)

				td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
					"user": td.Ignore(),
					"msg":  "Signup success!",
				}, nil))

				user.SessionCookie = res.Header.Get("Set-Cookie")
			}
		}
	}

	{
		t.Log("Setup: Init user sockets")

		for _, user := range []*UserT{&user1, &user2, &user3} {

			header := http.Header{}
			header.Set("Cookie", user.SessionCookie)
			wsConn, res, err := websocket.DefaultDialer.Dial(wsPath, header)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusSwitchingProtocols, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			require.NotNil(t, wsConn)

			defer wsConn.CloseHandler()(websocket.CloseNormalClosure, user.Username+": GoodBye!")

			user.WSConn = wsConn
			user.ServerEventMsg = make(chan map[string]any)

			go func() {
				userCommChan := user.ServerEventMsg

				for {
					userCommChan := userCommChan
					userWSConn := user.WSConn

					var wsMsg map[string]any

					if err := userWSConn.ReadJSON(&wsMsg); err != nil {
						break
					}

					if wsMsg == nil {
						continue
					}

					userCommChan <- wsMsg
				}

				close(userCommChan)
			}()
		}
	}

	t.Log("----------")

	{
		t.Log("Action: user1 edits his profile")

		user1.Birthday = bday("1992-04-29")
		user1.Bio = "Editing profile..."
		user1.Name = "Zane Robert"

		reqBody, err := makeReqBody(map[string]any{
			"birthday": user1.Birthday,
			"bio":      user1.Bio,
			"name":     user1.Name,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", appPathPriv+"/me/edit_profile", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)
	}

	{
		t.Log("Action: user1 changes his profile picture")

		ppic, err := os.ReadFile("./test_files/profile_pic.png")
		require.NoError(t, err)

		reqBody, err := makeReqBody(map[string]any{"picture_data": ppic})
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", appPathPriv+"/me/change_profile_picture", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)
	}

	{
		t.Log("Action: user1 follows user2 | user2 is notified")

		req, err := http.NewRequest("POST", appPathPriv+"/users/"+user2.Username+"/follow", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)

		ServerEventMsg := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "user_follow",
				"details": td.SuperMapOf(map[string]any{
					"follower_user": td.SuperMapOf(map[string]any{"username": user1.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user3 follows user2 | user2 is notified")

		req, err := http.NewRequest("POST", appPathPriv+"/users/"+user2.Username+"/follow", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)

		ServerEventMsg := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "user_follow",
				"details": td.SuperMapOf(map[string]any{
					"follower_user": td.SuperMapOf(map[string]any{"username": user3.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 follows user3 | user3 is notified")

		req, err := http.NewRequest("POST", appPathPriv+"/users/"+user3.Username+"/follow", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)

		ServerEventMsg := <-user3.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "user_follow",
				"details": td.SuperMapOf(map[string]any{
					"follower_user": td.SuperMapOf(map[string]any{"username": user2.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 checks her followers | confirms new followers")
		<-(time.NewTimer(100 * time.Millisecond)).C

		req, err := http.NewRequest("GET", appPathPublic+"/"+user2.Username+"/followers", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		followers, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), followers, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username":  user1.Username,
				"me_follow": false,
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"username":  user3.Username,
				"me_follow": true,
			}, nil)),
		))
	}

	{
		t.Log("Action: user3 follows user1 | user1 is notified")

		req, err := http.NewRequest("POST", appPathPriv+"/users/"+user1.Username+"/follow", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)

		ServerEventMsg := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "user_follow",
				"details": td.SuperMapOf(map[string]any{
					"follower_user": td.SuperMapOf(map[string]any{"username": user3.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user3 checks her following | confirms new following")
		<-(time.NewTimer(100 * time.Millisecond)).C

		req, err := http.NewRequest("GET", appPathPublic+"/"+user3.Username+"/followings", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		following, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), following, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username":  user1.Username,
				"me_follow": true,
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"username":  user2.Username,
				"me_follow": true,
			}, nil)),
		))
	}

	{
		t.Log("Action: user3 unfollows user2")

		req, err := http.NewRequest("DELETE", appPathPriv+"/users/"+user2.Username+"/unfollow", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)
	}

	{
		t.Log("Action: user2 rechecks her followers | confirms user3's gone")
		<-(time.NewTimer(100 * time.Millisecond)).C

		req, err := http.NewRequest("GET", appPathPublic+"/"+user2.Username+"/followers", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		followers, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), followers, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user1.Username,
			}, nil)),
			td.Not(td.Contains(td.SuperMapOf(map[string]any{
				"username": user3.Username,
			}, nil))),
		))
	}

	{
		t.Log("Action: user3 rechecks her following | confirms user2's gone")
		<-(time.NewTimer(100 * time.Millisecond)).C

		req, err := http.NewRequest("GET", appPathPublic+"/"+user3.Username+"/followings", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		following, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), following, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user1.Username,
			}, nil)),
			td.Not(td.Contains(td.SuperMapOf(map[string]any{
				"username": user2.Username,
			}, nil))),
		))
	}

	{
		t.Log("Action: user1 views his profile | confirms all changes")
		<-(time.NewTimer(100 * time.Millisecond)).C

		req, err := http.NewRequest("GET", appPathPublic+"/"+user1.Username, nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		profile, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), profile, td.SuperMapOf(map[string]any{
			"username":         user1.Username,
			"name":             user1.Name,
			"bio":              user1.Bio,
			"posts_count":      td.Lax(0),
			"followers_count":  td.Lax(1),
			"followings_count": td.Lax(1),
		}, nil))
	}

	t.Log("--------")

	user1PostId, user3PostId := "", ""

	{
		t.Log("Mid-setup: user1 and user3 creates post mentioning user2 | user2 is notified for both | user2 reacts to and saves both post")

		{
			//Action: user1 creates a post mentioning user2 | user2 is notified

			photo1, err := os.ReadFile("./test_files/photo_1.png")
			require.NoError(t, err)

			photo2, err := os.ReadFile("./test_files/photo_1.png")
			require.NoError(t, err)

			reqBody, err := makeReqBody(map[string]any{
				"media_data_list": [][]byte{photo1, photo2},
				"type":            "photo",
				"description":     fmt.Sprintf("This is a post by @%s mentioning @%s", user1.Username, user2.Username),
				"at":              time.Now().UnixMilli(),
			})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", appPathPriv+"/new_post", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", user1.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(t, err)

			td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
				"id": td.Ignore(),
			}, nil))

			user1PostId = rb["id"].(string)

			ServerEventMsg_mentionNotif := <-user2.ServerEventMsg

			td.Cmp(td.Require(t), ServerEventMsg_mentionNotif, td.Map(map[string]any{
				"event": "new notification",
				"data": td.SuperMapOf(map[string]any{
					"id":   td.Ignore(),
					"type": "mention_in_post",
					"details": td.SuperMapOf(map[string]any{
						"mentioning_user": td.SuperMapOf(map[string]any{"username": user1.Username}, nil),
					}, nil),
				}, nil),
			}, nil))

			/* --- CONTENT RECOMMENDATION SYSTEM not yet implemented ---

			// user2 receives this post in her home feed | due to her follow network
			user2_ServerEventMsg_newPost := <-user2.ServerEventMsg

			td.Cmp(td.Require(t), user2_ServerEventMsg_newPost, td.SuperMapOf(map[string]any{
				"id": user1PostId,
			}, nil))

			// user3 also receives this post in her home feed | due to her follow network
			user3_ServerEventMsg_newPost := <-user3.ServerEventMsg

			td.Cmp(td.Require(t), user3_ServerEventMsg_newPost, td.SuperMapOf(map[string]any{
				"id": user1PostId,
				}, nil))

			*/
		}

		{
			//Action: user3 creates a post mentioning user2 | user2 is notified

			photo1, err := os.ReadFile("./test_files/photo_1.png")
			require.NoError(t, err)

			photo2, err := os.ReadFile("./test_files/photo_1.png")
			require.NoError(t, err)

			reqBody, err := makeReqBody(map[string]any{
				"media_data_list": [][]byte{photo1, photo2},
				"type":            "photo",
				"description":     fmt.Sprintf("This is a post from @%s mentioning @%s", user3.Username, user2.Username),
				"at":              time.Now().UnixMilli(),
			})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", appPathPriv+"/new_post", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", user3.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(t, err)

			td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
				"id": td.Ignore(),
			}, nil))

			user3PostId = rb["id"].(string)

			ServerEventMsg_mentionNotif := <-user2.ServerEventMsg

			td.Cmp(td.Require(t), ServerEventMsg_mentionNotif, td.Map(map[string]any{
				"event": "new notification",
				"data": td.SuperMapOf(map[string]any{
					"id":   td.Ignore(),
					"type": "mention_in_post",
					"details": td.SuperMapOf(map[string]any{
						"mentioning_user": td.SuperMapOf(map[string]any{"username": user3.Username}, nil),
					}, nil),
				}, nil),
			}, nil))

			/* --- CONTENT RECOMMENDATION SYSTEM not yet implemented ---

			// user2 receives this post in her home feed | due to her follow network
			user2_ServerEventMsg_newPost := <-user2.ServerEventMsg

			td.Cmp(td.Require(t), user2_ServerEventMsg_newPost, td.SuperMapOf(map[string]any{
				"id": user3PostId,
			}, nil))

			// user1 also receives this post in her home feed | due to his follow network
			user1_ServerEventMsg_newPost := <-user1.ServerEventMsg

			td.Cmp(td.Require(t), user1_ServerEventMsg_newPost, td.SuperMapOf(map[string]any{
				"id": user3PostId,
			}, nil))

			*/
		}

		{
			// Action: user2 reacts to user1's post | user1 is notified

			reqBody, err := makeReqBody(map[string]any{
				"reaction": "🤔",
				"at":       time.Now().UnixMilli(),
			})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1PostId+"/react", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[bool](res.Body)
			require.NoError(t, err)
			require.True(t, rb)

			// user1 is notified
			ServerEventMsg := <-user1.ServerEventMsg

			td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
				"event": "new notification",
				"data": td.SuperMapOf(map[string]any{
					"id":   td.Ignore(),
					"type": "reaction_to_post",
					"details": td.SuperMapOf(map[string]any{
						"reactor_user": td.SuperMapOf(map[string]any{"username": user2.Username}, nil),
					}, nil),
				}, nil),
			}, nil))
		}

		{
			// Action: user2 reacts to user3's post | user3 is notified

			reqBody, err := makeReqBody(map[string]any{
				"reaction": "🤔",
				"at":       time.Now().UnixMilli(),
			})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user3PostId+"/react", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[bool](res.Body)
			require.NoError(t, err)
			require.True(t, rb)

			// user3 is notified
			ServerEventMsg := <-user3.ServerEventMsg

			td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
				"event": "new notification",
				"data": td.SuperMapOf(map[string]any{
					"id":   td.Ignore(),
					"type": "reaction_to_post",
					"details": td.SuperMapOf(map[string]any{
						"reactor_user": td.SuperMapOf(map[string]any{"username": user2.Username}, nil),
					}, nil),
				}, nil),
			}, nil))
		}

		{
			// Action: user2 saves user1's post

			req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1PostId+"/save", nil)
			require.NoError(t, err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[bool](res.Body)
			require.NoError(t, err)
			require.True(t, rb)
		}

		{
			// Action: user2 saves user3's post

			req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user3PostId+"/save", nil)
			require.NoError(t, err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[bool](res.Body)
			require.NoError(t, err)
			require.True(t, rb)
		}
	}

	t.Log("--------")

	{
		t.Log("Action: user2 checks posts in which she's been mentioned")

		req, err := http.NewRequest("GET", appPathPriv+"/me/mentioned_posts", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		posts, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), posts, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"id": user1PostId,
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user1.Username,
				}, nil),
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"id": user3PostId,
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user3.Username,
				}, nil),
			}, nil)),
		))
	}

	{
		t.Log("Action: user2 checks posts she's reacted to")

		req, err := http.NewRequest("GET", appPathPriv+"/me/reacted_posts", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		posts, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), posts, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"id": user1PostId,
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user1.Username,
				}, nil),
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"id": user3PostId,
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user3.Username,
				}, nil),
			}, nil)),
		))
	}

	{
		t.Log("Action: user2 checks posts she's saved")

		req, err := http.NewRequest("GET", appPathPriv+"/me/saved_posts", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		posts, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), posts, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"id": user1PostId,
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user1.Username,
				}, nil),
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"id": user3PostId,
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user3.Username,
				}, nil),
			}, nil)),
		))
	}
}
