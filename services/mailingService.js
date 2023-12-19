import nodemailer from 'nodemailer'

const transporter = nodemailer.createTransport({
  host: "smtp.gmail.com",
  port: 465,
  secure: true,
  auth: {
    // eslint-disable-next-line no-undef
    user: process.env.MAILING_EMAIL,
    // eslint-disable-next-line no-undef
    pass: process.env.MAILING_PASSWORD,
  },
});

const sendMail = ({to, subject, html}) => {
  // use nodemailer to send message to email
  transporter.sendMail({
    from: 'app@i9lyfe.com',
    to,
    subject,
    html,
  })
}

export default sendMail