package authRoutes

import (
	"i9lyfe/src/domain/auth/authMiddlewares"
	"i9lyfe/src/domain/auth/login/loginControllers"
	"i9lyfe/src/domain/auth/passwordReset/passwordResetControllers"
	"i9lyfe/src/domain/auth/signup/signupControllers"

	"github.com/gofiber/fiber/v3"
)

func Routes(router fiber.Router) {
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
		"/login",
		loginControllers.Login,
	)

	router.Post(
		"/forgot_password/request_password_reset",
		passwordResetControllers.RequestPasswordReset,
	)

	router.Post(
		"/forgot_password/confirm_email",
		authMiddlewares.PasswordResetSession,
		passwordResetControllers.ConfirmEmail,
	)

	router.Post(
		"/forgot_password/reset_password",
		authMiddlewares.PasswordResetSession,
		passwordResetControllers.ResetPassword,
	)
}
