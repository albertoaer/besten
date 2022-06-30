struct Range:
    value Int,
    end Int

fn next: r Range do
    r.value = r.value + 1
    return r

fn end: r Range do
    return r.value >= r.end

struct StepRange:
    in Range,
    step Int

fn next: r StepRange do
    r.value = r.value + r.step
    return r

fn end: r StepRange do
    return r.value >= r.end

struct StepGenerator:
    value Int,
    step Int

fn next: g StepGenerator do
    g.value = g.value + g.step
    return g

struct VecIndexer:
    value Int,
    vec Vec|Any

fn indexer: vec Vec|Any do
    return {0, vec}VecIndexer

fn next: vi VecIndexer do
    vi.value = vi.value + 1
    return vi

fn end: vi VecIndexer do
    return vi.value >= len(vi.vec)