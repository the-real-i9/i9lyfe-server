export const commaSeparateString = (str) => str.replaceAll(" ", ", ")

export const generateCodeWithExpiration = () => {
  const code = Math.trunc(Math.random() * 900000 + 100000)
  const expirationTime = new Date(Date.now() + (1 * 60 * 60 * 1000))

  return [code, expirationTime]
}