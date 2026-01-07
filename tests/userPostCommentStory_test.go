package tests

import (
	"fmt"
	"i9lyfe/src/appGlobals"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPostCommentStory(t *testing.T) {
	// t.Parallel()
	require := require.New(t)

	user1 := UserT{
		Email:    "harveyspecter@gmail.com",
		Username: "harvey",
		Name:     "Harvey Specter",
		Password: "harvey_psl",
		Birthday: bday("1993-11-07"),
		Bio:      "Whatever!",
	}

	user2 := UserT{
		Email:    "mikeross@gmail.com",
		Username: "mikeross",
		Name:     "Mike Ross",
		Password: "mikeross_psl",
		Birthday: bday("1999-11-07"),
		Bio:      "Whatever!",
	}

	user3 := UserT{
		Email:    "alexwilliams@gmail.com",
		Username: "alex",
		Name:     "Alex Williams",
		Password: "williams_psl",
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
				req.Header.Add("Content-Type", "application/json")

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
				req.Header.Add("Content-Type", "application/json")

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
				req.Header.Add("Content-Type", "application/json")

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
			user := user

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

	t.Log("-----")

	user1Post1Id := ""

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
			"media_sizes": [][2]int64{{blurImageInfo.Size(), actualImageInfo.Size()}},
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/post_upload/authorize", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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
		t.Log("Action: user1 creates post1")

		reqBody, err := makeReqBody(map[string]any{
			"media_cloud_names": mediaCloudNames,
			"type":              "photo:portrait",
			"description":       "This is No.1 #trending",
			"at":                time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/new_post", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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

		user1Post1Id = rb["id"].(string)
	}

	{
		t.Log("Action: user2 reacts to user1's post1 | user1 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"emoji": "🤔",
			"at":    time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1Post1Id+"/react", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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
		t.Log("Action: user3 reacts to user1's post1 | user1 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"emoji": "😀",
			"at":    time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1Post1Id+"/react", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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
					"reactor_user": td.SuperMapOf(map[string]any{"username": user3.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 checks reactors to her post1")

		req, err := http.NewRequest("GET", appPathPriv+"/posts/"+user1Post1Id+"/reactors", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		reactors, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), reactors, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user2.Username,
				"emoji":    "🤔",
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user3.Username,
				"emoji":    "😀",
			}, nil)),
		))
	}

	/* {
		t.Log("Action: user1 filters reactors to her post1 by a certain emoji")


		req, err := http.NewRequest("GET", appPathPriv+"/posts/"+user1Post1Id+"/reactors/🤔", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		reactors, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), reactors, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user2.Username,
				"emoji": "🤔",
			}, nil)),
			td.Not(td.Contains(td.SuperMapOf(map[string]any{
				"username": user3.Username,
				"emoji": "😀",
			}, nil))),
		))
	} */

	{
		t.Log("Action: user3 removes her reaction from user1's post1")

		req, err := http.NewRequest("DELETE", appPathPriv+"/posts/"+user1Post1Id+"/remove_reaction", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)

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
		t.Log("Action: user1 rechecks reactors to her post1 | user3's reaction gone")

		req, err := http.NewRequest("GET", appPathPriv+"/posts/"+user1Post1Id+"/reactors", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		reactors, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), reactors, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user2.Username,
				"emoji":    "🤔",
			}, nil)),
			td.Not(td.Contains(td.SuperMapOf(map[string]any{
				"username": user3.Username,
				"emoji":    "😀",
			}, nil))),
		))
	}

	user2Comment1User1Post1Id := ""

	{
		t.Log("Action: user2 comments on user1's post1 | user1 is notified")

		var (
			uploadUrl           string
			attachmentCloudName string
			imagePath           = "./test_files/attach.jpg"
			contentType         = "image/jpeg"
		)

		imageInfo, err := os.Stat(imagePath)
		require.NoError(err)

		{

			t.Log("--- Authorize attachment upload ---")

			reqBody, err := makeReqBody(map[string]any{
				"attachment_mime": contentType,
				"attachment_size": imageInfo.Size(),
			})
			require.NoError(err)

			req, err := http.NewRequest("POST", appPathPriv+"/comment_upload/authorize", reqBody)
			require.NoError(err)
			req.Header.Set("Cookie", user1.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

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
				"uploadUrl":           td.Ignore(),
				"attachmentCloudName": td.Ignore(),
			}, nil))

			uploadUrl = rb["uploadUrl"].(string)
			attachmentCloudName = rb["attachmentCloudName"].(string)
		}

		{
			t.Log("Upload session started:")

			sessionUrl := startResumableUpload(uploadUrl, contentType, t)

			uploadFileInChunks(sessionUrl, imagePath, contentType, logProgress, t)

			defer func(attCn string) {
				err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(attCn).Delete(t.Context())
				require.NoError(err)
			}(attachmentCloudName)

			t.Log("Upload complete")
		}

		reqBody, err := makeReqBody(map[string]any{
			"comment_text":          fmt.Sprintf("This is a comment from %s", user2.Username),
			"attachment_cloud_name": attachmentCloudName,
			"at":                    time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1Post1Id+"/comment", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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

		user2Comment1User1Post1Id = rb["id"].(string)

		// user1 is notified
		ServerEventMsg := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "comment_on_post",
				"details": td.SuperMapOf(map[string]any{
					"commenter_user": td.SuperMapOf(map[string]any{"username": user2.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	user3Comment1User1Post1Id := ""

	{
		t.Log("Action: user3 comments on user1's post1 | user1 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"comment_text": fmt.Sprintf("This is a comment from %s", user3.Username),
			"at":           time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1Post1Id+"/comment", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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

		user3Comment1User1Post1Id = rb["id"].(string)

		// user1 is notified
		ServerEventMsg := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "comment_on_post",
				"details": td.SuperMapOf(map[string]any{
					"commenter_user": td.SuperMapOf(map[string]any{"username": user3.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 checks comments on her post1")

		req, err := http.NewRequest("GET", appPathPriv+"/posts/"+user1Post1Id+"/comments", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		comments, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), comments, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user2.Username,
				}, nil),
				"comment_text": fmt.Sprintf("This is a comment from %s", user2.Username),
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user3.Username,
				}, nil),
				"comment_text": fmt.Sprintf("This is a comment from %s", user3.Username),
			}, nil)),
		))
	}

	{
		t.Log("Action: user3 removes her comment on user1's post1")

		req, err := http.NewRequest("DELETE", appPathPriv+"/posts/"+user1Post1Id+"/comments/"+user3Comment1User1Post1Id, nil)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)

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

	/* {
		t.Log("Action: user1 rechecks comments on her post1 | user3's comment is gone")


		req, err := http.NewRequest("GET", appPathPriv+"/posts/"+user1Post1Id+"/comments", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		comments, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), comments, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user2.Username,
				}, nil),
				"comment_text": fmt.Sprintf("This is a comment from %s", user2.Username),
			}, nil)),
			td.Not(td.Contains(td.SuperMapOf(map[string]any{
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user3.Username,
				}, nil),
				"comment_text": fmt.Sprintf("This is a comment from %s", user3.Username),
			}, nil))),
		))
	} */

	{
		t.Log("Action: user1 views user2's comment on her post1")

		req, err := http.NewRequest("GET", appPathPriv+"/comments/"+user2Comment1User1Post1Id, nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

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
			"id": user2Comment1User1Post1Id,
		}, nil))
	}

	user1Reply1User2Comment1User1Post1Id := ""

	{
		t.Log("Action: user1 replied to user2's comment on her post1 | user2 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"comment_text": fmt.Sprintf("This is a reply from %s", user1.Username),
			"at":           time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/comments/"+user2Comment1User1Post1Id+"/comment", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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

		user1Reply1User2Comment1User1Post1Id = rb["id"].(string)

		// user2 is notified
		ServerEventMsg := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "comment_on_comment",
				"details": td.SuperMapOf(map[string]any{
					"commenter_user": td.SuperMapOf(map[string]any{"username": user1.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	user3Reply1User2Comment1User1Post1Id := ""

	{
		t.Log("Action: user3 replied to user2's comment on user1's post1 | user2 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"comment_text": fmt.Sprintf("I %s, second %s on this!", user3.Username, user1.Username),
			"at":           time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/comments/"+user2Comment1User1Post1Id+"/comment", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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

		user3Reply1User2Comment1User1Post1Id = rb["id"].(string)

		// user2 is notified
		ServerEventMsg := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "comment_on_comment",
				"details": td.SuperMapOf(map[string]any{
					"commenter_user": td.SuperMapOf(map[string]any{"username": user3.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 checks replies to her comment1 on user1's post1")

		req, err := http.NewRequest("GET", appPathPriv+"/comments/"+user2Comment1User1Post1Id+"/comments", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		replies, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), replies, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user1.Username,
				}, nil),
				"comment_text": fmt.Sprintf("This is a reply from %s", user1.Username),
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"owner_user": td.SuperMapOf(map[string]any{
					"username": user3.Username,
				}, nil),
				"comment_text": fmt.Sprintf("I %s, second %s on this!", user3.Username, user1.Username),
			}, nil)),
		))
	}

	{
		t.Log("Action: user3 removes her reply to user2's comment1 on user1's post1")

		req, err := http.NewRequest("DELETE", appPathPriv+"/comments/"+user2Comment1User1Post1Id+"/comments/"+user3Reply1User2Comment1User1Post1Id, nil)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)

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

	/*
		{
			t.Log("Action: user2 rechecks replies to her comment1 on user1's post1 | user3's reply is gone")


			req, err := http.NewRequest("GET", appPathPriv+"/comments/"+user2Comment1User1Post1Id+"/comments", nil)
			require.NoError(err)
			req.Header.Set("Cookie", user2.SessionCookie)

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			reactors, err := succResBody[[]map[string]any](res.Body)
			require.NoError(err)

			td.Cmp(td.Require(t), reactors, td.All(
				td.Contains(td.SuperMapOf(map[string]any{
					"owner_user": td.SuperMapOf(map[string]any{
						"username": user1.Username,
					}, nil),
					"comment_text": fmt.Sprintf("This is a reply from %s", user1.Username),
				}, nil)),
				td.Not(td.Contains(td.SuperMapOf(map[string]any{
					"owner_user": td.SuperMapOf(map[string]any{
						"username": user3.Username,
					}, nil),
					"comment_text": fmt.Sprintf("I %s, second %s on this!", user3.Username, user1.Username),
				}, nil))),
			))
		} */

	{
		t.Log("Action: user2 reacts to user1's reply to her comment1 on user1's post1 | user1 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"emoji": "😆",
			"at":    time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/comments/"+user1Reply1User2Comment1User1Post1Id+"/react", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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
				"type": "reaction_to_comment",
				"details": td.SuperMapOf(map[string]any{
					"reactor_user": td.SuperMapOf(map[string]any{"username": user2.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user3 reacts to user1's reply to user2's comment1 on user1's post1 | user1 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"emoji": "😂",
			"at":    time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/comments/"+user1Reply1User2Comment1User1Post1Id+"/react", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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
				"type": "reaction_to_comment",
				"details": td.SuperMapOf(map[string]any{
					"reactor_user": td.SuperMapOf(map[string]any{"username": user3.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 checks reactors to her reply to user2's comment1 on her post1")

		req, err := http.NewRequest("GET", appPathPriv+"/comments/"+user1Reply1User2Comment1User1Post1Id+"/reactors", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		reactors, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), reactors, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user2.Username,
				"emoji":    "😆",
			}, nil)),
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user3.Username,
				"emoji":    "😂",
			}, nil)),
		))
	}

	/* {
		t.Log("Action: user1 filters reactors to her reply to user2's comment1 on her post1 by a certain emoji")


		req, err := http.NewRequest("GET", appPathPriv+"/comments/"+user1Reply1User2Comment1User1Post1Id+"/reactors/😆", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		reactors, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), reactors, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user2.Username,
				"emoji":    "😆",
			}, nil)),
			td.Not(td.Contains(td.SuperMapOf(map[string]any{
				"username": user3.Username,
				"emoji":    "😂",
			}, nil))),
		))
	} */

	{
		t.Log("Action: user3 removes her reaction to user1's reply to user2's comment1 on user1's post1")

		req, err := http.NewRequest("DELETE", appPathPriv+"/comments/"+user1Reply1User2Comment1User1Post1Id+"/remove_reaction", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)

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
		t.Log("Action: user1 rechecks reactors to her reply to user2's comment1 on her post1 | user3's reaction gone")

		req, err := http.NewRequest("GET", appPathPriv+"/comments/"+user1Reply1User2Comment1User1Post1Id+"/reactors", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		reactors, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), reactors, td.All(
			td.Contains(td.SuperMapOf(map[string]any{
				"username": user2.Username,
				"emoji":    "😆",
			}, nil)),
			td.Not(td.Contains(td.SuperMapOf(map[string]any{
				"username": user3.Username,
				"emoji":    "😂",
			}, nil))),
		))
	}

	user1Post2Id := ""

	{
		t.Log("Action: user1 creates post2 mentioning user2 | user2 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"media_cloud_names": mediaCloudNames,
			"type":              "photo:portrait",
			"description":       fmt.Sprintf("This is a post mentioning @%s", user2.Username),
			"at":                time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/new_post", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"id": td.Ignore(),
			}, nil))

		user1Post2Id = rb["id"].(string)

		ServerEventMsg := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), ServerEventMsg, td.Map(map[string]any{
			"event": "new notification",
			"data": td.SuperMapOf(map[string]any{
				"id":   td.Ignore(),
				"type": "mention_in_post",
				"details": td.SuperMapOf(map[string]any{
					"mentioning_user": td.SuperMapOf(map[string]any{"username": user1.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 views user1's post2 following the mention notification received")

		req, err := http.NewRequest("GET", appPathPriv+"/posts/"+user1Post2Id, nil)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)

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
			"id": user1Post2Id,
		}, nil))
	}

	user3Post1Id := ""

	{
		t.Log("Action: user3 creates post1")

		reqBody, err := makeReqBody(map[string]any{
			"media_cloud_names": mediaCloudNames,
			"type":              "photo:portrait",
			"description":       "I'm beautiful",
			"at":                time.Now().UnixMilli(),
		})
		require.NoError(err)

		req, err := http.NewRequest("POST", appPathPriv+"/new_post", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user3.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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

		user3Post1Id = rb["id"].(string)
	}

	{
		t.Log("Action: user2 reposts user3's post1 | user3 is notified")

		req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user3Post1Id+"/repost", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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
				"type": "repost",
				"details": td.SuperMapOf(map[string]any{
					"reposter_user": td.SuperMapOf(map[string]any{"username": user2.Username}, nil),
				}, nil),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 saves user3's post1")

		req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user3Post1Id+"/save", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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
		t.Log("Action: user1 unsaves user3's post1")

		req, err := http.NewRequest("DELETE", appPathPriv+"/posts/"+user3Post1Id+"/unsave", nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

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
