var tap = require('tap')
var kon = require('../index')

tap.test("kon::test that passes", function(t) {
  var ktest = kon.test("this is a passing test", function(k) {
    return k.pass()
  })

  var result = ktest()

  t.is(result.name, "this is a passing test", "result.kname should be what was passed to kon.test")
  t.true(result.pass, "result.pass should be true")
  t.end()
})

tap.test("kon::test that fails", function(t) {
  var ktest = kon.test("this is a failing test", function(k) {
    return k.fail()
  })

  var result = ktest()

  t.is(result.name, "this is a failing test", "result.kname should be what was passed to kon.test")
  t.false(result.pass, "result.pass should be false")
  t.end()
})

tap.test("kon::ok checks whether something is not false|undefined|null", function(t) {
  var passingKtest = kon.test("kon::ok checks for existence", function(k) {
    k.ok(true, "kon.ok returns true for the boolean true")
    k.ok(1, "kon.ok returns true for the number 1")
    k.ok({}, "kon.ok returns true for an object")
    k.ok('', "kon.ok returns true for an empty string")
  })
  t.true(passingKtest().pass, "kon::ok works")

  var failingKtest = kon.test("kon::ok fails for non-existence and not true things ", function(k) {
    k.ok(undefined, "kon.ok returns false for undefined")
    k.ok(null, "kon.ok returns false for null")
    k.ok(false, "kon.ok returns false for the boolean false")
  })
  var failingResult = failingKtest()
  t.false(failingResult.pass, "kon::ok works against false-like things")
  t.strictSame(failingResult.failures, {
    "kon.ok returns false for undefined": { subject: undefined },
    "kon.ok returns false for null": { subject: null },
    "kon.ok returns false for the boolean false": { subject: false }
  }, "ktest contains info about failing assertions")

  t.end()
})

tap.test("kon::notOk checks whether something is actually false|undefined|null", function(t) {
  var passingKtest = kon.test("kon::notOk checks for non-existence and not true", function(k) {
    k.notOk(undefined, "kon.notOk returns true for undefined")
    k.notOk(null, "kon.notOk returns true for null")
    k.notOk(false, "kon.notOk returns true for the boolean false")
  })
  t.true(passingKtest().pass, "kon::notOk works")

  var failingKtest = kon.test("kon::notOk fails for things that exist ", function(k) {
    k.notOk(true, "kon.notOk returns false for the boolean true")
    k.notOk(1, "kon.notOk returns false for numbers")
    k.notOk({}, "kon.notOk returns false for objects")
    k.notOk('', "kon.notOk returns false for strings")
  })
  var failingResult = failingKtest()
  t.false(failingResult.pass, "kon::notOk works against existing things")
  t.strictSame(failingResult.failures, {
    "kon.notOk returns false for the boolean true": { subject: true },
    "kon.notOk returns false for numbers": { subject: 1 },
    "kon.notOk returns false for objects": { subject: {} },
    "kon.notOk returns false for strings": { subject: '' }
  }, "ktest contains info about failing assertions")
  t.end()
})

tap.test("kon::willBeOk checks if a promise/thing resolves to something that is not false|undefined|null", function(t) {
  t.plan(6)

  var ktest = kon.test("kon::willBeOk", function(k) {
    k.willBeOk(true)
  })
  ktest().promise.then(function(result) {
    t.true(result.pass, "the ktest with ::willBeOk(true) passed")
  })

  var failing = kon.test("kon::willBeOk", function(k) {
    k.willBeOk(false, "willBeOk evaluates as false for the boolean false")
  })
  failing().promise.then(function(result) {
    t.false(result.pass, "the ktest with a failure and promise failed")
    t.strictSame(result.failures, {
      "willBeOk evaluates as false for the boolean false": { subject: false }
    }, "ktest() promised result contains info about failing assertions")
  })

  var ktestWithPromise = kon.test("kon::willBeOk", function(k) {
    k.willBeOk(Promise.resolve(true))
  })
  ktestWithPromise().promise.then(function(result) {
    t.true(result.pass, "the ktest with willBeOk(Promise.resolve(true)) passed")
  })

  var failingktestWithPromise = kon.test("kon::willBeOk", function(k) {
    k.willBeOk(Promise.resolve(false), "k.willBeOk(Promise.resolve(false)) should not pass")
  })
  failingktestWithPromise().promise.then(function(result) {
    t.false(result.pass, "the ktest with willBeOk(Promise.resolve(false)) failed")
    t.strictSame(result.failures, {
      "k.willBeOk(Promise.resolve(false)) should not pass": { subject: false }
    }, "ktest() promised result contains info about failing assertions")
  })
})
