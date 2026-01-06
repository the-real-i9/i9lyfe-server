package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func XTestUserAuthStory(t *testing.T) {
	// t.Parallel()

	user1 := UserT{
		Email:    "suberu@gmail.com",
		Username: "suberu",
		Name:     "Suberu Garuda",
		Password: "sketeppy",
		Birthday: bday("2000-11-07"),
		Bio:      "Whatever!",
	}

	{
		t.Log("Action: user1 requests a new account")

		reqBody, err := makeReqBody(map[string]any{"email": user1.Email})
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

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", user1.Email),
			}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 sends an incorrect email verf code")

		reqBody, err := makeReqBody(map[string]any{"code": "011111"})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusBadRequest, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)

		require.Equal(t, "Incorrect verification code! Check or Re-submit your email.", rb)
	}

	{
		t.Log("Action: user1 sends the correct email verification code")

		verfCode := os.Getenv("DUMMY_TOKEN")

		reqBody, err := makeReqBody(map[string]any{"code": verfCode})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
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

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": fmt.Sprintf("Your email, %s, has been verified!", user1.Email),
			}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 submits her information")

		reqBody, err := makeReqBody(map[string]any{
			"username": user1.Username,
			"name":     user1.Name,
			"password": user1.Password,
			"birthday": user1.Birthday,
			"bio":      user1.Bio,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
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
				"user": td.Ignore(),
				"msg":  "Signup success!",
			}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 signs out")

		req, err := http.NewRequest("GET", signoutPath, nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}
	}

	{
		t.Log("Action: user1 signs in with incorrect credentials")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user1.Email,
			"password":        "millinix",
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signinPath, reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusNotFound, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)
		require.Equal(t, "Incorrect email or password", rb)
	}

	{
		t.Log("Action: user1 signs in with correct credentials")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user1.Email,
			"password":        user1.Password,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signinPath, reqBody)
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

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": "Signin success!",
			}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 signs out again")

		req, err := http.NewRequest("GET", signoutPath, nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}
	}

	{
		t.Log("Action: user1 requests password reset")

		reqBody, err := makeReqBody(map[string]any{"email": user1.Email})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", forgotPasswordPath+"/request_password_reset", reqBody)
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

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": fmt.Sprintf("Enter the 6-digit number token sent to %s to reset your password", user1.Email),
			}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 sends an incorrect email confirmation token")

		reqBody, err := makeReqBody(map[string]any{"token": "011111"})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", forgotPasswordPath+"/confirm_email", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusBadRequest, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)
		require.Equal(t, "Incorrect password reset token! Check or Re-submit your email.", rb)
	}

	{
		t.Log("Action: user1 sends the correct email confirmation token")

		token := os.Getenv("DUMMY_TOKEN")

		reqBody, err := makeReqBody(map[string]any{"token": token})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", forgotPasswordPath+"/confirm_email", reqBody)
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

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": fmt.Sprintf("%s, you're about to reset your password!", user1.Email),
			}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 resets her password")

		newPassword := "millinie"
		reqBody, err := makeReqBody(map[string]any{
			"newPassword":        newPassword,
			"confirmNewPassword": newPassword,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", forgotPasswordPath+"/reset_password", reqBody)
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

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": "Your password has been changed successfully",
			}, nil))

		user1.Password = newPassword
	}

	{
		t.Log("Action: user1 signs in with new password")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user1.Username,
			"password":        user1.Password,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signinPath, reqBody)
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

		td.Cmp(td.Require(t), rb, td.SuperMapOf(
			map[string]any{
				"msg": "Signin success!",
			}, nil))
	}

	{
		t.Log("Action: userX requests a new account with already existing email")

		reqBody, err := makeReqBody(map[string]any{"email": user1.Email})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/request_new_account", reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusBadRequest, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)

		require.Equal(t, "A user with this email already exists.", rb)
	}
}
