import app from "./app.js"

const quickTest = async () => {
  
}

app.listen(5000, () => {
  console.log(`Server running at http://localhost:${process.env.PORT || 5000}`)
  quickTest()
})
