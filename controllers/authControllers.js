export const newAccountRequestController = async (req, res) => {
  // parse the request body
  // contact registrationRequestService which will respond with {ok, err: { code, reason} | null} response
  // if ok === false, then check reason, else
  // res.status(200).json({ message: "Check the verification code sent to email" })

  try {
    const { email } = req.body
  } catch (error) {}
}

export const emailVerificationController = async (req, res) => {
  try {
  } catch (error) {}
}

export const signupController = async (req, res) => {
  try {
  } catch (error) {}
}

export const signinController = async (req, res) => {
  try {
  } catch (error) {}
}
