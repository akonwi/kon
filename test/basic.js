var tap = require('tap')
var kon = require('../index')

tap.test("kon::test that passes", function(t) {
  t.plan(2)

  var ktest = kon.test("this is a test", function(k) {
    return k.pass()
  })

  t.is(ktest.kname, "this is a test", "ktest.kname should be what was passed to kon.test")
  t.ok(ktest(), "ktest passes")
})

tap.test("kon::test that fails", function(t) {
  t.plan(1)

  var ktest = kon.test("this is a test", function(k) {
    return k.fail()
  })

  t.notOk(ktest(), "ktest fails")
})

tap.test("kon::ok checks whether something is not false|undefined|null", function(t) {
  t.ok(kon.ok(true), "kon.ok returns true for the boolean true")
  t.ok(kon.ok(1), "kon.ok returns true for the number 1")
  t.ok(kon.ok({}), "kon.ok returns true for an object")
  t.ok(kon.ok(''), "kon.ok returns true for an empty string")
  t.notOk(kon.ok(undefined), "kon.ok returns false for undefined")
  t.notOk(kon.ok(null), "kon.ok returns false for null")
  t.notOk(kon.ok(false), "kon.ok returns false for the boolean false")
  t.end()
})

tap.test("kon::willBeOk checks if a promise resolves to something that is not false|undefined|null", function(t) {
  t.plan(14)
  kon.willBeOk(true).then(function(ok) {
    t.ok(ok, "kon.willBeOk resolves to true for the boolean true")
  })
  kon.willBeOk(Promise.resolve(true)).then(function(ok) {
    t.ok(ok, "kon.willBeOk resolves to true for a promise that resolves to true")
  })
  kon.willBeOk(1).then(function(ok) {
    t.ok(ok, "kon.willBeOk resolves to true for a number")
  })
  kon.willBeOk(Promise.resolve(1)).then(function(ok) {
    t.ok(ok, "kon.willBeOk resolves to true for a promise that resolves to a number")
  })
  kon.willBeOk({}).then(function(ok) {
    t.ok(ok, "kon.willBeOk resolves to true for an object")
  })
  kon.willBeOk(Promise.resolve({})).then(function(ok) {
    t.ok(ok, "kon.willBeOk resolves to true for a promise that resolves to an object")
  })
  kon.willBeOk('').then(function(ok) {
    t.ok(ok, "kon.willBeOk resolves to true for an empty string")
  })
  kon.willBeOk(Promise.resolve('')).then(function(ok) {
    t.ok(ok, "kon.willBeOk resolves to true for a promise that resolves to a string")
  })
  kon.willBeOk(undefined).then(function(ok) {
    t.notOk(ok, "kon.willBeOk resolves to false for undefined")
  })
  kon.willBeOk(Promise.resolve(undefined)).then(function(ok) {
    t.notOk(ok, "kon.willBeOk resolves to false for a promise that resolves to undefined")
  })
  kon.willBeOk(null).then(function(ok) {
    t.notOk(ok, "kon.willBeOk resolves to false for null")
  })
  kon.willBeOk(Promise.resolve(null)).then(function(ok) {
    t.notOk(ok, "kon.willBeOk resolves to false for a promise that resolves to null")
  })
  kon.willBeOk(false).then(function(ok) {
    t.notOk(ok, "kon.willBeOk resolves to false for the boolean false")
  })
  kon.willBeOk(Promise.resolve(false)).then(function(ok) {
    t.notOk(ok, "kon.willBeOk resolves to false for a promise that resolves to false")
  })
})

tap.test("ktest does what it is supposed to", function(t) {
  var ktest = kon.test("a passing test with promises", function(k) {
    return k.willBeOk(Promise.resolve(true));
  })

  ktest().then(function(ok) {
    t.ok(ok, "ktest resolved to a pass")
    t.end()
  })
})
