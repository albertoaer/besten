fn abs: x do
    if x < 0 do
        return -x
    return x

fn pow: x, y do
    var d = 1
    while y > 0 do
        d = d * x
        y = y-1
    return d

alias fn pow:2 op **