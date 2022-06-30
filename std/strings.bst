import "utils.bst"

op ++: a Str, b Str do
    val result = *a ++ *b
    return str(result)

op ==: a Str, b Str do
    val av = *a
    val bv = *b
    if len(av) != len(bv) do
        return false
    for i in indexer(av) do
        if av[i.value] != bv[i.value] do
            return false
    return true

op []: target Str, range {Int, Int} do
    if (range[0] < 0) || range[1] > len(target) do
        throw "Out of bounds"
    val v = *target
    val nv = [ref v]
    for x in {range[0], range[1]}Range do
        v[x.value] -> nv
    return str(nv)

op []: target Str, idx Int do
    return str(vec((*target)[idx]))