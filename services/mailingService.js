import nodemailer from "nodemailer"

/** @interface */
export class PrimaryMailSender {
  /** @param {string} email */
  // eslint-disable-next-line no-unused-vars
  send(email) {
    throw new Error("send method must be implemented")
  }
}

export class EmailVerificationSuccessMailSender extends PrimaryMailSender {
  send(email) {
    sendMail({
      to: email,
      subject: "i9lyfe - Email verification success",
      html: `<p>Your email <strong>${email}</strong> has been verified!</p>`,
    })
  }
}

export class PwdResetSuccessMailSender extends PrimaryMailSender {
  send(email) {
    sendMail({
      to: email,
      subject: "i9lyfe - Password reset successful",
      html: `<p>${email}, your password has been changed successfully!</p>`,
    })
  }
}

/** @interface */
export class TokenMailSender {
  /**
   * @param {string} email
   * @param {number} token
   */
  // eslint-disable-next-line no-unused-vars
  sendToken(email, token) {
    throw new Error("send method must be implemented")
  }
}

export class EmailVerificationTokenMailSender extends TokenMailSender {
  sendToken(email, token) {
    sendMail({
      to: email,
      subject: "i9lyfe - Verify your email",
      html: `<p>Your email verification token is <strong>${token}</strong></p>`,
    })
  }
}

export class PwdResetTokenMailSender extends TokenMailSender {
  sendToken(email, token) {
    sendMail({
      to: email,
      subject: "i9lyfe - Confirm your email: Password Reset",
      html: `<p>Your password reset token is <strong>${token}</strong></p>`,
    })
  }
}

/**
 * @param {Object} mailInfo
 * @param {string} mailInfo.to
 * @param {string} mailInfo.subject
 * @param {string} mailInfo.html
 */
export const sendMail = ({ to, subject, html }) => {
  try {
    const transporter = nodemailer.createTransport({
      host: "smtp.gmail.com",
      port: 465,
      secure: true,
      auth: {
        user: process.env.MAILING_EMAIL,
        pass: process.env.MAILING_PASSWORD,
      },
    })

    transporter.sendMail({
      from: "<no-reply@accounts.i9lyfe.com>",
      to,
      subject,
      html,
    })
  } catch (error) {
    console.error("sendMail Error", error)
  }
}

export default sendMail
