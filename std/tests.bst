import "utils.bst"

fn assert: a, b do
    if not: a == b do
        throw "Assertion error"

fn test: header, 'funcs do
    fn try_run: func do
        rescue e do
            return false
        func:
        return true
    op ---: a, v do
        print: "Test: ", header
        return v
    op ---: a {}|Void, v Vec|{}|Void do
        print: "Test: "
        return vec(a) ++ v
    funcs = header --- funcs
    val passed = {0, len(funcs)}Range
    for x in indexer: funcs do
        if try_run: funcs[x.value] do
            next: passed
        else do
            print: "\t", funcs[x.value], " failed"
    print: "\t", passed.value, " of ", passed.end, " test passed"