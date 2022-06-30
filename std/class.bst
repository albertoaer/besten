import "unsafe.bst" unsafe

alias Prop for Any

struct Class:
    props Map|Prop,
    methods Map|{Class,Vec|Prop}|Prop

fn class do
    return {[Map|Prop], [Map|{Class,Vec|Prop}|Prop]}Class

op []: self Class, name Str do
    val p = self.props[name]
    if not: p[1] do
        throw "Undefined property"
    return p[0]

fn setbykey: prop Prop, name Str, self Class do
    self.props[name] = prop

fn setbykey: method {Class,Vec|Prop}|Prop, name Str, self Class do
    self.methods[name] = method

fn call: self, name, 'x do
    val method = self.methods[name]
    if not: method[1] do
        throw "Undefined method"
    return callfn: method[0], self, x

fn call: self, name do
    return call: self, name, '[Vec|Prop]