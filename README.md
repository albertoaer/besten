# Besten

## What holds this repository?
1. Besten Lexer
2. Besten Parser
3. Besten Interpreter
4. Besten Module Loader

### Besten Lexer
Located in [./internal/lexer](./internal/lexer)

A set of algorithm that makes posible to convert source files into a hierarchy of tokens

### Besten Parser
Located in [./internal/parser](./internal/parser)

Intermediary that converts organized tokens into Besten instructions that will be run in the interpreter

### Besten Interpreter
Located in [./internal/runtime](./internal/runtime)

A set of instructions that are grouped into programs
Each instruction is not bytecode, it's more like a high level set of simple instruction that modifies stack, context, invokes functions and manipulates data

### Besten Module Loader
Located in [./internal/modules](./internal/modules)

Auxiliary module that abstracts the module loading process