package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const signupPath string = HOST_URL + "/api/auth/signup"
const signinPath string = HOST_URL + "/api/auth/signin"
const forgotPasswordPath string = HOST_URL + "/api/auth/forgot_password"
const signoutPath string = HOST_URL + "/api/app/private/signout"

func TestUserAuthStory(t *testing.T) {

	user1 := UserT{
		Email:         "suberu@gmail.com",
		Username:      "suberu",
		Name:          "Suberu Garuda",
		Password:      "sketeppy",
		Birthday:      973555200000,
		Bio:           "Whatever!",
		SessionCookie: "",
	}

	{
		t.Log("user1 requests a new account")

		reqBody, err := makeReqBody(map[string]any{"email": user1.Email})
		require.NoError(t, err)

		res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		rb, err := resBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")
		require.Equal(t, fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", user1.Email), rb["msg"])

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("user1 sends an incorrect email verf code")

		reqBody, err := makeReqBody(map[string]any{"code": "011111"})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		eb, err := errBody(res.Body)
		require.Equal(t, http.StatusBadRequest, res.StatusCode, eb)
		require.NoError(t, err)
		require.Equal(t, "Incorrect verification code! Check or Re-submit your email.", eb)
	}

	{
		t.Log("user1 sends the correct email verification code")

		verfCode := os.Getenv("DUMMY_TOKEN")

		reqBody, err := makeReqBody(map[string]any{"code": verfCode})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		rb, err := resBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")
		require.Equal(t, fmt.Sprintf("Your email, %s, has been verified!", user1.Email), rb["msg"])

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("user1 submits her information")

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
		require.Equal(t, http.StatusCreated, res.StatusCode)

		rb, err := resBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")
		require.Contains(t, rb, "user")
		require.Equal(t, "Signup success!", rb["msg"])

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("user1 signs out")

		req, err := http.NewRequest("GET", signoutPath, nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
	}

	{
		t.Log("user1 signs in with incorrect credentials")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user1.Email,
			"password":        "millinix",
		})
		require.NoError(t, err)

		res, err := http.Post(signinPath, "application/json", reqBody)
		require.NoError(t, err)

		eb, err := errBody(res.Body)
		require.Equal(t, http.StatusNotFound, res.StatusCode, eb)
		require.NoError(t, err)
		require.Equal(t, "Incorrect email or password", eb)
	}

	{
		t.Log("user1 signs in with correct credentials")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user1.Email,
			"password":        user1.Password,
		})
		require.NoError(t, err)

		res, err := http.Post(signinPath, "application/json", reqBody)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		rb, err := resBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")
		require.Contains(t, rb["msg"], "Signin success!")

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("user1 signs out again")

		req, err := http.NewRequest("GET", signoutPath, nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
	}

	{
		t.Log("user1 requests password reset")

		reqBody, err := makeReqBody(map[string]any{"email": user1.Email})
		require.NoError(t, err)

		res, err := http.Post(forgotPasswordPath+"/request_password_reset", "application/json", reqBody)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		rb, err := resBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")
		require.Equal(t, fmt.Sprintf("Enter the 6-digit number token sent to %s to reset your password", user1.Email), rb["msg"])

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("user1 sends an incorrect email confirmation token")

		reqBody, err := makeReqBody(map[string]any{"token": "011111"})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", forgotPasswordPath+"/confirm_action", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		eb, err := errBody(res.Body)
		require.Equal(t, http.StatusBadRequest, res.StatusCode, eb)
		require.NoError(t, err)
		require.Equal(t, "Incorrect password reset token! Check or Re-submit your email.", eb)
	}

	{
		t.Log("user1 sends the correct email confirmation token")

		token := os.Getenv("DUMMY_TOKEN")

		reqBody, err := makeReqBody(map[string]any{"token": token})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", forgotPasswordPath+"/confirm_action", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		rb, err := resBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")
		require.Equal(t, fmt.Sprintf("%s, you're about to reset your password!", user1.Email), rb["msg"])

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("user1 resets her password")

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

		require.Equal(t, http.StatusOK, res.StatusCode)
		rb, err := resBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")
		require.Equal(t, "Your password has been changed successfully", rb["msg"])

		user1.Password = newPassword
	}

	{
		t.Log("user1 signs in with new password")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user1.Username,
			"password":        user1.Password,
		})
		require.NoError(t, err)

		res, err := http.Post(signinPath, "application/json", reqBody)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		rb, err := resBody[map[string]any](res.Body)
		require.NoError(t, err)
		require.Contains(t, rb, "msg")
		require.Equal(t, "Signin success!", rb["msg"])
	}

	{
		t.Log("userX requests a new account with already existing email")

		reqBody, err := makeReqBody(map[string]any{"email": user1.Email})
		require.NoError(t, err)

		res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
		require.NoError(t, err)

		eb, err := errBody(res.Body)
		require.Equal(t, http.StatusBadRequest, res.StatusCode, eb)
		require.NoError(t, err)
		require.Equal(t, "A user with this email already exists.", eb)
	}
}
