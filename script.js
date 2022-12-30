



function request(r) {
  log(r.Url)
  if (r.Url == "/nop") {
    log("nop function")
    r.Status = 202
    return
  }
  log(r.Data)
  r.Status = 200
  r.DataOut = "ole!\n"
  h2r("http://localhost:8080/nop", function(resp) {
    log(resp.Status)
  })
}
