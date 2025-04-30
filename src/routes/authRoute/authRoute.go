package authRoute

import (
	"i9lyfe/src/controllers/auth/passwordResetControllers"
	"i9lyfe/src/controllers/auth/signinControllers"
	"i9lyfe/src/controllers/auth/signupControllers"
	"i9lyfe/src/middlewares/authMiddlewares"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router,
) {
	router.Post(
		"/signup/request_new_account",
		signupControllers.RequestNewAccount,
	)

	router.Post(
		"/signup/verify_email",
		authMiddlewares.SignupSession,
		signupControllers.VerifyEmail,
	)

	router.Post(
		"/signup/register_user",
		authMiddlewares.SignupSession,
		signupControllers.RegisterUser,
	)

	router.Post(
		"/signin",
		signinControllers.Signin,
	)

	router.Post(
		"/forgot_password/request_password_reset",
		passwordResetControllers.RequestPasswordReset,
	)

	router.Post(
		"/forgot_password/confirm_action",
		authMiddlewares.PasswordResetSession,
		passwordResetControllers.ConfirmAction,
	)

	router.Post(
		"/forgot_password/reset_password",
		authMiddlewares.PasswordResetSession,
		passwordResetControllers.ResetPassword,
	)
}
