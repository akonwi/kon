willBeOk = function(thing) {
}

createK = function(name, result) {
  result.name = name
  result.pass = true
  result.failures = {}

  var ok = function(thing, message) {
    if (thing === undefined || thing === null || thing === false) {
      result.pass = false
      result.failures[message] = { subject: thing }
    }
  }

  var willBeOk = function(thing, message) {
    result.promise = new Promise(function(resolve, reject) {
      Promise.resolve(thing).then(function(value) {
        ok(value, message)
        resolve(result)
      })
    })
  }

  return {
    pass: function() {
      result.pass = true
    },
    fail: function() {
      result.pass = false
    },
    ok: ok,
    willBeOk: willBeOk,
    notOk: function(thing, message) {
      if (thing === undefined || thing === null || thing === false)
        result.pass = true
      else {
        result.pass = false
        result.failures[message] = { subject: thing }
      }
    }
  }
}

module.exports = {
  test: function(name, cb) {
    var result = {}
    var ktest = function() {
      cb(createK(name, result))
      return result;
    }
    return ktest
  }
}
