export function searchAndFilter(req, res, next) {
  const { filter } = req.query

  const validFilters = ["user", "photo", "video", "reel", "story", "hashtag"]
  if (filter && !validFilters.includes(filter)) {
    res.status(422).send({
      error: {
        queryParam: "filter",
        msg: "invalid filter value",
        validFilters,
      },
    })
  }

  return next()
}
