/**
 * @param {import('express').Request} req
 * @param {import('express').Response} res
 */
export const passwordResetController = async (req, res) => {
  const { stage } = req.query

  const stageHandlers = {
    password_reset_request: (req, res) => passwordResetRequestHandler(req, res),
    email_confirmation: (req, res) => passwordResetEmailConfirmationHandler(req, res),
    password_reset: (req, res) => passwordResetHandler(req, res),
  }
  stageHandlers[stage](req, res)
}

const passwordResetRequestHandler = async () => {}

const passwordResetEmailConfirmationHandler = async () => {}

const passwordResetHandler = async () => {}
