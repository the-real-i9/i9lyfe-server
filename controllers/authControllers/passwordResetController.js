/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const passwordResetController = async (req, res) => {
  const { stage } = req.query

  const stageHandlers = {
    email_submission: (req, res) => passwordResetRequestHandler(req, res),
    token_validation: (req, res) => emailVerificationHandler(req, res),
    user_registration: (req, res) => userRegistrationHandler(req, res),
  }
  stageHandlers[stage](req, res)
}

const passwordResetRequest = async () => {}

const passwordResetEmailConfirmation = async () => {}

const resetPassword = async () => {}
