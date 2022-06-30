<img width='128' height='128' src='https://user-images.githubusercontent.com/24974091/176786448-42a89ae4-5e19-4a8d-a7bc-cb8a6bce2b70.png'></img>

Besten is a programming language made in my spare time, it may have no more commits as I'm working on other repositories

I encourage you to check the source code out as the language contains interesting features like:
- Complete working type system
- A virtual machine full featured reusable for other projects
- Concurrent module loader
- Templated functions and data structures for its manipulation
- Lexer with Python like block indentations
- Extensible abstract syntax tree for expressions
- And so on...

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