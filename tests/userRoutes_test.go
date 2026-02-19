package tests

import (
	"fmt"
	"i9lyfe/src/appGlobals"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserProfile(t *testing.T) {
	teardown, err := getUserProfileSetup(t.Context(), user1, user2)
	require.NoError(t, err)

	t.Run("view profile: [client not logged in | user has followers]", func(t *testing.T) {
		req := httptest.NewRequest("GET", appPathPublic+"/"+user1.Username, nil)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := app.Test(req)
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
			"followings_count": td.Lax(0),
			"me_follow":        false,
			"follows_me":       false,
		}, nil))
	})

	t.Run("view profile: [user follows client]", func(t *testing.T) {
		reqBody, err := makeReqBody(map[string]any{"username": user1.Username})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", testSessionPath+"/auth_user", reqBody)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := app.Test(req)
		require.NoError(t, err)

		/* ---------- */

		req = httptest.NewRequest("GET", appPathPublic+"/"+user2.Username, nil)
		req.Header.Add("Cookie", res.Header.Get("Set-Cookie"))
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err = app.Test(req)
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
			"username":         user2.Username,
			"name":             user2.Name,
			"bio":              user2.Bio,
			"posts_count":      td.Lax(0),
			"followers_count":  td.Lax(0),
			"followings_count": td.Lax(1),
			"me_follow":        false,
			"follows_me":       true,
		}, nil))
	})

	err = teardown(t.Context())
	require.NoError(t, err)
}

func XTestUserPersonalOperationsStory(t *testing.T) {
	// t.Parallel()
	require := require.New(t)

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
				require.NoError(err)

				req, err := http.NewRequest("POST", signupPath+"/request_new_account", reqBody)
				require.NoError(err)
				req.Header.Add("Content-Type", "application/vnd.msgpack")

				res, err := http.DefaultClient.Do(req)
				require.NoError(err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(err)

				td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
					"msg": fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", user.Email),
				}, nil))

				user.SessionCookie = res.Header.Get("Set-Cookie")
			}

			{
				verfCode := os.Getenv("DUMMY_TOKEN")

				reqBody, err := makeReqBody(map[string]any{"code": verfCode})
				require.NoError(err)

				req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
				require.NoError(err)
				req.Header.Set("Cookie", user.SessionCookie)
				req.Header.Add("Content-Type", "application/vnd.msgpack")

				res, err := http.DefaultClient.Do(req)
				require.NoError(err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(err)

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
				require.NoError(err)

				req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
				require.NoError(err)
				req.Header.Set("Cookie", user.SessionCookie)
				req.Header.Add("Content-Type", "application/vnd.msgpack")

				res, err := http.DefaultClient.Do(req)
				require.NoError(err)

				if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(err)

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
			require.NoError(err)

			if !assert.Equal(t, http.StatusSwitchingProtocols, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
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
		require.NoError(err)

		req, err := http.NewRequest("PUT", appPathPriv+"/me/edit_profile", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)
	}

	{
		t.Log("Action: user1 changes his profile picture")

		var (
			uploadUrl           string
			profilePicCloudName string
			filePath            = "./test_files/profile_pic.png"
			contentType         = "image/png"
		)

		{
			fileInfo, err := os.Stat(filePath)
			require.NoError(err)

			t.Log("--- Authorize profile picture upload ---")

			reqBody, err := makeReqBody(map[string]any{"pic_mime": contentType, "pic_size": [3]int64{fileInfo.Size(), fileInfo.Size(), fileInfo.Size()}})
			require.NoError(err)

			req, err := http.NewRequest("POST", appPathPriv+"/me/profile_pic_upload/authorize", reqBody)
			require.NoError(err)
			req.Header.Set("Cookie", user1.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(err)

			td.Cmp(td.Require(t), rb, td.Map(map[string]any{
				"uploadUrl":           td.Ignore(),
				"profilePicCloudName": td.Ignore(),
			}, nil))

			uploadUrl = rb["uploadUrl"].(string)
			profilePicCloudName = rb["profilePicCloudName"].(string)
		}

		{
			t.Log("Upload session started:")

			varUploadUrl := make([]string, 3)
			_, err := fmt.Sscanf(uploadUrl, "small:%s medium:%s large:%s", &varUploadUrl[0], &varUploadUrl[1], &varUploadUrl[2])
			require.NoError(err)

			for i, smlUploadUrl := range varUploadUrl {
				varSize := []string{"small", "medium", "large"}

				t.Logf("Uploading %s profile pic started", varSize[i])

				sessionUrl := startResumableUpload(smlUploadUrl, contentType, t)

				uploadFileInChunks(sessionUrl, filePath, contentType, logProgress, t)

				t.Logf("Uploading %s profile pic complete", varSize[i])
			}

			defer func(ppcn string) {
				varPPicCloudName := make([]string, 3)
				_, err = fmt.Sscanf(ppcn, "small:%s medium:%s large:%s", &varPPicCloudName[0], &varPPicCloudName[1], &varPPicCloudName[2])
				require.NoError(err)

				for _, smlPPicCn := range varPPicCloudName {
					err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(smlPPicCn).Delete(t.Context())
					require.NoError(err)
				}
			}(profilePicCloudName)

			t.Log("Upload complete")
		}

		reqBody, err := makeReqBody(map[string]any{"profile_pic_cloud_name": profilePicCloudName})
		require.NoError(err)

		req, err := http.NewRequest("PUT", appPathPriv+"/me/change_profile_picture", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)
	}

	{
		t.Log("Action: user1 follows user2 | user2 is notified")

		req, err := http.NewRequest("POST", appPathPriv+"/users/"+user2.Username+"/follow", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)

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
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)

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
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)

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

		req, err := http.NewRequest("GET", appPathPublic+"/"+user2.Username+"/followers", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		followers, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

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
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)

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

		req, err := http.NewRequest("GET", appPathPublic+"/"+user3.Username+"/followings", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		following, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

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
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)
	}

	{

		<-(time.NewTimer(200 * time.Millisecond).C)

		t.Log("Action: user2 rechecks her followers | confirms user3's gone")

		req, err := http.NewRequest("GET", appPathPublic+"/"+user2.Username+"/followers", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		followers, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

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
		<-(time.NewTimer(200 * time.Millisecond).C)

		t.Log("Action: user3 rechecks her following | confirms user2's gone")

		req, err := http.NewRequest("GET", appPathPublic+"/"+user3.Username+"/followings", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		following, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

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

		req, err := http.NewRequest("GET", appPathPublic+"/"+user1.Username, nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		profile, err := succResBody[map[string]any](res.Body)
		require.NoError(err)

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

		var (
			uploadUrls      []string
			mediaCloudNames []string
			blurImagePath   = "./test_files/photo_1_blur.png"
			actualImagePath = "./test_files/photo_1.png"
			contentType     = "image/png"
		)

		blurImageInfo, err := os.Stat(blurImagePath)
		require.NoError(err)
		actualImageInfo, err := os.Stat(actualImagePath)
		require.NoError(err)

		{

			t.Log("--- Authorize post media upload ---")

			reqBody, err := makeReqBody(map[string]any{
				"post_type":   "photo:portrait",
				"media_mime":  [2]string{contentType, contentType},
				"media_sizes": [][2]int64{{blurImageInfo.Size(), actualImageInfo.Size()}, {blurImageInfo.Size(), actualImageInfo.Size()}},
			})
			require.NoError(err)

			req, err := http.NewRequest("POST", appPathPriv+"/post_upload/authorize", reqBody)
			require.NoError(err)
			req.Header.Set("Cookie", user1.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[[]map[string]any](res.Body)
			require.NoError(err)

			td.Cmp(td.Require(t), rb, td.ArrayEach(td.SuperMapOf(map[string]any{
				"uploadUrl":      td.Ignore(),
				"mediaCloudName": td.Ignore(),
			}, nil)))

			for _, it := range rb {
				uploadUrls = append(uploadUrls, it["uploadUrl"].(string))
				mediaCloudNames = append(mediaCloudNames, it["mediaCloudName"].(string))
			}
		}

		{
			t.Log("Upload session started:")

			for uui, uploadUrl := range uploadUrls {
				varUploadUrl := make([]string, 2)
				_, err := fmt.Sscanf(uploadUrl, "blur_placeholder:%s actual:%s", &varUploadUrl[0], &varUploadUrl[1])
				require.NoError(err)

				for i, baUploadUrl := range varUploadUrl {
					varMedia := []string{"blur_placeholder", "actual"}
					varPath := []string{blurImagePath, actualImagePath}

					t.Logf("Uploading %s post media started", varMedia[i])

					sessionUrl := startResumableUpload(baUploadUrl, contentType, t)

					uploadFileInChunks(sessionUrl, varPath[i], contentType, logProgress, t)

					t.Logf("Uploading %s post media complete", varMedia[i])
				}

				defer func(mcn string) {
					varMediaCloudName := make([]string, 2)
					_, err = fmt.Sscanf(mcn, "blur_placeholder:%s actual:%s", &varMediaCloudName[0], &varMediaCloudName[1])
					require.NoError(err)

					for _, baMcn := range varMediaCloudName {
						err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(baMcn).Delete(t.Context())
						require.NoError(err)
					}
				}(mediaCloudNames[uui])
			}

			t.Log("Upload complete")
		}

		{
			//Action: user1 creates a post mentioning user2 | user2 is notified

			reqBody, err := makeReqBody(map[string]any{
				"media_cloud_names": mediaCloudNames,
				"type":              "photo:portrait",
				"description":       fmt.Sprintf("This is a post by @%s mentioning @%s", user1.Username, user2.Username),
				"at":                time.Now().UnixMilli(),
			})
			require.NoError(err)

			req, err := http.NewRequest("POST", appPathPriv+"/new_post", reqBody)
			require.NoError(err)
			req.Header.Set("Cookie", user1.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(err)

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

			reqBody, err := makeReqBody(map[string]any{
				"media_cloud_names": mediaCloudNames,
				"type":              "photo:portrait",
				"description":       fmt.Sprintf("This is a post from @%s mentioning @%s", user3.Username, user2.Username),
				"at":                time.Now().UnixMilli(),
			})
			require.NoError(err)

			req, err := http.NewRequest("POST", appPathPriv+"/new_post", reqBody)
			require.NoError(err)
			req.Header.Set("Cookie", user3.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(err)

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
				"emoji": "🤔",
				"at":    time.Now().UnixMilli(),
			})
			require.NoError(err)

			req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1PostId+"/react", reqBody)
			require.NoError(err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[bool](res.Body)
			require.NoError(err)
			require.True(rb)

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
				"emoji": "🤔",
				"at":    time.Now().UnixMilli(),
			})
			require.NoError(err)

			req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user3PostId+"/react", reqBody)
			require.NoError(err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[bool](res.Body)
			require.NoError(err)
			require.True(rb)

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
			require.NoError(err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[bool](res.Body)
			require.NoError(err)
			require.True(rb)
		}

		{
			// Action: user2 saves user3's post

			req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user3PostId+"/save", nil)
			require.NoError(err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[bool](res.Body)
			require.NoError(err)
			require.True(rb)
		}
	}

	t.Log("--------")

	{
		t.Log("Action: user2 checks posts in which she's been mentioned")

		req, err := http.NewRequest("GET", appPathPriv+"/me/mentioned_posts", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		posts, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

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
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		posts, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

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
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		posts, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

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
