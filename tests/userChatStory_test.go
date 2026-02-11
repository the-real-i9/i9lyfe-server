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

func XTestUserChatStory(t *testing.T) {
	// t.Parallel()
	require := require.New(t)

	user1 := UserT{
		Email:    "harrydasouza@gmail.com",
		Username: "harry",
		Name:     "Harry Da Souza",
		Password: "harry_dasou",
		Birthday: bday("1993-11-07"),
		Bio:      "Whatever!",
	}

	user2 := UserT{
		Email:    "conradharrigan@gmail.com",
		Username: "conrad",
		Name:     "Conrad Harrigan",
		Password: "grandpa_harr",
		Birthday: bday("1999-11-07"),
		Bio:      "Whatever!",
	}

	{
		t.Log("Setup: create new account for users")

		for _, user := range []*UserT{&user1, &user2} {
			user := user

			{
				reqBody, err := makeReqBody(map[string]any{"email": user.Email})
				require.NoError(err)

				res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
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

		for _, user := range []*UserT{&user1, &user2} {
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

	t.Log("----------")

	user1NewMsgId := ""

	{
		t.Log("Action: user1 sends message to user2")

		err := user1.WSConn.WriteJSON(map[string]any{
			"action": "chat: send message",
			"data": map[string]any{
				"msg": map[string]any{
					"type": "text",
					"props": map[string]any{
						"text_content": "Hi. How're you doing?",
					},
				},
				"toUser": user2.Username,
				"at":     time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		// user1's server reply (response) to event sent
		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: send message",
			"data": td.Map(map[string]any{
				"new_msg_id": td.Ignore(),
			}, nil),
		}, nil))

		user1NewMsgId = user1ServerReply["data"].(map[string]any)["new_msg_id"].(string)
	}

	{
		t.Log("Action: user2 receives the message | acknowledges 'delivered'")

		user2NewMsgReceived := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2NewMsgReceived, td.Map(map[string]any{
			"event": "chat: new message",
			"data": td.SuperMapOf(map[string]any{
				"id": user1NewMsgId,
				"content": td.SuperMapOf(map[string]any{
					"type": "text",
					"props": td.SuperMapOf(map[string]any{
						"text_content": "Hi. How're you doing?",
					}, nil),
				}, nil),
				"delivery_status": "sent",
				"sender": td.SuperMapOf(map[string]any{
					"username": user1.Username,
				}, nil),
			}, nil),
		}, nil))

		err := user2.WSConn.WriteJSON(map[string]any{
			"action": "chat: ack message delivered",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: ack message delivered",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user1 receives the 'delivered' acknowledgement | marks message as 'delivered'")

		user1DelvAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1DelvAckReceipt, td.Map(map[string]any{
			"event": "chat: message delivered",
			"data": td.SuperMapOf(map[string]any{
				"partner_username": user2.Username,
				"msg_id":           user1NewMsgId,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 then acknowledges 'read'")

		err := user2.WSConn.WriteJSON(map[string]any{
			"action": "chat: ack message read",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: ack message read",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user1 receives the 'read' acknowledgement | marks message as 'read'")

		user1ReadAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ReadAckReceipt, td.Map(map[string]any{
			"event": "chat: message read",
			"data": td.SuperMapOf(map[string]any{
				"partner_username": user2.Username,
				"msg_id":           user1NewMsgId,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 reacts to user1's message")

		err := user2.WSConn.WriteJSON(map[string]any{
			"action": "chat: react to message",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"emoji":           "🚀",
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		// user2's server reply (response) to event sent
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: react to message",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user1 receives user2's reaction to his message | attaches it to chat snippet")

		user1ReadAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ReadAckReceipt, td.Map(map[string]any{
			"event": "chat: message reaction",
			"data": td.SuperMapOf(map[string]any{
				"partner_username": user2.Username,
				"to_msg_id":        user1NewMsgId,
				"reaction": td.Map(map[string]any{
					"emoji":   "🚀",
					"reactor": td.Ignore(),
				}, nil),
			}, nil),
		}, nil))
	}

	user2NewMsgId := ""

	{
		t.Log("Action: user2 sends message to user1")

		var (
			uploadUrl       string
			mediaCloudName  string
			blurImagePath   = "./test_files/photo_2_blur.jpg"
			actualImagePath = "./test_files/photo_2.jpg"
			contentType     = "image/jpeg"
		)

		blurImageInfo, err := os.Stat(blurImagePath)
		require.NoError(err)
		actualImageInfo, err := os.Stat(actualImagePath)
		require.NoError(err)

		{

			t.Log("--- Authorize message media upload ---")

			reqBody, err := makeReqBody(map[string]any{
				"msg_type":   "photo",
				"media_mime": [2]string{contentType, contentType},
				"media_size": [2]int64{blurImageInfo.Size(), actualImageInfo.Size()},
			})
			require.NoError(err)

			req, err := http.NewRequest("POST", appPathPriv+"/chat_upload/authorize/visual", reqBody)
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
				"uploadUrl":      td.Ignore(),
				"mediaCloudName": td.Ignore(),
			}, nil))

			uploadUrl = rb["uploadUrl"].(string)
			mediaCloudName = rb["mediaCloudName"].(string)
		}

		{
			t.Log("Upload session started:")

			varUploadUrl := make([]string, 2)
			_, err := fmt.Sscanf(uploadUrl, "blur_placeholder:%s actual:%s", &varUploadUrl[0], &varUploadUrl[1])
			require.NoError(err)

			for i, baUploadUrl := range varUploadUrl {
				varMedia := []string{"blur_placeholder", "actual"}
				varPath := []string{blurImagePath, actualImagePath}

				t.Logf("Uploading %s message media started", varMedia[i])

				sessionUrl := startResumableUpload(baUploadUrl, contentType, t)

				uploadFileInChunks(sessionUrl, varPath[i], contentType, logProgress, t)

				t.Logf("Uploading %s message media complete", varMedia[i])
			}

			defer func(mcn string) {
				varMediaCloudName := make([]string, 2)
				_, err = fmt.Sscanf(mcn, "blur_placeholder:%s actual:%s", &varMediaCloudName[0], &varMediaCloudName[1])
				require.NoError(err)

				for _, baMcn := range varMediaCloudName {
					err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(baMcn).Delete(t.Context())
					require.NoError(err)
				}
			}(mediaCloudName)

			t.Log("Upload complete")
		}

		err = user2.WSConn.WriteJSON(map[string]any{
			"action": "chat: send message",
			"data": map[string]any{
				"msg": map[string]any{
					"type": "photo",
					"props": map[string]any{
						"media_cloud_name": mediaCloudName,
						"caption":          "Check this out! Isn't this beautiful?!",
					},
				},
				"toUser": user1.Username,
				"at":     time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		// user2's server reply (response) to event sent
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: send message",
			"data": td.Map(map[string]any{
				"new_msg_id": td.Ignore(),
			}, nil),
		}, nil))

		user2NewMsgId = user2ServerReply["data"].(map[string]any)["new_msg_id"].(string)
	}

	{
		t.Log("Action: user1 receives the message | acknowledges 'delivered'")

		user1NewMsgReceived := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1NewMsgReceived, td.Map(map[string]any{
			"event": "chat: new message",
			"data": td.SuperMapOf(map[string]any{
				"id": user2NewMsgId,
				"content": td.SuperMapOf(map[string]any{
					"type": "photo",
					"props": td.SuperMapOf(map[string]any{
						"caption":   "Check this out! Isn't this beautiful?!",
						"media_url": td.Ignore(),
					}, nil),
				}, nil),
				"delivery_status": "sent",
				"sender": td.SuperMapOf(map[string]any{
					"username": user2.Username,
				}, nil),
			}, nil),
		}, nil))

		err := user1.WSConn.WriteJSON(map[string]any{
			"action": "chat: ack message delivered",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msgId":           user2NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: ack message delivered",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user2 receives the 'delivered' acknowledgement | marks message as 'delivered'")

		user2DelvAckReceipt := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2DelvAckReceipt, td.Map(map[string]any{
			"event": "chat: message delivered",
			"data": td.SuperMapOf(map[string]any{
				"partner_username": user1.Username,
				"msg_id":           user2NewMsgId,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 then acknowledges 'read'")

		err := user1.WSConn.WriteJSON(map[string]any{
			"action": "chat: ack message read",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msgId":           user2NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: ack message read",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user2 receives the 'read' acknowledgement | marks message as 'read'")

		user2ReadAckReceipt := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ReadAckReceipt, td.Map(map[string]any{
			"event": "chat: message read",
			"data": td.SuperMapOf(map[string]any{
				"partner_username": user1.Username,
				"msg_id":           user2NewMsgId,
			}, nil),
		}, nil))
	}

	{
		<-(time.NewTimer(200 * time.Millisecond).C)
		t.Log("Action: user1 opens his chat history with user2")

		err := user1.WSConn.WriteJSON(map[string]any{
			"action": "chat: get history",
			"data": map[string]any{
				"partnerUsername": user2.Username,
			},
		})
		require.NoError(err)

		// user1's server reply (response) to event sent
		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: get history",
			"data": td.All(
				td.Contains(td.SuperMapOf(map[string]any{
					"id": user1NewMsgId,
					"content": td.SuperMapOf(map[string]any{
						"type": "text",
						"props": td.SuperMapOf(map[string]any{
							"text_content": "Hi. How're you doing?",
						}, nil),
					}, nil),
					"delivery_status": "read",
					"reactions": td.All(td.Contains(td.Map(map[string]any{
						"emoji": "🚀",
						"reactor": td.Map(map[string]any{
							"username":        user2.Username,
							"profile_pic_url": td.Ignore(),
						}, nil),
					}, nil))),
					"sender": td.SuperMapOf(map[string]any{
						"username": user1.Username,
					}, nil),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"id": user2NewMsgId,
					"content": td.SuperMapOf(map[string]any{
						"type": "photo",
						"props": td.SuperMapOf(map[string]any{
							"caption":   "Check this out! Isn't this beautiful?!",
							"media_url": td.Ignore(),
						}, nil),
					}, nil),
					"delivery_status": "read",
					"sender": td.SuperMapOf(map[string]any{
						"username": user2.Username,
					}, nil),
				}, nil)),
			),
		}, nil))
	}

	{
		t.Log("Action: user2 removes reaction to user1's message")

		err := user2.WSConn.WriteJSON(map[string]any{
			"action": "chat: remove reaction to message",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		// user2's server reply (response) to event sent
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "chat: remove reaction to message",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user1 is notified of user2's reaction removal to his message")

		user1ReadAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ReadAckReceipt, td.Map(map[string]any{
			"event": "chat: message reaction removed",
			"data": td.SuperMapOf(map[string]any{
				"partner_username": user2.Username,
				"msg_id":           user1NewMsgId,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 deletes his message for everyone")
	}

	{
		t.Log("Action: user2 deletes his message for himself")
	}

	{
		t.Log("Action: user1 deletes user2's message")
	}
}
