import "iterators.bst"

fn create_by_range: t, r do
    val res = [Vec|ref t]
    for i in r do
        [ref t] -> res
    return res

fn clone: a do
    val b = [ref a]
    for x in a do
        x -> b
    return b

op ++: a, b do
    val result = [ref a]
    for i in indexer(a) do
        a[i.value] -> result
    for i in indexer(b) do
        b[i.value] -> result
    return result