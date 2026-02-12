package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignup(t *testing.T) {

	t.Run("Request new account: with already existing email", func(t *testing.T) {
		err := requestNewAccountSetup(t.Context(), user1)
		require.NoError(t, err)

		email := user1.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", signupPath+"/request_new_account", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
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

		err = requestNewAccountCleanUp(t.Context(), user1.Username)
		require.NoError(t, err)
	})

	t.Run("Request new account", func(t *testing.T) {
		email := user1.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", signupPath+"/request_new_account", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
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
				"msg": fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", email),
			}, nil))
	})

	t.Run("Verify email with incorrect code", func(t *testing.T) {
		email := user1.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", testSessionPath+"/signup/request_new_account", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
		require.NoError(t, err)

		/* ---------- */

		reqBody, err = makeReqBody(map[string]any{"code": "011111"})
		require.NoError(t, err)

		req = httptest.NewRequest("POST", signupPath+"/verify_email", reqBody)
		req.Header.Set("Cookie", res.Header.Get("Set-Cookie"))
		req.Header.Add("Content-Type", "application/json")

		res, err = app.Test(req)
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
	})

	t.Run("Verify email with correct code", func(t *testing.T) {
		email := user1.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", testSessionPath+"/signup/request_new_account", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
		require.NoError(t, err)

		/* -------- */

		reqBody, err = makeReqBody(map[string]any{"code": os.Getenv("DUMMY_TOKEN")})
		require.NoError(t, err)

		req = httptest.NewRequest("POST", signupPath+"/verify_email", reqBody)
		req.Header.Set("Cookie", res.Header.Get("Set-Cookie"))
		req.Header.Add("Content-Type", "application/json")

		res, err = app.Test(req)
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
				"msg": fmt.Sprintf("Your email, %s, has been verified!", email),
			}, nil))
	})

	t.Run("Register user: submit info", func(t *testing.T) {
		email := user1.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", testSessionPath+"/signup/verify_email", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
		require.NoError(t, err)

		/* ------------ */
		reqBody, err = makeReqBody(map[string]any{
			"username": user1.Username,
			"name":     user1.Name,
			"password": user1.Password,
			"birthday": user1.Birthday,
			"bio":      user1.Bio,
		})
		require.NoError(t, err)

		req = httptest.NewRequest("POST", signupPath+"/register_user", reqBody)
		req.Header.Set("Cookie", res.Header.Get("Set-Cookie"))
		req.Header.Add("Content-Type", "application/json")

		res, err = app.Test(req)
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

		<-time.NewTimer(500 * time.Millisecond).C /* wait for redis to queue and bg worker to add to cache */

		userExists, err := rdb().HExists(t.Context(), "users", user1.Username).Result()
		require.NoError(t, err)
		require.True(t, userExists)

		err = registerUserCleanUp(t.Context(), user1.Username)
		require.NoError(t, err)
	})
}

func TestSignin(t *testing.T) {
	err := signinUserPrep(t.Context(), user2)
	require.NoError(t, err)

	t.Run("Signin user: incorrect credentials", func(t *testing.T) {
		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user2.Email,
			"password":        "millinix",
		})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", signinPath, reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
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
	})

	t.Run("Signin user: correct credentials", func(t *testing.T) {
		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user2.Email,
			"password":        user2.Password,
		})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", signinPath, reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
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
	})
	err = signinUserCleanUp(t.Context(), user2.Username)
	require.NoError(t, err)
}

func TestSignout(t *testing.T) {
	username := user3.Username

	reqBody, err := makeReqBody(map[string]any{"username": username})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", testSessionPath+"/auth_user", reqBody)
	req.Header.Add("Content-Type", "application/json")

	res, err := app.Test(req)
	require.NoError(t, err)

	/* ------------ */

	req = httptest.NewRequest("GET", signoutPath, nil)
	req.Header.Set("Cookie", res.Header.Get("Set-Cookie"))
	req.Header.Add("Content-Type", "application/json")

	res, err = app.Test(req)
	require.NoError(t, err)

	if !assert.Equal(t, http.StatusOK, res.StatusCode) {
		rb, err := errResBody(res.Body)
		require.NoError(t, err)
		t.Log("unexpected error:", rb)
		return
	}
}

func TestForgotPassword(t *testing.T) {

	err := forgotPasswordPrep(t.Context(), user4)
	require.NoError(t, err)

	t.Run("Request password reset", func(t *testing.T) {
		email := user4.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", forgotPasswordPath+"/request_password_reset", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
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
				"msg": fmt.Sprintf("Enter the 6-digit number token sent to %s to reset your password", user4.Email),
			}, nil))
	})

	t.Run("Confirm email with incorrect token", func(t *testing.T) {
		email := user4.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", testSessionPath+"/forgot_password/request_password_reset", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
		require.NoError(t, err)

		/* ---------- */

		reqBody, err = makeReqBody(map[string]any{"token": "011111"})
		require.NoError(t, err)

		req = httptest.NewRequest("POST", forgotPasswordPath+"/confirm_email", reqBody)
		req.Header.Set("Cookie", res.Header.Get("Set-Cookie"))
		req.Header.Add("Content-Type", "application/json")

		res, err = app.Test(req)
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
	})

	t.Run("Confirm email with correct token", func(t *testing.T) {
		email := user4.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", testSessionPath+"/forgot_password/request_password_reset", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
		require.NoError(t, err)

		/* -------- */

		reqBody, err = makeReqBody(map[string]any{"token": os.Getenv("DUMMY_TOKEN")})
		require.NoError(t, err)

		req = httptest.NewRequest("POST", forgotPasswordPath+"/confirm_email", reqBody)
		req.Header.Set("Cookie", res.Header.Get("Set-Cookie"))
		req.Header.Add("Content-Type", "application/json")

		res, err = app.Test(req)
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
				"msg": fmt.Sprintf("%s, you're about to reset your password!", user4.Email),
			}, nil))
	})

	t.Run("Reset password: submit new password", func(t *testing.T) {
		email := user4.Email

		reqBody, err := makeReqBody(map[string]any{"email": email})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", testSessionPath+"/forgot_password/confirm_email", reqBody)
		req.Header.Add("Content-Type", "application/json")

		res, err := app.Test(req)
		require.NoError(t, err)

		/* ------------ */
		user4.Password = "millinie"

		reqBody, err = makeReqBody(map[string]any{
			"newPassword":        user4.Password,
			"confirmNewPassword": user4.Password,
		})
		require.NoError(t, err)

		req = httptest.NewRequest("POST", forgotPasswordPath+"/reset_password", reqBody)
		req.Header.Set("Cookie", res.Header.Get("Set-Cookie"))
		req.Header.Add("Content-Type", "application/json")

		res, err = app.Test(req)
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
	})

	err = forgotPasswordCleanUp(t.Context(), user4.Username)
	require.NoError(t, err)
}
