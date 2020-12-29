# Documentation

## Language features

- `map` is parallel by default, order of execution on the elements of the list not specified
- Want to have some non-exception based error system. Maybe an `Error` type, or a `maybe` type which wraps an error? You'd then get a runtime error if the `maybe` type is passed to a function which doesn't expect it.

## Glossary

Argument vs Parameter

- Argument is the data passed to a function at runtime
- Parameter is the variable defined by a function that receives the value when the function is called
