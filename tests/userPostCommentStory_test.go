package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/fasthttp/websocket"
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

				require.Contains(t, rb, "msg")
				require.Equal(t, fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", user.Email), rb["msg"])

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

				require.Contains(t, rb, "msg")
				require.Equal(t, fmt.Sprintf("Your email, %s, has been verified!", user.Email), rb["msg"])

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

				require.Contains(t, rb, "msg")
				require.Contains(t, rb, "user")
				require.Equal(t, "Signup success!", rb["msg"])

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

			go func(user *UserT) {
				var wsMsg map[string]any

				for {
					if err := wsConn.ReadJSON(&wsMsg); err != nil {
						break
					}
				}

				user.ServerWSMsg = wsMsg

			}(user)

			user.WSConn = wsConn
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

		require.Contains(t, rb, "id")

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

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")

		// user1 is notified
		require.NotEmpty(t, user1.ServerWSMsg)
		require.Equal(t, "new notification", user1.ServerWSMsg["event"])

		recvNotif := user1.ServerWSMsg["data"]

		require.Contains(t, recvNotif, "id")
		require.Contains(t, recvNotif, "reaction_to_post")

		user1.ServerWSMsg = nil
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

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")

		// user1 is notified
		require.NotEmpty(t, user1.ServerWSMsg)
		require.Equal(t, "new notification", user1.ServerWSMsg["event"])

		recvNotif := user1.ServerWSMsg["data"]

		require.Contains(t, recvNotif, "id")
		require.Contains(t, recvNotif, "reaction_to_post")

		user1.ServerWSMsg = nil
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

		for _, reactor := range reactors {
			require.Contains(t, reactor, "username")

			require.Contains(t, []string{user2.Username, user3.Username}, reactor["username"])

			if reactor["username"].(string) == user2.Username {
				require.Equal(t, "🤔", reactor["reaction"])
			}

			if reactor["username"].(string) == user3.Username {
				require.Equal(t, "😀", reactor["reaction"])
			}
		}
	}
}
