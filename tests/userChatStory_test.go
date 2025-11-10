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

func XTestUserChatStory(t *testing.T) {
	t.Parallel()

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
				require.NoError(t, err)

				res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
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

		for _, user := range []*UserT{&user1, &user2} {
			user := user

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

	user1NewMsgId := ""

	{
		t.Log("Action: user1 sends message to user2")

		err := user1.WSConn.WriteJSON(map[string]any{
			"event": "chat: send message: text",
			"data": map[string]any{
				"props": map[string]any{
					"content": "Hi. How're you doing?",
				},
				"toUser": user2.Username,
				"at":     time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		// user1's server reply (response) to event sent
		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: send message: text",
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
				"id":              user1NewMsgId,
				"type":            "text",
				"props":           td.Contains(`"content":"Hi. How're you doing?"`),
				"delivery_status": "sent",
				"sender": td.SuperMapOf(map[string]any{
					"username": user1.Username,
				}, nil),
			}, nil),
		}, nil))

		err := user2.WSConn.WriteJSON(map[string]any{
			"event": "chat: ack message delivered",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: ack message delivered",
			"data":    true,
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
			"event": "chat: ack message read",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: ack message read",
			"data":    true,
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
			"event": "chat: react to message",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"reaction":        "🚀",
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		// user2's server reply (response) to event sent
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: react to message",
			"data":    true,
		}, nil))
	}

	{
		t.Log("Action: user1 receives user2's reaction to his message | attaches it to chat snippet")

		user1ReadAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ReadAckReceipt, td.Map(map[string]any{
			"event": "chat: message reaction",
			"data": td.SuperMapOf(map[string]any{
				"partner_username": user2.Username,
				"msg_id":           user1NewMsgId,
				"reaction":         "🚀",
			}, nil),
		}, nil))
	}

	user2NewMsgId := ""

	{
		t.Log("Action: user2 sends message to user1")

		photo, err := os.ReadFile("./test_files/photo_1.png")
		require.NoError(t, err)

		err = user2.WSConn.WriteJSON(map[string]any{
			"event": "chat: send message: photo",
			"data": map[string]any{
				"props": map[string]any{
					"data":    photo,
					"size":    len(photo),
					"caption": "Check this out! Isn't this beautiful?!",
				},
				"toUser": user1.Username,
				"at":     time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		// user2's server reply (response) to event sent
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: send message: photo",
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
				"id":              user2NewMsgId,
				"type":            "photo",
				"props":           td.All(td.Contains(`"caption":"Check this out! Isn't this beautiful?!"`), td.Contains(`"data_url":`)),
				"delivery_status": "sent",
				"sender": td.SuperMapOf(map[string]any{
					"username": user2.Username,
				}, nil),
			}, nil),
		}, nil))

		err := user1.WSConn.WriteJSON(map[string]any{
			"event": "chat: ack message delivered",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msgId":           user2NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: ack message delivered",
			"data":    true,
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
			"event": "chat: ack message read",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msgId":           user2NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: ack message read",
			"data":    true,
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
		t.Log("Action: user1 opens his chat history with user2")

		err := user1.WSConn.WriteJSON(map[string]any{
			"event": "chat: get history",
			"data": map[string]any{
				"partnerUsername": user2.Username,
			},
		})
		require.NoError(t, err)

		// user1's server reply (response) to event sent
		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: get history",
			"data": td.All(
				td.Contains(td.SuperMapOf(map[string]any{
					"id":              user1NewMsgId,
					"type":            "text",
					"props":           td.Contains(`"content":"Hi. How're you doing?"`),
					"delivery_status": "read",
					"reactions": td.All(td.Contains(td.Map(map[string]any{
						"reaction": "🚀",
						"user": td.Map(map[string]any{
							"username":        user2.Username,
							"profile_pic_url": td.Ignore(),
						}, nil),
					}, nil))),
					"sender": td.SuperMapOf(map[string]any{
						"username": user1.Username,
						"presence": td.Any("online", "offline"),
					}, nil),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"id":              user2NewMsgId,
					"type":            "photo",
					"props":           td.All(td.Contains(`"caption":"Check this out! Isn't this beautiful?!"`), td.Contains(`"data_url":`)),
					"delivery_status": "read",
					"sender": td.SuperMapOf(map[string]any{
						"username": user2.Username,
						"presence": td.Any("online", "offline"),
					}, nil),
				}, nil)),
			),
		}, nil))
	}

	{
		t.Log("Action: user2 removes reaction to user1's message")

		err := user2.WSConn.WriteJSON(map[string]any{
			"event": "chat: remove reaction to message",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		// user2's server reply (response) to event sent
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":   "server reply",
			"toEvent": "chat: remove reaction to message",
			"data":    true,
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
