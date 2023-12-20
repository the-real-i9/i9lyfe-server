import nodemailer from "nodemailer"

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
