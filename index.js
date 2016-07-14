ok = function(thing) {
  if (thing === undefined || thing === null || thing === false) return false
  else return true
}

willBeOk = function(thing) {
  if (!ok(thing) ||  !thing.then) return Promise.resolve(ok(thing))
  else return thing.then(willBeOk)
}

k = {
  pass: function() { return true },
  fail: function() { return false },
  ok: ok,
  willBeOk: willBeOk,
  test: function(name, cb) {
    var result = cb(k)
    var ktest = function() {
      return result;
    }
    ktest.kname = name
    return ktest
  }
}

module.exports = k
