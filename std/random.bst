import "iterators.bst"
import "math.bst"

struct Random:
    seed Int,
    last Int,
    idx Int

fn new_random: seed do
    if seed < pow(10, 8) do
        throw "Invalid seed"
    return {abs(seed), 0, 1}Random

fn next: random Random do
    val prefix = (random.seed%(10**random.idx))
    random.last = abs(prefix) ^ random.last
    random.idx = (random.idx + 1)%8
    return random

fn value: random Random, max Int, offset Int do
    val sum = random.last + random.seed%(10**4)
    return (sum%max)+offset