import "strings.bst"

op *: a Atom do
    return raw(a)

op ==: a Atom, b Atom do
    return *a == *b