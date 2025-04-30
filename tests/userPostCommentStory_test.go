package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const appPathPriv = HOST_URL + "/api/app/private"
const wsPath = WSHOST_URL + "/api/app/private/ws"

func TestPostCommentStory(t *testing.T) {
	t.Parallel()

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

				td.Cmp(td.Require(t), rb, td.SuperMapOf(
					map[string]any{
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

				td.Cmp(td.Require(t), rb, td.SuperMapOf(
					map[string]any{
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

				td.Cmp(td.Require(t), rb, td.SuperMapOf(
					map[string]any{
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

			commChan := make(chan map[string]any)

			go func(commChan chan<- map[string]any) {

				for {
					var wsMsg map[string]any

					if err := wsConn.ReadJSON(&wsMsg); err != nil {
						break
					}

					commChan <- wsMsg
				}

				close(commChan)

			}(commChan)

			user.WSConn = wsConn
			user.ServerWSMsg = commChan
		}
	}

	t.Log("-----")

	user1Post1Id := ""

	{
		t.Log("Action: user1 creates post1")

		photo1, err := os.ReadFile("./test_files/photo_1.png")
		require.NoError(t, err)

		reqBody, err := makeReqBody(map[string]any{
			"media_data_list": [][]byte{photo1},
			"type":            "photo",
			"description":     "This is No.1 #trending",
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

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"id": td.Ignore(),
			}, nil))

		user1Post1Id = rb["id"].(string)
	}

	{
		t.Log("user2 reacts to user1's post1 | user1 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"reaction": "🤔",
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1Post1Id+"/react", reqBody)
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
		serverWSMsg := <-user1.ServerWSMsg

		require.NotEmpty(t, serverWSMsg)

		td.Cmp(td.Require(t), serverWSMsg, td.SuperMapOf(
			map[string]any{
				"event": "new notification",
				"data": td.SuperMapOf(
					map[string]any{
						"id":           td.Ignore(),
						"type":         "reaction_to_post",
						"reactor_user": td.SuperSliceOf([]any{"username", user2.Username}, nil),
					},
					nil),
			}, nil))
	}

	{
		t.Log("user3 reacts to user1's post1 | user1 is notified")

		reqBody, err := makeReqBody(map[string]any{
			"reaction": "😀",
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", appPathPriv+"/posts/"+user1Post1Id+"/react", reqBody)
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

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)

		// user1 is notified
		serverWSMsg := <-user1.ServerWSMsg

		td.Cmp(td.Require(t), serverWSMsg, td.SuperMapOf(
			map[string]any{
				"event": "new notification",
				"data": td.SuperMapOf(
					map[string]any{
						"id":           td.Ignore(),
						"type":         "reaction_to_post",
						"reactor_user": td.SuperSliceOf([]any{"username", user3.Username}, nil),
					}, nil),
			}, nil))
	}

	{
		t.Log("user1 checks reactors to her post1")

		req, err := http.NewRequest("GET", appPathPriv+"/posts/"+user1Post1Id+"/reactors", nil)
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

		reactors, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		require.Len(t, reactors, 2)

		td.Cmp(td.Require(t), reactors, td.Contains(td.Any(
			td.SuperMapOf(map[string]any{
				"username": user2.Username,
				"reaction": "🤔",
			}, nil),
			td.SuperMapOf(map[string]any{
				"username": user3.Username,
				"reaction": "😀",
			}, nil),
		)))
	}
}
