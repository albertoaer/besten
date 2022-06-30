#!./besten
import "./std/iterators.bst"

fn main: args Vec|Str do
    for line in indexer: vec(
        "Hello, I'm besten",
        "I encourage you to check out my code",
        "And even try to write some lines ;)"
        ) do
            print: line.vec[line.value]