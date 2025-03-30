package authRoutes

import (
	"i9lyfe/src/controllers/auth/passwordResetControllers"
	"i9lyfe/src/controllers/auth/siginControllers"
	"i9lyfe/src/controllers/auth/signupControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Post("/signup/request_new_account", signupControllers.RequestNewAccount)
	router.Post("/signup/verify_email", signupControllers.VerifyEmail)

	router.Post("/signup/register_user", signupControllers.RegisterUser)

	router.Post("/signin", signinControllers.Signin)

	router.Post("/forgot_password/request_password_reset", passwordResetControllers.RequestPasswordReset)

	router.Post("/forgot_password/confirm_action", passwordResetControllers.ConfirmAction)

	router.Post("/forgot_password/reset_password", passwordResetControllers.ResetPassword)
}
