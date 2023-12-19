import nodemailer from 'nodemailer'

const transporter = nodemailer.createTransport({
  host: "smtp.gmail.com",
  port: 465,
  secure: true,
  auth: {
    user: process.env.MAILING_EMAIL,
    pass: process.env.MAILING_PASSWORD,
  },
});

const sendMail = ({to, subject, html}) => {
  // use nodemailer to send message to email
  transporter.sendMail({
    from: '<no-reply@accounts.i9lyfe.com>',
    to,
    subject,
    html,
  })
}

export default sendMail